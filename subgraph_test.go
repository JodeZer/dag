package dag

import (
	"testing"
)

func TestGetDescendantsGraphByDepth(t *testing.T) {
	// Create a DAG with multiple levels
	dag := NewGenericDAG[string]()
	dag.AddVertexByID("root", "Root")
	dag.AddVertexByID("child1", "Child 1")
	dag.AddVertexByID("child2", "Child 2")
	dag.AddVertexByID("grandchild1", "Grandchild 1")
	dag.AddVertexByID("grandchild2", "Grandchild 2")
	dag.AddVertexByID("greatgrandchild1", "Great-Grandchild 1")
	dag.AddEdge("root", "child1")
	dag.AddEdge("root", "child2")
	dag.AddEdge("child1", "grandchild1")
	dag.AddEdge("child2", "grandchild2")
	dag.AddEdge("grandchild1", "greatgrandchild1")

	tests := []struct {
		name      string
		maxDepth  int
		wantOrder int // expected number of vertices
		wantSize  int // expected number of edges
	}{
		{
			name:      "depth 0 - only root",
			maxDepth:  0,
			wantOrder: 1,
			wantSize:  0,
		},
		{
			name:      "depth 1 - root and children",
			maxDepth:  1,
			wantOrder: 3,
			wantSize:  2,
		},
		{
			name:      "depth 2 - root, children, grandchildren",
			maxDepth:  2,
			wantOrder: 5,
			wantSize:  4,
		},
		{
			name:      "depth 3 - full graph",
			maxDepth:  3,
			wantOrder: 6,
			wantSize:  5,
		},
		{
			name:      "negative depth - unlimited",
			maxDepth:  -1,
			wantOrder: 6,
			wantSize:  5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subgraph, rootID, err := dag.GetDescendantsGraphByDepth("root", tt.maxDepth)
			if err != nil {
				t.Fatalf("GetDescendantsGraphByDepth() error = %v", err)
			}

			if rootID != "root" {
				t.Errorf("GetDescendantsGraphByDepth() rootID = %v, want %v", rootID, "root")
			}

			if order := subgraph.GetOrder(); order != tt.wantOrder {
				t.Errorf("GetDescendantsGraphByDepth() order = %v, want %v", order, tt.wantOrder)
			}

			if size := subgraph.GetSize(); size != tt.wantSize {
				t.Errorf("GetDescendantsGraphByDepth() size = %v, want %v", size, tt.wantSize)
			}

			// Verify root exists in subgraph
			rootValue, err := subgraph.GetVertex(rootID)
			if err != nil {
				t.Errorf("GetDescendantsGraphByDepth() root not found: %v", err)
			}
			if rootValue != "Root" {
				t.Errorf("GetDescendantsGraphByDepth() root value = %v, want %v", rootValue, "Root")
			}
		})
	}
}

func TestGetDescendantsGraphByDepth_DiamondPattern(t *testing.T) {
	// Create a DAG with a diamond pattern:
	//     A
	//    / \
	//   B   C
	//    \ /
	//     D
	dag := NewGenericDAG[string]()
	dag.AddVertexByID("A", "A")
	dag.AddVertexByID("B", "B")
	dag.AddVertexByID("C", "C")
	dag.AddVertexByID("D", "D")
	dag.AddEdge("A", "B")
	dag.AddEdge("A", "C")
	dag.AddEdge("B", "D")
	dag.AddEdge("C", "D")

	// Test with depth 2 (should include D)
	subgraph, rootID, err := dag.GetDescendantsGraphByDepth("A", 2)
	if err != nil {
		t.Fatalf("GetDescendantsGraphByDepth() error = %v", err)
	}

	if order := subgraph.GetOrder(); order != 4 {
		t.Errorf("GetDescendantsGraphByDepth() order = %v, want %v", order, 4)
	}

	// Verify D exists in subgraph (it should only be added once)
	_, err = subgraph.GetVertex("D")
	if err != nil {
		t.Errorf("GetDescendantsGraphByDepth() D not found: %v", err)
	}

	// Verify edges are correct (should have 4 edges: A->B, A->C, B->D, C->D)
	if size := subgraph.GetSize(); size != 4 {
		t.Errorf("GetDescendantsGraphByDepth() size = %v, want %v", size, 4)
	}

	if rootID != "A" {
		t.Errorf("GetDescendantsGraphByDepth() rootID = %v, want %v", rootID, "A")
	}
}

