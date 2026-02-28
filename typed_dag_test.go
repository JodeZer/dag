package dag

import (
	"encoding/json"
	"testing"
)

// TestTypedDAGTypeSafety tests that TypedDAG provides compile-time type safety
func TestTypedDAGTypeSafety(t *testing.T) {
	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	dag := New[Person]()
	alice := Person{Name: "Alice", Age: 30}
	bob := Person{Name: "Bob", Age: 25}

	// Type-safe vertex addition
	err := dag.AddVertexByID("p1", alice)
	if err != nil {
		t.Fatalf("AddVertexByID failed: %v", err)
	}

	err = dag.AddVertexByID("p2", bob)
	if err != nil {
		t.Fatalf("AddVertexByID failed: %v", err)
	}

	// Type-safe vertex access - no type assertion needed
	person, err := dag.GetVertex("p1")
	if err != nil {
		t.Fatalf("GetVertex failed: %v", err)
	}

	// Compile-time type checking - person is of type Person, not interface{}
	if person.Name != "Alice" {
		t.Errorf("Expected person.Name to be 'Alice', got '%s'", person.Name)
	}
	if person.Age != 30 {
		t.Errorf("Expected person.Age to be 30, got %d", person.Age)
	}

	// Type-safe GetVertices - returns map[string]Person
	vertices := dag.GetVertices()
	if len(vertices) != 2 {
		t.Errorf("Expected 2 vertices, got %d", len(vertices))
	}

	if p, ok := vertices["p1"]; ok {
		if p.Name != "Alice" {
			t.Errorf("Expected p1.Name to be 'Alice', got '%s'", p.Name)
		}
	} else {
		t.Error("Vertex p1 not found")
	}
}

// TestTypedDAGMarshalJSON tests auto-infer serialization without generic parameter
func TestTypedDAGMarshalJSON(t *testing.T) {
	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	dag := New[Person]()
	err := dag.AddVertexByID("p1", Person{Name: "Alice", Age: 30})
	if err != nil {
		t.Fatalf("AddVertexByID failed: %v", err)
	}
	err = dag.AddVertexByID("p2", Person{Name: "Bob", Age: 25})
	if err != nil {
		t.Fatalf("AddVertexByID failed: %v", err)
	}
	err = dag.AddEdge("p1", "p2")
	if err != nil {
		t.Fatalf("AddEdge failed: %v", err)
	}

	// Auto-infer serialization - no generic parameter needed
	data, err := dag.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	// Verify JSON structure
	var result map[string]json.RawMessage
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Check vertices exist
	if _, ok := result["vs"]; !ok {
		t.Error("Missing 'vs' field in JSON")
	}
	if _, ok := result["es"]; !ok {
		t.Error("Missing 'es' field in JSON")
	}
}

// TestTypedDAGUnmarshalJSON tests deserialization with type parameter
func TestTypedDAGUnmarshalJSON(t *testing.T) {
	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	data := []byte(`{"vs":[{"i":"p1","v":{"name":"Alice","age":30}},{"i":"p2","v":{"name":"Bob","age":25}}],"es":[{"s":"p1","d":"p2"}]}`)

	// Deserialize with type parameter (reasonable - need to know the type)
	restored, err := UnmarshalJSON[Person](data, defaultOptions())
	if err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}

	// Type-safe vertex access
	alice, err := restored.GetVertex("p1")
	if err != nil {
		t.Fatalf("GetVertex failed: %v", err)
	}

	if alice.Name != "Alice" {
		t.Errorf("Expected alice.Name to be 'Alice', got '%s'", alice.Name)
	}
	if alice.Age != 30 {
		t.Errorf("Expected alice.Age to be 30, got %d", alice.Age)
	}

	// Verify edge exists
	isEdge, _ := restored.IsEdge("p1", "p2")
	if !isEdge {
		t.Error("Expected edge p1 -> p2 to exist")
	}
}

