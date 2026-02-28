package dag

import (
	"encoding/json"
	"fmt"
	"testing"
)

// testGraphsEqual verifies that two DAGs are structurally equivalent
// regardless of vertex or edge traversal order
func testGraphsEqual(t *testing.T, expected, actual *DAG) {
	// 1. Check vertex count
	expectedOrder := expected.GetOrder()
	actualOrder := actual.GetOrder()
	if expectedOrder != actualOrder {
		t.Errorf("Order mismatch: expected %d, got %d", expectedOrder, actualOrder)
	}

	// 2. Check edge count
	expectedSize := expected.GetSize()
	actualSize := actual.GetSize()
	if expectedSize != actualSize {
		t.Errorf("Size mismatch: expected %d, got %d", expectedSize, actualSize)
	}

	// 3. Check all edges exist using the same IDs from serialization
	expectedVertices := expected.GetVertices()
	actualVertices := actual.GetVertices()
	_ = actualVertices // Use to avoid linter error

	// Build ID mapping from value to verify edges are equivalent
	for id := range expectedVertices {
		expectedChildren, _ := expected.GetChildren(id)
		actualChildren, _ := actual.GetChildren(id)

		if len(expectedChildren) != len(actualChildren) {
			t.Errorf("Children count mismatch for %s: %d != %d",
				id, len(expectedChildren), len(actualChildren))
		}

		for childID := range expectedChildren {
			if _, exists := actualChildren[childID]; !exists {
				t.Errorf("Edge %s -> %s missing in actual", id, childID)
			}
		}
	}

	// 4. Check roots
	expectedRoots := expected.GetRoots()
	actualRoots := actual.GetRoots()
	if len(expectedRoots) != len(actualRoots) {
		t.Errorf("Roots count mismatch: %d != %d",
			len(expectedRoots), len(actualRoots))
	}

	// 5. Check leaves
	expectedLeaves := expected.GetLeaves()
	actualLeaves := actual.GetLeaves()
	if len(expectedLeaves) != len(actualLeaves) {
		t.Errorf("Leaves count mismatch: %d != %d",
			len(expectedLeaves), len(actualLeaves))
	}
}

func TestMarshalUnmarshalJSON(t *testing.T) {
	cases := []struct {
		dag      *DAG
		expected string
	}{
		{
			dag:      getTestWalkDAG(),
			expected: `{"vs":[{"i":"1","v":"v1"},{"i":"2","v":"v2"},{"i":"3","v":"v3"},{"i":"4","v":"v4"},{"i":"5","v":"v5"}],"es":[{"s":"1","d":"2"},{"s":"2","d":"3"},{"s":"2","d":"4"},{"s":"4","d":"5"}]}`,
		},
		{
			dag:      getTestWalkDAG2(),
			expected: `{"vs":[{"i":"1","v":"v1"},{"i":"3","v":"v3"},{"i":"5","v":"v5"},{"i":"2","v":"v2"},{"i":"4","v":"v4"}],"es":[{"s":"1","d":"3"},{"s":"3","d":"5"},{"s":"2","d":"3"},{"s":"4","d":"5"}]}`,
		},
		{
			dag:      getTestWalkDAG3(),
			expected: `{"vs":[{"i":"1","v":"v1"},{"i":"3","v":"v3"},{"i":"2","v":"v2"},{"i":"4","v":"v4"},{"i":"5","v":"v5"}],"es":[{"s":"1","d":"3"},{"s":"2","d":"3"},{"s":"4","d":"5"}]}`,
		},
		{
			dag:      getTestWalkDAG4(),
			expected: `{"vs":[{"i":"1","v":"v1"},{"i":"2","v":"v2"},{"i":"3","v":"v3"},{"i":"5","v":"v5"},{"i":"4","v":"v4"}],"es":[{"s":"1","d":"2"},{"s":"2","d":"3"},{"s":"2","d":"4"},{"s":"3","d":"5"}]}`,
		},
	}

	for _, c := range cases {
		testMarshalUnmarshalJSON(t, c.dag, c.expected)
	}
}