func TestGetAncestorsGraphByDepth(t *testing.T) {
	// Create a DAG with multiple levels
	dag := NewGenericDAG[string]()
	dag.AddVertexByID("root", "Root")
	dag.AddVertexByID("child1", "Child 1")
	dag.AddVertexByID("child2", "Child 2")
	dag.AddVertexByID("grandchild1", "Grandchild 1")
	dag.AddVertexByID("grandchild2", "Grandchild 2")
	dag.AddVertexByID("greatgrandchild1", "Great-Grandchild 1")
	dag.AddEdge("root", "child1")
	dag.AddEdge("root", "child2")
	dag.AddEdge("child1", "grandchild1")
	dag.AddEdge("child2", "grandchild2")
	dag.AddEdge("grandchild1", "greatgrandchild1")

	tests := []struct {
		name      string
		maxDepth  int
		wantOrder int // expected number of vertices
		wantSize  int // expected number of edges
	}{
		{
			name:      "depth 0 - only leaf",
			maxDepth:  0,
			wantOrder: 1,
			wantSize:  0,
		},
		{
			name:      "depth 1 - leaf and parent",
			maxDepth:  1,
			wantOrder: 2,
			wantSize:  1,
		},
		{
			name:      "depth 2 - leaf, parent, grandparent",
			maxDepth:  2,
			wantOrder: 3,
			wantSize:  2,
		},
		{
			name:      "depth 3 - full ancestors",
			maxDepth:  3,
			wantOrder: 4,
			wantSize:  3,
		},
		{
			name:      "negative depth - unlimited",
			maxDepth:  -1,
			wantOrder: 4,
			wantSize:  3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subgraph, leafID, err := dag.GetAncestorsGraphByDepth("greatgrandchild1", tt.maxDepth)
			if err != nil {
				t.Fatalf("GetAncestorsGraphByDepth() error = %v", err)
			}

			if leafID != "greatgrandchild1" {
				t.Errorf("GetAncestorsGraphByDepth() leafID = %v, want %v", leafID, "greatgrandchild1")
			}

			if order := subgraph.GetOrder(); order != tt.wantOrder {
				t.Errorf("GetAncestorsGraphByDepth() order = %v, want %v", order, tt.wantOrder)
			}

			if size := subgraph.GetSize(); size != tt.wantSize {
				t.Errorf("GetAncestorsGraphByDepth() size = %v, want %v", size, tt.wantSize)
			}

			// Verify leaf exists in subgraph
			leafValue, err := subgraph.GetVertex(leafID)
			if err != nil {
				t.Errorf("GetAncestorsGraphByDepth() leaf not found: %v", err)
			}
			if leafValue != "Great-Grandchild 1" {
				t.Errorf("GetAncestorsGraphByDepth() leaf value = %v, want %v", leafValue, "Great-Grandchild 1")
			}
		})
	}
}

func TestGetEdges(t *testing.T) {
	dag := NewGenericDAG[string]()
	dag.AddVertexByID("A", "A")
	dag.AddVertexByID("B", "B")
	dag.AddVertexByID("C", "C")
	dag.AddEdge("A", "B")
	dag.AddEdge("B", "C")

	edges := dag.GetEdges()

	if count := edges.Count(); count != 2 {
		t.Errorf("GetEdges() count = %v, want %v", count, 2)
	}

	// Verify edges
	wantEdges := []struct{ src, dst string }{
		{"A", "B"},
		{"B", "C"},
	}

	if len(edges.Edges) != len(wantEdges) {
		t.Fatalf("GetEdges() returned %d edges, want %d", len(edges.Edges), len(wantEdges))
	}

	for i, want := range wantEdges {
		if edges.Edges[i].SrcID != want.src {
			t.Errorf("GetEdges()[%d].SrcID = %v, want %v", i, edges.Edges[i].SrcID, want.src)
		}
		if edges.Edges[i].DstID != want.dst {
			t.Errorf("GetEdges()[%d].DstID = %v, want %v", i, edges.Edges[i].DstID, want.dst)
		}
	}
}

func TestGetVerticesList(t *testing.T) {
	dag := NewGenericDAG[string]()
	dag.AddVertexByID("A", "Value A")
	dag.AddVertexByID("B", "Value B")
	dag.AddVertexByID("C", "Value C")

	nodes := dag.GetVerticesList()

	if count := nodes.Count(); count != 3 {
		t.Errorf("GetVerticesList() count = %v, want %v", count, 3)
	}

	// Verify all nodes exist
	nodeMap := make(map[string]string)
	for _, node := range nodes.Nodes {
		nodeMap[node.ID] = node.Value
	}

	if nodeMap["A"] != "Value A" {
		t.Errorf("GetVerticesList() missing or wrong value for A")
	}
	if nodeMap["B"] != "Value B" {
		t.Errorf("GetVerticesList() missing or wrong value for B")
	}
	if nodeMap["C"] != "Value C" {
		t.Errorf("GetVerticesList() missing or wrong value for C")
	}
}