// TestTypedDAGMarshalUnmarshalRoundtrip tests full serialization roundtrip
func TestTypedDAGMarshalUnmarshalRoundtrip(t *testing.T) {
	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	// Create original DAG
	original := New[Person]()
	err := original.AddVertexByID("p1", Person{Name: "Alice", Age: 30})
	if err != nil {
		t.Fatalf("AddVertexByID failed: %v", err)
	}
	err = original.AddVertexByID("p2", Person{Name: "Bob", Age: 25})
	if err != nil {
		t.Fatalf("AddVertexByID failed: %v", err)
	}
	err = original.AddVertexByID("p3", Person{Name: "Charlie", Age: 35})
	if err != nil {
		t.Fatalf("AddVertexByID failed: %v", err)
	}
	err = original.AddEdge("p1", "p2")
	if err != nil {
		t.Fatalf("AddEdge failed: %v", err)
	}
	err = original.AddEdge("p2", "p3")
	if err != nil {
		t.Fatalf("AddEdge failed: %v", err)
	}

	// Serialize
	data, err := original.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	// Deserialize
	restored, err := UnmarshalJSON[Person](data, defaultOptions())
	if err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}

	// Verify structure
	if restored.GetOrder() != original.GetOrder() {
		t.Errorf("Order mismatch: expected %d, got %d", original.GetOrder(), restored.GetOrder())
	}
	if restored.GetSize() != original.GetSize() {
		t.Errorf("Size mismatch: expected %d, got %d", original.GetSize(), restored.GetSize())
	}

	// Verify vertices
	originalVertices := original.GetVertices()
	restoredVertices := restored.GetVertices()

	if len(originalVertices) != len(restoredVertices) {
		t.Errorf("Vertices count mismatch: %d != %d", len(originalVertices), len(restoredVertices))
	}

	for id, p := range originalVertices {
		restoredP, ok := restoredVertices[id]
		if !ok {
			t.Errorf("Vertex %s not found in restored DAG", id)
			continue
		}
		if p.Name != restoredP.Name || p.Age != restoredP.Age {
			t.Errorf("Vertex %s mismatch: expected %+v, got %+v", id, p, restoredP)
		}
	}
}

// TestTypedDAGWithSimpleTypes tests TypedDAG with simple types
func TestTypedDAGWithSimpleTypes(t *testing.T) {
	// Test with string type
	dag := New[string]()
	err := dag.AddVertexByID("v1", "value1")
	if err != nil {
		t.Fatalf("AddVertexByID failed: %v", err)
	}
	err = dag.AddVertexByID("v2", "value2")
	if err != nil {
		t.Fatalf("AddVertexByID failed: %v", err)
	}
	err = dag.AddEdge("v1", "v2")
	if err != nil {
		t.Fatalf("AddEdge failed: %v", err)
	}

	v, err := dag.GetVertex("v1")
	if err != nil {
		t.Fatalf("GetVertex failed: %v", err)
	}
	if v != "value1" {
		t.Errorf("Expected 'value1', got '%s'", v)
	}

	// Test with int type
	dag2 := New[int]()
	err = dag2.AddVertexByID("n1", 42)
	if err != nil {
		t.Fatalf("AddVertexByID failed: %v", err)
	}
	err = dag2.AddVertexByID("n2", 100)
	if err != nil {
		t.Fatalf("AddVertexByID failed: %v", err)
	}
	err = dag2.AddEdge("n1", "n2")
	if err != nil {
		t.Fatalf("AddEdge failed: %v", err)
	}

	n, err := dag2.GetVertex("n1")
	if err != nil {
		t.Fatalf("GetVertex failed: %v", err)
	}
	if n != 42 {
		t.Errorf("Expected 42, got %d", n)
	}
}