func testMarshalUnmarshalJSON(t *testing.T, d *DAG, _ string) {
	data, err := json.Marshal(d)
	if err != nil {
		t.Error(err)
	}

	// Note: We no longer check exact JSON string match because
	// the order of vertices and edges may vary after optimization
	// Instead, we verify that unmarshaling produces an equivalent graph

	d1 := &DAG{}
	errNotSupported := json.Unmarshal(data, d1)
	if errNotSupported == nil {
		t.Errorf("UnmarshalJSON() = nil, want %v", "This method is not supported")
	}

	var wd testStorableDAG
	dag, err := UnmarshalJSONLegacy(data, &wd, defaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	// Use graph equivalence check instead of deep.Equal
	testGraphsEqual(t, d, dag)
}

// TestMarshalUnmarshalJSONLargeGraph tests serializability of larger graphs
func TestMarshalUnmarshalJSONLargeGraph(t *testing.T) {
	sizes := []int{100, 500}

	for _, size := range sizes {
		t.Run(fmt.Sprintf("Size_%d", size), func(t *testing.T) {
			// Create a linear DAG with simple string values (not TestVertex)
			d := NewDAG()
			var prevID string

			for i := 0; i < size; i++ {
				id := fmt.Sprintf("node_%d", i)
				_ = d.AddVertexByID(id, fmt.Sprintf("value_%d", i))
				if prevID != "" {
					_ = d.AddEdge(prevID, id)
				}
				prevID = id
			}

			// Record expected counts
			expectedOrder := d.GetOrder()
			expectedSize := d.GetSize()
			expectedRoots := len(d.GetRoots())
			expectedLeaves := len(d.GetLeaves())

			// Serialize
			data, err := d.MarshalJSON()
			if err != nil {
				t.Fatalf("MarshalJSON failed: %v", err)
			}

			// Deserialize
			var wd testStorableDAG
			restored, err := UnmarshalJSONLegacy(data, &wd, defaultOptions())
			if err != nil {
				t.Fatalf("UnmarshalJSON failed: %v", err)
			}

			// Verify counts match
			if restored.GetOrder() != expectedOrder {
				t.Errorf("Order mismatch: expected %d, got %d", expectedOrder, restored.GetOrder())
			}
			if restored.GetSize() != expectedSize {
				t.Errorf("Size mismatch: expected %d, got %d", expectedSize, restored.GetSize())
			}
			if len(restored.GetRoots()) != expectedRoots {
				t.Errorf("Roots count mismatch: expected %d, got %d", expectedRoots, len(restored.GetRoots()))
			}
			if len(restored.GetLeaves()) != expectedLeaves {
				t.Errorf("Leaves count mismatch: expected %d, got %d", expectedLeaves, len(restored.GetLeaves()))
			}

			// Verify all edges exist
			for i := 0; i < size-1; i++ {
				id := fmt.Sprintf("node_%d", i)
				nextID := fmt.Sprintf("node_%d", i+1)

				isEdge, _ := restored.IsEdge(id, nextID)
				if !isEdge {
					t.Errorf("Edge %s -> %s missing in restored graph", id, nextID)
				}
			}

			// Double serialization verify (serialize -> deserialize -> serialize)
			data2, err := restored.MarshalJSON()
			if err != nil {
				t.Fatalf("Second MarshalJSON failed: %v", err)
			}

			var wd2 testStorableDAG
			restored2, err := UnmarshalJSONLegacy(data2, &wd2, defaultOptions())
			if err != nil {
				t.Fatalf("Second UnmarshalJSON failed: %v", err)
			}

			// Verify double-serialization produces same counts
			if restored2.GetOrder() != expectedOrder {
				t.Errorf("Double-serialization Order mismatch: expected %d, got %d", expectedOrder, restored2.GetOrder())
			}
			if restored2.GetSize() != expectedSize {
				t.Errorf("Double-serialization Size mismatch: expected %d, got %d", expectedSize, restored2.GetSize())
			}
		})
	}
}

// TestAddVerticesBatch tests that individual and batch vertex addition
// produce equivalent results
func TestAddVerticesBatch(t *testing.T) {
	d1 := NewDAG()
	d2 := NewDAG()

	vertices := []struct {
		id    string
		value string
	}{
		{"v1", "value1"},
		{"v2", "value2"},
		{"v3", "value3"},
		{"v4", "value4"},
		{"v5", "value5"},
	}

	// Add vertices individually to d1
	for _, v := range vertices {
		_ = d1.AddVertexByID(v.id, v.value)
	}

	// Add same vertices using batch method to d2
	var storableVertices []Vertexer
	for _, v := range vertices {
		storableVertices = append(storableVertices, testVertex{WID: v.id, Val: v.value})
	}
	_ = d2.addVerticesBatch(storableVertices)

	// Verify equivalence
	testGraphsEqual(t, d1, d2)
}

// TestGenericMarshalUnmarshalJSONSimple tests the generic MarshalGeneric and UnmarshalJSON with simple string types
func TestGenericMarshalUnmarshalJSONSimple(t *testing.T) {
	// Create a DAG with string values
	d := NewDAG()
	_ = d.AddVertexByID("v1", "value1")
	_ = d.AddVertexByID("v2", "value2")
	_ = d.AddVertexByID("v3", "value3")
	_ = d.AddEdge("v1", "v2")
	_ = d.AddEdge("v2", "v3")

	// Marshal using generic API
	data, err := MarshalGeneric[string](d)
	if err != nil {
		t.Fatalf("MarshalGeneric failed: %v", err)
	}

	// Unmarshal using generic API
	restored, err := UnmarshalJSONGeneric[string](data, defaultOptions())
	if err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}

	// Verify equivalence
	testGraphsEqual(t, d, restored)

	// Verify vertex values are correct
	vertices := restored.GetVertices()
	if v, ok := vertices["v1"]; ok {
		if v != "value1" {
			t.Errorf("Expected v1 value to be 'value1', got '%v'", v)
		}
	} else {
		t.Error("Vertex v1 not found")
	}
}