func TestCopyOption(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	dag := NewGenericDAG[Person]()
	dag.AddVertexByID("p1", Person{Name: "Alice", Age: 30})
	dag.AddVertexByID("p2", Person{Name: "Bob", Age: 25})
	dag.AddEdge("p1", "p2")

	// Test ShareData
	edgesShared := dag.GetEdgesWithOption(ShareData)
	nodesShared := dag.GetVerticesListWithOption(ShareData)

	// Test CopyData
	edgesCopied := dag.GetEdgesWithOption(CopyData)
	nodesCopied := dag.GetVerticesListWithOption(CopyData)

	// Both should have the same count
	if edgesShared.Count() != edgesCopied.Count() {
		t.Errorf("ShareData and CopyData should have same edge count")
	}
	if nodesShared.Count() != nodesCopied.Count() {
		t.Errorf("ShareData and CopyData should have same node count")
	}

	// Modify copied edges - should not affect original
	edgesCopied.Edges[0].SrcID = "modified"
	if edgesShared.Edges[0].SrcID == "modified" {
		t.Error("CopyData modification affected shared data")
	}
}

func TestGetEdgesByDepth(t *testing.T) {
	// Create a DAG with multiple levels
	dag := NewGenericDAG[string]()
	dag.AddVertexByID("root", "Root")
	dag.AddVertexByID("child1", "Child 1")
	dag.AddVertexByID("child2", "Child 2")
	dag.AddVertexByID("grandchild1", "Grandchild 1")
	dag.AddEdge("root", "child1")
	dag.AddEdge("root", "child2")
	dag.AddEdge("child1", "grandchild1")

	tests := []struct {
		name        string
		minDepth    int
		maxDepth    int
		wantCount   int
		description string
	}{
		{
			name:        "only root level edges",
			minDepth:    0,
			maxDepth:    0,
			wantCount:   0, // no edge's target is at depth 0
			description: "root level should have no edges",
		},
		{
			name:        "edges at depth 1",
			minDepth:    1,
			maxDepth:    1,
			wantCount:   2, // root->child1, root->child2
			description: "depth 1 edges (edges whose target is at depth 1)",
		},
		{
			name:        "edges at depth 2",
			minDepth:    2,
			maxDepth:    2,
			wantCount:   1, // child1->grandchild1
			description: "depth 2 edges",
		},
		{
			name:        "edges at depth 1-2",
			minDepth:    1,
			maxDepth:    2,
			wantCount:   3, // root->child1, root->child2, child1->grandchild1
			description: "depth 1-2 edges",
		},
		{
			name:        "unlimited depth",
			minDepth:    0,
			maxDepth:    -1,
			wantCount:   3, // all edges
			description: "unlimited depth edges",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			edges, err := dag.GetEdgesByDepth("root", tt.minDepth, tt.maxDepth)
			if err != nil {
				t.Fatalf("GetEdgesByDepth() error = %v", err)
			}

			if count := edges.Count(); count != tt.wantCount {
				t.Errorf("GetEdgesByDepth() count = %v, want %v (%s)", count, tt.wantCount, tt.description)
			}
		})
	}
}

func TestGetEdgesByDepth_InvalidInputs(t *testing.T) {
	dag := NewGenericDAG[string]()
	dag.AddVertexByID("root", "Root")
	dag.AddVertexByID("child", "Child")
	dag.AddEdge("root", "child")

	// Test empty root ID
	_, err := dag.GetEdgesByDepth("", 0, 1)
	if err == nil {
		t.Error("GetEdgesByDepth() should return error for empty root ID")
	}

	// Test unknown root ID
	_, err = dag.GetEdgesByDepth("unknown", 0, 1)
	if err == nil {
		t.Error("GetEdgesByDepth() should return error for unknown root ID")
	}

	// Test negative minDepth
	_, err = dag.GetEdgesByDepth("root", -1, 1)
	if err == nil {
		t.Error("GetEdgesByDepth() should return error for negative minDepth")
	}
}