// TestTypedDAGWithPointerType tests TypedDAG with pointer types
func TestTypedDAGWithPointerType(t *testing.T) {
	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	dag := New[*Person]()
	alice := &Person{Name: "Alice", Age: 30}
	bob := &Person{Name: "Bob", Age: 25}

	err := dag.AddVertexByID("p1", alice)
	if err != nil {
		t.Fatalf("AddVertexByID failed: %v", err)
	}
	err = dag.AddVertexByID("p2", bob)
	if err != nil {
		t.Fatalf("AddVertexByID failed: %v", err)
	}
	err = dag.AddEdge("p1", "p2")
	if err != nil {
		t.Fatalf("AddEdge failed: %v", err)
	}

	// Type-safe vertex access
	p, err := dag.GetVertex("p1")
	if err != nil {
		t.Fatalf("GetVertex failed: %v", err)
	}
	if p.Name != "Alice" {
		t.Errorf("Expected p.Name to be 'Alice', got '%s'", p.Name)
	}
}

// TestTypedDAGGetLeaves tests GetLeaves with TypedDAG
func TestTypedDAGGetLeaves(t *testing.T) {
	type Task struct {
		Name string
	}

	dag := New[Task]()
	err := dag.AddVertexByID("t1", Task{Name: "Task 1"})
	if err != nil {
		t.Fatalf("AddVertexByID failed: %v", err)
	}
	err = dag.AddVertexByID("t2", Task{Name: "Task 2"})
	if err != nil {
		t.Fatalf("AddVertexByID failed: %v", err)
	}
	err = dag.AddVertexByID("t3", Task{Name: "Task 3"})
	if err != nil {
		t.Fatalf("AddVertexByID failed: %v", err)
	}
	err = dag.AddEdge("t1", "t2")
	if err != nil {
		t.Fatalf("AddEdge failed: %v", err)
	}
	err = dag.AddEdge("t1", "t3")
	if err != nil {
		t.Fatalf("AddEdge failed: %v", err)
	}

	leaves := dag.GetLeaves()
	if len(leaves) != 2 {
		t.Errorf("Expected 2 leaves, got %d", len(leaves))
	}

	// Verify leaf types are correct
	for _, task := range leaves {
		if task.Name != "Task 2" && task.Name != "Task 3" {
			t.Errorf("Unexpected leaf task: %s", task.Name)
		}
	}
}

// TestTypedDAGGetRoots tests GetRoots with TypedDAG
func TestTypedDAGGetRoots(t *testing.T) {
	type Task struct {
		Name string
	}

	dag := New[Task]()
	err := dag.AddVertexByID("t1", Task{Name: "Task 1"})
	if err != nil {
		t.Fatalf("AddVertexByID failed: %v", err)
	}
	err = dag.AddVertexByID("t2", Task{Name: "Task 2"})
	if err != nil {
		t.Fatalf("AddVertexByID failed: %v", err)
	}
	err = dag.AddVertexByID("t3", Task{Name: "Task 3"})
	if err != nil {
		t.Fatalf("AddVertexByID failed: %v", err)
	}
	err = dag.AddEdge("t1", "t2")
	if err != nil {
		t.Fatalf("AddEdge failed: %v", err)
	}
	err = dag.AddEdge("t2", "t3")
	if err != nil {
		t.Fatalf("AddEdge failed: %v", err)
	}

	roots := dag.GetRoots()
	if len(roots) != 1 {
		t.Errorf("Expected 1 root, got %d", len(roots))
	}

	// Verify root type is correct
	for _, task := range roots {
		if task.Name != "Task 1" {
			t.Errorf("Unexpected root task: %s", task.Name)
		}
	}
}