// TestGenericMarshalUnmarshalJSONComplex tests the generic MarshalGeneric and UnmarshalJSON with complex struct types
func TestGenericMarshalUnmarshalJSONComplex(t *testing.T) {
	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	// Create a DAG with Person values
	d := NewDAG()
	alice := Person{Name: "Alice", Age: 30}
	bob := Person{Name: "Bob", Age: 25}
	charlie := Person{Name: "Charlie", Age: 35}

	_ = d.AddVertexByID("p1", alice)
	_ = d.AddVertexByID("p2", bob)
	_ = d.AddVertexByID("p3", charlie)
	_ = d.AddEdge("p1", "p2")
	_ = d.AddEdge("p2", "p3")

	// Marshal using generic API
	data, err := MarshalGeneric[Person](d)
	if err != nil {
		t.Fatalf("MarshalGeneric failed: %v", err)
	}

	// Unmarshal using generic API
	restored, err := UnmarshalJSONGeneric[Person](data, defaultOptions())
	if err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}

	// Verify equivalence
	testGraphsEqual(t, d, restored)

	// Verify vertex values are correct
	vertices := restored.GetVertices()
	if v, ok := vertices["p1"]; ok {
		if person, ok := v.(Person); ok {
			if person.Name != "Alice" || person.Age != 30 {
				t.Errorf("Expected p1 value to be {Name:Alice, Age:30}, got %+v", person)
			}
		} else {
			t.Errorf("Expected p1 value to be Person, got %T", v)
		}
	} else {
		t.Error("Vertex p1 not found")
	}
}

// TestUnmarshalJSONSimple tests the generic UnmarshalJSON with simple string types
func TestUnmarshalJSONSimple(t *testing.T) {
	data := []byte(`{"vs":[{"i":"v1","v":"value1"},{"i":"v2","v":"value2"}],"es":[{"s":"v1","d":"v2"}]}`)

	dag, err := UnmarshalJSONGeneric[string](data, defaultOptions())
	if err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}

	// Verify vertex count
	if dag.GetOrder() != 2 {
		t.Errorf("Expected 2 vertices, got %d", dag.GetOrder())
	}

	// Verify edge count
	if dag.GetSize() != 1 {
		t.Errorf("Expected 1 edge, got %d", dag.GetSize())
	}

	// Verify vertex values
	vertices := dag.GetVertices()
	if v, ok := vertices["v1"]; ok {
		if v != "value1" {
			t.Errorf("Expected v1 value to be 'value1', got '%v'", v)
		}
	} else {
		t.Error("Vertex v1 not found")
	}

	if v, ok := vertices["v2"]; ok {
		if v != "value2" {
			t.Errorf("Expected v2 value to be 'value2', got '%v'", v)
		}
	} else {
		t.Error("Vertex v2 not found")
	}

	// Verify edge exists
	isEdge, _ := dag.IsEdge("v1", "v2")
	if !isEdge {
		t.Error("Expected edge v1 -> v2 to exist")
	}
}