func TestDepthLimitedSubgraph_ComplexDAG(t *testing.T) {
	// Create a more complex DAG:
	//       root
	//      / | \
	//     a  b  c
	//    /|    |\
	//   d e    f g
	//      \  /
	//       h
	dag := NewGenericDAG[string]()
	dag.AddVertexByID("root", "root")
	dag.AddVertexByID("a", "a")
	dag.AddVertexByID("b", "b")
	dag.AddVertexByID("c", "c")
	dag.AddVertexByID("d", "d")
	dag.AddVertexByID("e", "e")
	dag.AddVertexByID("f", "f")
	dag.AddVertexByID("g", "g")
	dag.AddVertexByID("h", "h")

	dag.AddEdge("root", "a")
	dag.AddEdge("root", "b")
	dag.AddEdge("root", "c")
	dag.AddEdge("a", "d")
	dag.AddEdge("a", "e")
	dag.AddEdge("c", "f")
	dag.AddEdge("c", "g")
	dag.AddEdge("e", "h")
	dag.AddEdge("f", "h")

	// Test depth 2 - should not include h
	subgraph, rootID, err := dag.GetDescendantsGraphByDepth("root", 2)
	if err != nil {
		t.Fatalf("GetDescendantsGraphByDepth() error = %v", err)
	}

	if order := subgraph.GetOrder(); order != 8 {
		t.Errorf("GetDescendantsGraphByDepth() order = %v, want %v", order, 8)
	}

	// Verify h is not in subgraph
	_, err = subgraph.GetVertex("h")
	if err == nil {
		t.Error("GetDescendantsGraphByDepth() h should not be in subgraph at depth 2")
	}

	if rootID != "root" {
		t.Errorf("GetDescendantsGraphByDepth() rootID = %v, want %v", rootID, "root")
	}

	// Test depth 3 - should include h
	subgraph3, rootID3, err := dag.GetDescendantsGraphByDepth("root", 3)
	if err != nil {
		t.Fatalf("GetDescendantsGraphByDepth() error = %v", err)
	}

	if order := subgraph3.GetOrder(); order != 9 {
		t.Errorf("GetDescendantsGraphByDepth() order = %v, want %v", order, 9)
	}

	// Verify h is in subgraph
	_, err = subgraph3.GetVertex("h")
	if err != nil {
		t.Error("GetDescendantsGraphByDepth() h should be in subgraph at depth 3")
	}

	if rootID3 != "root" {
		t.Errorf("GetDescendantsGraphByDepth() rootID = %v, want %v", rootID3, "root")
	}
}

func TestTypedDAG_GetDescendantsGraphByDepth(t *testing.T) {
	dag := New[string]()
	dag.AddVertexByID("root", "Root")
	dag.AddVertexByID("child", "Child")
	dag.AddEdge("root", "child")

	subgraph, rootID, err := dag.GetDescendantsGraphByDepth("root", 1)
	if err != nil {
		t.Fatalf("GetDescendantsGraphByDepth() error = %v", err)
	}

	if order := subgraph.GetOrder(); order != 2 {
		t.Errorf("GetDescendantsGraphByDepth() order = %v, want %v", order, 2)
	}

	if rootID != "root" {
		t.Errorf("GetDescendantsGraphByDepth() rootID = %v, want %v", rootID, "root")
	}
}

func TestTypedDAG_GetEdges(t *testing.T) {
	dag := New[string]()
	dag.AddVertexByID("A", "A")
	dag.AddVertexByID("B", "B")
	dag.AddEdge("A", "B")

	edges := dag.GetEdges()

	if count := edges.Count(); count != 1 {
		t.Errorf("GetEdges() count = %v, want %v", count, 1)
	}
}

func TestTypedDAG_GetVerticesList(t *testing.T) {
	dag := New[string]()
	dag.AddVertexByID("A", "Value A")
	dag.AddVertexByID("B", "Value B")

	nodes := dag.GetVerticesList()

	if count := nodes.Count(); count != 2 {
		t.Errorf("GetVerticesList() count = %v, want %v", count, 2)
	}
}

func TestEdgeList_Copy(t *testing.T) {
	el := NewEdgeList(2)
	el.AddEdge("A", "B")
	el.AddEdge("B", "C")

	copied := el.Copy()

	if copied.Count() != el.Count() {
		t.Errorf("Copy() count = %v, want %v", copied.Count(), el.Count())
	}

	// Modify copy - should not affect original
	copied.Edges[0].SrcID = "modified"
	if el.Edges[0].SrcID == "modified" {
		t.Error("Copy() modification affected original")
	}
}

func TestNodeList_Copy(t *testing.T) {
	nl := NewNodeList[string](2)
	nl.AddNode("A", "Value A")
	nl.AddNode("B", "Value B")

	copied := nl.Copy()

	if copied.Count() != nl.Count() {
		t.Errorf("Copy() count = %v, want %v", copied.Count(), nl.Count())
	}
}