// TestTypedDAGGetChildren tests GetChildren with TypedDAG
func TestTypedDAGGetChildren(t *testing.T) {
	type Task struct {
		Name string
	}

	dag := New[Task]()
	err := dag.AddVertexByID("t1", Task{Name: "Task 1"})
	if err != nil {
		t.Fatalf("AddVertexByID failed: %v", err)
	}
	err = dag.AddVertexByID("t2", Task{Name: "Task 2"})
	if err != nil {
		t.Fatalf("AddVertexByID failed: %v", err)
	}
	err = dag.AddVertexByID("t3", Task{Name: "Task 3"})
	if err != nil {
		t.Fatalf("AddVertexByID failed: %v", err)
	}
	err = dag.AddEdge("t1", "t2")
	if err != nil {
		t.Fatalf("AddEdge failed: %v", err)
	}
	err = dag.AddEdge("t1", "t3")
	if err != nil {
		t.Fatalf("AddEdge failed: %v", err)
	}

	children, err := dag.GetChildren("t1")
	if err != nil {
		t.Fatalf("GetChildren failed: %v", err)
	}

	if len(children) != 2 {
		t.Errorf("Expected 2 children, got %d", len(children))
	}

	// Verify child types are correct
	for id, task := range children {
		if task.Name != "Task 2" && task.Name != "Task 3" {
			t.Errorf("Unexpected child task: %s", task.Name)
		}
		if id != "t2" && id != "t3" {
			t.Errorf("Unexpected child id: %s", id)
		}
	}
}

// TestTypedDAGGetParents tests GetParents with TypedDAG
func TestTypedDAGGetParents(t *testing.T) {
	type Task struct {
		Name string
	}

	dag := New[Task]()
	err := dag.AddVertexByID("t1", Task{Name: "Task 1"})
	if err != nil {
		t.Fatalf("AddVertexByID failed: %v", err)
	}
	err = dag.AddVertexByID("t2", Task{Name: "Task 2"})
	if err != nil {
		t.Fatalf("AddVertexByID failed: %v", err)
	}
	err = dag.AddVertexByID("t3", Task{Name: "Task 3"})
	if err != nil {
		t.Fatalf("AddVertexByID failed: %v", err)
	}
	err = dag.AddEdge("t1", "t3")
	if err != nil {
		t.Fatalf("AddEdge failed: %v", err)
	}
	err = dag.AddEdge("t2", "t3")
	if err != nil {
		t.Fatalf("AddEdge failed: %v", err)
	}

	parents, err := dag.GetParents("t3")
	if err != nil {
		t.Fatalf("GetParents failed: %v", err)
	}

	if len(parents) != 2 {
		t.Errorf("Expected 2 parents, got %d", len(parents))
	}

	// Verify parent types are correct
	for id, task := range parents {
		if task.Name != "Task 1" && task.Name != "Task 2" {
			t.Errorf("Unexpected parent task: %s", task.Name)
		}
		if id != "t1" && id != "t2" {
			t.Errorf("Unexpected parent id: %s", id)
		}
	}
}

// TestTypedDAGGetDescendantsGraph tests GetDescendantsGraph with TypedDAG
func TestTypedDAGGetDescendantsGraph(t *testing.T) {
	type Task struct {
		Name string
	}

	dag := New[Task]()
	err := dag.AddVertexByID("t1", Task{Name: "Task 1"})
	if err != nil {
		t.Fatalf("AddVertexByID failed: %v", err)
	}
	err = dag.AddVertexByID("t2", Task{Name: "Task 2"})
	if err != nil {
		t.Fatalf("AddVertexByID failed: %v", err)
	}
	err = dag.AddVertexByID("t3", Task{Name: "Task 3"})
	if err != nil {
		t.Fatalf("AddVertexByID failed: %v", err)
	}
	err = dag.AddEdge("t1", "t2")
	if err != nil {
		t.Fatalf("AddEdge failed: %v", err)
	}
	err = dag.AddEdge("t2", "t3")
	if err != nil {
		t.Fatalf("AddEdge failed: %v", err)
	}

	subgraph, rootID, err := dag.GetDescendantsGraph("t1")
	if err != nil {
		t.Fatalf("GetDescendantsGraph failed: %v", err)
	}

	if subgraph.GetOrder() != 3 {
		t.Errorf("Expected subgraph to have 3 vertices, got %d", subgraph.GetOrder())
	}

	if rootID == "" {
		t.Error("Expected root ID to be non-empty")
	}

	// Verify type-safe access using the returned root ID
	task, err := subgraph.GetVertex(rootID)
	if err != nil {
		t.Fatalf("GetVertex failed: %v", err)
	}
	if task.Name != "Task 1" {
		t.Errorf("Expected task.Name to be 'Task 1', got '%s'", task.Name)
	}
}