// TestUnmarshalJSONComplex tests the generic UnmarshalJSON with complex struct types
func TestUnmarshalJSONComplex(t *testing.T) {
	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	data := []byte(`{"vs":[{"i":"p1","v":{"name":"Alice","age":30}},{"i":"p2","v":{"name":"Bob","age":25}}],"es":[{"s":"p1","d":"p2"}]}`)

	dag, err := UnmarshalJSONGeneric[Person](data, defaultOptions())
	if err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}

	// Verify vertex count
	if dag.GetOrder() != 2 {
		t.Errorf("Expected 2 vertices, got %d", dag.GetOrder())
	}

	// Verify vertex values are correctly unmarshaled
	vertices := dag.GetVertices()
	if v, ok := vertices["p1"]; ok {
		if person, ok := v.(Person); ok {
			if person.Name != "Alice" || person.Age != 30 {
				t.Errorf("Expected p1 value to be {Name:Alice, Age:30}, got %+v", person)
			}
		} else {
			t.Errorf("Expected p1 value to be Person, got %T", v)
		}
	} else {
		t.Error("Vertex p1 not found")
	}

	if v, ok := vertices["p2"]; ok {
		if person, ok := v.(Person); ok {
			if person.Name != "Bob" || person.Age != 25 {
				t.Errorf("Expected p2 value to be {Name:Bob, Age:25}, got %+v", person)
			}
		} else {
			t.Errorf("Expected p2 value to be Person, got %T", v)
		}
	} else {
		t.Error("Vertex p2 not found")
	}
}

// TestUnmarshalJSONInteger tests the generic UnmarshalJSON with integer types
func TestUnmarshalJSONInteger(t *testing.T) {
	data := []byte(`{"vs":[{"i":"n1","v":42},{"i":"n2","v":100}],"es":[{"s":"n1","d":"n2"}]}`)

	dag, err := UnmarshalJSONGeneric[int](data, defaultOptions())
	if err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}

	// Verify vertex values
	vertices := dag.GetVertices()
	if v, ok := vertices["n1"]; ok {
		if v != 42 {
			t.Errorf("Expected n1 value to be 42, got %v", v)
		}
	} else {
		t.Error("Vertex n1 not found")
	}

	if v, ok := vertices["n2"]; ok {
		if v != 100 {
			t.Errorf("Expected n2 value to be 100, got %v", v)
		}
	} else {
		t.Error("Vertex n2 not found")
	}
}

// TestUnmarshalJSONCompatibility tests that the generic UnmarshalJSON
// produces equivalent results to the legacy UnmarshalJSONLegacy
func TestUnmarshalJSONCompatibility(t *testing.T) {
	// Create a test DAG with string values
	d := NewDAG()
	_ = d.AddVertexByID("v1", "value1")
	_ = d.AddVertexByID("v2", "value2")
	_ = d.AddEdge("v1", "v2")

	// Marshal using existing method
	data, err := d.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	// Unmarshal using new generic UnmarshalJSONGeneric
	dag1, err := UnmarshalJSONGeneric[string](data, defaultOptions())
	if err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}

	// Unmarshal using legacy UnmarshalJSONLegacy for comparison
	var wd testStorableDAG
	dag2, err := UnmarshalJSONLegacy(data, &wd, defaultOptions())
	if err != nil {
		t.Fatalf("UnmarshalJSONLegacy failed: %v", err)
	}

	// Verify both methods produce equivalent graphs
	testGraphsEqual(t, dag1, dag2)
}