// TestTypedDAGGetAncestorsGraph tests GetAncestorsGraph with TypedDAG
func TestTypedDAGGetAncestorsGraph(t *testing.T) {
	type Task struct {
		Name string
	}

	dag := New[Task]()
	err := dag.AddVertexByID("t1", Task{Name: "Task 1"})
	if err != nil {
		t.Fatalf("AddVertexByID failed: %v", err)
	}
	err = dag.AddVertexByID("t2", Task{Name: "Task 2"})
	if err != nil {
		t.Fatalf("AddVertexByID failed: %v", err)
	}
	err = dag.AddVertexByID("t3", Task{Name: "Task 3"})
	if err != nil {
		t.Fatalf("AddVertexByID failed: %v", err)
	}
	err = dag.AddEdge("t1", "t2")
	if err != nil {
		t.Fatalf("AddEdge failed: %v", err)
	}
	err = dag.AddEdge("t2", "t3")
	if err != nil {
		t.Fatalf("AddEdge failed: %v", err)
	}

	subgraph, leafID, err := dag.GetAncestorsGraph("t3")
	if err != nil {
		t.Fatalf("GetAncestorsGraph failed: %v", err)
	}

	if subgraph.GetOrder() != 3 {
		t.Errorf("Expected subgraph to have 3 vertices, got %d", subgraph.GetOrder())
	}

	if leafID == "" {
		t.Error("Expected leaf ID to be non-empty")
	}

	// Verify type-safe access using the returned leaf ID
	task, err := subgraph.GetVertex(leafID)
	if err != nil {
		t.Fatalf("GetVertex failed: %v", err)
	}
	if task.Name != "Task 3" {
		t.Errorf("Expected task.Name to be 'Task 3', got '%s'", task.Name)
	}
}

// TestTypedDAGCopy tests Copy with TypedDAG
func TestTypedDAGCopy(t *testing.T) {
	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	original := New[Person]()
	err := original.AddVertexByID("p1", Person{Name: "Alice", Age: 30})
	if err != nil {
		t.Fatalf("AddVertexByID failed: %v", err)
	}
	err = original.AddVertexByID("p2", Person{Name: "Bob", Age: 25})
	if err != nil {
		t.Fatalf("AddVertexByID failed: %v", err)
	}
	err = original.AddEdge("p1", "p2")
	if err != nil {
		t.Fatalf("AddEdge failed: %v", err)
	}

	copy, err := original.Copy()
	if err != nil {
		t.Fatalf("Copy failed: %v", err)
	}

	// Verify copy has same structure
	if copy.GetOrder() != original.GetOrder() {
		t.Errorf("Order mismatch: expected %d, got %d", original.GetOrder(), copy.GetOrder())
	}
	if copy.GetSize() != original.GetSize() {
		t.Errorf("Size mismatch: expected %d, got %d", original.GetSize(), copy.GetSize())
	}

	// Verify type-safe access - need to find the vertex ID in the copy
	vertices := copy.GetVertices()
	if len(vertices) != 2 {
		t.Errorf("Expected 2 vertices, got %d", len(vertices))
	}

	// Find Alice in the copy and verify her details
	foundAlice := false
	for id, person := range vertices {
		if person.Name == "Alice" && person.Age == 30 {
			foundAlice = true
			// Also verify GetVertex works
			alice, err := copy.GetVertex(id)
			if err != nil {
				t.Fatalf("GetVertex failed: %v", err)
			}
			if alice.Name != "Alice" {
				t.Errorf("Expected alice.Name to be 'Alice', got '%s'", alice.Name)
			}
			break
		}
	}

	if !foundAlice {
		t.Error("Alice not found in copy")
	}

	// Verify copy is independent
	_ = copy.AddVertexByID("p3", Person{Name: "Charlie", Age: 35})
	if original.GetOrder() == copy.GetOrder() {
		t.Error("Copy should be independent of original")
	}
}

// TestTypedDAGToDAG tests ToDAG conversion for backward compatibility
func TestTypedDAGToDAG(t *testing.T) {
	type Person struct {
		Name string
	}

	dag := New[Person]()
	_ = dag.AddVertexByID("p1", Person{Name: "Alice"})

	inner := dag.ToDAG()
	if inner.GetOrder() != 1 {
		t.Errorf("Expected inner DAG to have 1 vertex, got %d", inner.GetOrder())
	}
}

// TestTypedDAGNewWithOptions tests creating TypedDAG with custom options
func TestTypedDAGNewWithOptions(t *testing.T) {
	type Person struct {
		Name string
	}

	options := Options{
		VertexHashFunc: func(v interface{}) interface{} {
			if p, ok := v.(Person); ok {
				return p.Name // Hash by name instead of full struct
			}
			return v
		},
	}

	dag := NewWithOptions[Person](options)
	_ = dag.AddVertexByID("p1", Person{Name: "Alice"})

	if dag.GetOrder() != 1 {
		t.Errorf("Expected 1 vertex, got %d", dag.GetOrder())
	}
}

// TestTypedDAGIsEmpty tests IsEmpty method
func TestTypedDAGIsEmpty(t *testing.T) {
	type Person struct {
		Name string
	}

	dag := New[Person]()

	if !dag.IsEmpty() {
		t.Error("Expected empty DAG")
	}

	_ = dag.AddVertexByID("p1", Person{Name: "Alice"})

	if dag.IsEmpty() {
		t.Error("Expected non-empty DAG")
	}
}

// TestTypedDAGSimpleRoundtrip tests a simple roundtrip with TypedDAG
func TestTypedDAGSimpleRoundtrip(t *testing.T) {
	// Create and serialize
	original := New[string]()
	err := original.AddVertexByID("a", "A")
	if err != nil {
		t.Fatalf("AddVertexByID failed: %v", err)
	}
	err = original.AddVertexByID("b", "B")
	if err != nil {
		t.Fatalf("AddVertexByID failed: %v", err)
	}
	err = original.AddEdge("a", "b")
	if err != nil {
		t.Fatalf("AddEdge failed: %v", err)
	}

	data, err := original.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	// Deserialize and verify
	restored, err := UnmarshalJSON[string](data, defaultOptions())
	if err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}

	a, err := restored.GetVertex("a")
	if err != nil {
		t.Fatalf("GetVertex failed: %v", err)
	}
	if a != "A" {
		t.Errorf("Expected 'A', got '%s'", a)
	}
}

// TestTypedDAGCompatibilityWithDAG tests interoperability between TypedDAG and DAG
func TestTypedDAGCompatibilityWithDAG(t *testing.T) {
	type Task struct {
		Name string
	}

	// Create a TypedDAG
	typed := New[Task]()
	err := typed.AddVertexByID("t1", Task{Name: "Task 1"})
	if err != nil {
		t.Fatalf("AddVertexByID failed: %v", err)
	}
	err = typed.AddVertexByID("t2", Task{Name: "Task 2"})
	if err != nil {
		t.Fatalf("AddVertexByID failed: %v", err)
	}
	err = typed.AddEdge("t1", "t2")
	if err != nil {
		t.Fatalf("AddEdge failed: %v", err)
	}

	// Convert to DAG using ToDAG()
	dag := typed.ToDAG()

	// Verify the underlying DAG has the same structure
	if dag.GetOrder() != 2 {
		t.Errorf("DAG order mismatch: got %d, want 2", dag.GetOrder())
	}
	if dag.GetSize() != 1 {
		t.Errorf("DAG size mismatch: got %d, want 1", dag.GetSize())
	}

	// Verify vertices exist
	v, err := dag.GetVertex("t1")
	if err != nil {
		t.Fatalf("GetVertex failed: %v", err)
	}
	task, ok := v.(Task)
	if !ok {
		t.Fatal("Vertex is not of type Task")
	}
	if task.Name != "Task 1" {
		t.Errorf("Task name mismatch: got '%s', want 'Task 1'", task.Name)
	}

	// Verify edge exists
	isEdge, _ := dag.IsEdge("t1", "t2")
	if !isEdge {
		t.Error("Expected edge t1 -> t2 to exist")
	}
}

// TestTypedDAGUnmarshalFromDAGMarshal tests that data marshaled from DAG
// can be unmarshaled by TypedDAG
func TestTypedDAGUnmarshalFromDAGMarshal(t *testing.T) {
	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	// Create a regular DAG
	dag := NewDAG()
	err := dag.AddVertexByID("p1", Person{Name: "Alice", Age: 30})
	if err != nil {
		t.Fatalf("AddVertexByID failed: %v", err)
	}
	err = dag.AddVertexByID("p2", Person{Name: "Bob", Age: 25})
	if err != nil {
		t.Fatalf("AddVertexByID failed: %v", err)
	}
	err = dag.AddEdge("p1", "p2")
	if err != nil {
		t.Fatalf("AddEdge failed: %v", err)
	}

	// Marshal using regular DAG Marshal
	data, err := dag.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	// Unmarshal using TypedDAG UnmarshalJSON
	typed, err := UnmarshalJSON[Person](data, defaultOptions())
	if err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}

	// Verify data
	alice, err := typed.GetVertex("p1")
	if err != nil {
		t.Fatalf("GetVertex failed: %v", err)
	}
	if alice.Name != "Alice" || alice.Age != 30 {
		t.Errorf("Person data mismatch: got %+v, want {Name:Alice, Age:30}", alice)
	}

	// Verify edge
	isEdge, _ := typed.IsEdge("p1", "p2")
	if !isEdge {
		t.Error("Expected edge p1 -> p2 to exist")
	}
}

// TestTypedDAGMarshalToDAGUnmarshal tests that data marshaled from TypedDAG
// can be unmarshaled by DAG (using the internal unmarshal functionality)
func TestTypedDAGMarshalToDAGUnmarshal(t *testing.T) {
	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	// Create a TypedDAG
	typed := New[Person]()
	err := typed.AddVertexByID("p1", Person{Name: "Alice", Age: 30})
	if err != nil {
		t.Fatalf("AddVertexByID failed: %v", err)
	}
	err = typed.AddVertexByID("p2", Person{Name: "Bob", Age: 25})
	if err != nil {
		t.Fatalf("AddVertexByID failed: %v", err)
	}
	err = typed.AddEdge("p1", "p2")
	if err != nil {
		t.Fatalf("AddEdge failed: %v", err)
	}

	// Marshal using TypedDAG MarshalJSON
	data, err := typed.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	// Unmarshal using TypedDAG UnmarshalJSON and convert to DAG
	typed2, err := UnmarshalJSON[Person](data, defaultOptions())
	if err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}

	// Convert to DAG using ToDAG()
	dag := typed2.ToDAG()

	// Verify data
	v, err := dag.GetVertex("p1")
	if err != nil {
		t.Fatalf("GetVertex failed: %v", err)
	}
	person, ok := v.(Person)
	if !ok {
		t.Fatal("Vertex is not of type Person")
	}
	if person.Name != "Alice" || person.Age != 30 {
		t.Errorf("Person data mismatch: got %+v, want {Name:Alice, Age:30}", person)
	}

	// Verify edge
	isEdge, _ := dag.IsEdge("p1", "p2")
	if !isEdge {
		t.Error("Expected edge p1 -> p2 to exist")
	}
}