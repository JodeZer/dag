package dag

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"
)

// ============================================================================
// Phase 1: Core Function Tests - Vertex Operations
// ============================================================================

// TestGenericDAG_AddVertex tests basic vertex addition
func TestGenericDAG_AddVertex(t *testing.T) {
	dag := NewGenericDAG[string]()
	v := "value1"

	id, err := dag.AddVertex(v)
	if err != nil {
		t.Fatalf("AddVertex failed: %v", err)
	}

	if id == "" {
		t.Error("Expected non-empty ID")
	}

	if dag.GetOrder() != 1 {
		t.Errorf("GetOrder() = %d, want 1", dag.GetOrder())
	}

	if dag.GetSize() != 0 {
		t.Errorf("GetSize() = %d, want 0", dag.GetSize())
	}

	// Verify vertex can be retrieved
	retrieved, err := dag.GetVertex(id)
	if err != nil {
		t.Fatalf("GetVertex failed: %v", err)
	}

	if retrieved != v {
		t.Errorf("GetVertex() = %v, want %v", retrieved, v)
	}
}

// TestGenericDAG_AddVertexByID tests vertex addition with specified ID
func TestGenericDAG_AddVertexByID(t *testing.T) {
	dag := NewGenericDAG[string]()
	v := "value1"
	id := "v1"

	err := dag.AddVertexByID(id, v)
	if err != nil {
		t.Fatalf("AddVertexByID failed: %v", err)
	}

	if dag.GetOrder() != 1 {
		t.Errorf("GetOrder() = %d, want 1", dag.GetOrder())
	}

	// Verify vertex can be retrieved
	retrieved, err := dag.GetVertex(id)
	if err != nil {
		t.Fatalf("GetVertex failed: %v", err)
	}

	if retrieved != v {
		t.Errorf("GetVertex() = %v, want %v", retrieved, v)
	}
}

// TestGenericDAG_AddVertexDuplicate tests error when adding duplicate vertex
func TestGenericDAG_AddVertexDuplicate(t *testing.T) {
	type testStruct struct {
		Value string
	}

	dag := NewGenericDAG[testStruct]()
	v := testStruct{Value: "test"}

	_, err := dag.AddVertex(v)
	if err != nil {
		t.Fatalf("First AddVertex failed: %v", err)
	}

	// Try to add same vertex again
	_, err = dag.AddVertex(v)
	if err == nil {
		t.Error("Expected error when adding duplicate vertex")
	}

	if _, ok := err.(VertexDuplicateError); !ok {
		t.Errorf("Expected VertexDuplicateError, got %T", err)
	}
}

// TestGenericDAG_AddVertexIDDuplicate tests error when using duplicate ID
func TestGenericDAG_AddVertexIDDuplicate(t *testing.T) {
	dag := NewGenericDAG[string]()
	id := "v1"

	err := dag.AddVertexByID(id, "value1")
	if err != nil {
		t.Fatalf("First AddVertexByID failed: %v", err)
	}

	// Try to use same ID again
	err = dag.AddVertexByID(id, "value2")
	if err == nil {
		t.Error("Expected error when using duplicate ID")
	}

	if _, ok := err.(IDDuplicateError); !ok {
		t.Errorf("Expected IDDuplicateError, got %T", err)
	}
}

// TestGenericDAG_GetVertex tests vertex retrieval
func TestGenericDAG_GetVertex(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	dag := NewGenericDAG[Person]()
	p := Person{Name: "Alice", Age: 30}
	id, _ := dag.AddVertex(p)

	retrieved, err := dag.GetVertex(id)
	if err != nil {
		t.Fatalf("GetVertex failed: %v", err)
	}

	if retrieved.Name != p.Name || retrieved.Age != p.Age {
		t.Errorf("GetVertex() = %+v, want %+v", retrieved, p)
	}
}

// TestGenericDAG_GetVertexUnknown tests error when retrieving unknown vertex
func TestGenericDAG_GetVertexUnknown(t *testing.T) {
	dag := NewGenericDAG[string]()

	_, err := dag.GetVertex("unknown")
	if err == nil {
		t.Error("Expected error when getting unknown vertex")
	}

	if _, ok := err.(IDUnknownError); !ok {
		t.Errorf("Expected IDUnknownError, got %T", err)
	}
}

// TestGenericDAG_GetVertexEmptyID tests error when retrieving with empty ID
func TestGenericDAG_GetVertexEmptyID(t *testing.T) {
	dag := NewGenericDAG[string]()

	_, err := dag.GetVertex("")
	if err == nil {
		t.Error("Expected error when getting vertex with empty ID")
	}

	if _, ok := err.(IDEmptyError); !ok {
		t.Errorf("Expected IDEmptyError, got %T", err)
	}
}

// TestGenericDAG_DeleteVertex tests vertex deletion
func TestGenericDAG_DeleteVertex(t *testing.T) {
	dag := NewGenericDAG[string]()
	v1ID, _ := dag.AddVertex("v1")
	v2ID, _ := dag.AddVertex("v2")
	_ = dag.AddEdge(v1ID, v2ID)

	err := dag.DeleteVertex(v1ID)
	if err != nil {
		t.Fatalf("DeleteVertex failed: %v", err)
	}

	if dag.GetOrder() != 1 {
		t.Errorf("GetOrder() = %d, want 1", dag.GetOrder())
	}

	if dag.GetSize() != 0 {
		t.Errorf("GetSize() = %d, want 0", dag.GetSize())
	}

	// Verify v1 no longer exists
	_, err = dag.GetVertex(v1ID)
	if err == nil {
		t.Error("Expected error when getting deleted vertex")
	}
}

// TestGenericDAG_DeleteVertexUnknown tests error when deleting unknown vertex
func TestGenericDAG_DeleteVertexUnknown(t *testing.T) {
	dag := NewGenericDAG[string]()

	err := dag.DeleteVertex("unknown")
	if err == nil {
		t.Error("Expected error when deleting unknown vertex")
	}

	if _, ok := err.(IDUnknownError); !ok {
		t.Errorf("Expected IDUnknownError, got %T", err)
	}
}

// TestGenericDAG_GetVertices tests retrieving all vertices
func TestGenericDAG_GetVertices(t *testing.T) {
	dag := NewGenericDAG[string]()
	id1, _ := dag.AddVertex("value1")
	id2, _ := dag.AddVertex("value2")
	id3, _ := dag.AddVertex("value3")

	vertices := dag.GetVertices()
	if len(vertices) != 3 {
		t.Errorf("GetVertices() = %d, want 3", len(vertices))
	}

	if v, ok := vertices[id1]; !ok || v != "value1" {
		t.Errorf("Vertex %s not found or incorrect", id1)
	}
	if v, ok := vertices[id2]; !ok || v != "value2" {
		t.Errorf("Vertex %s not found or incorrect", id2)
	}
	if v, ok := vertices[id3]; !ok || v != "value3" {
		t.Errorf("Vertex %s not found or incorrect", id3)
	}
}

// ============================================================================
// Phase 1: Core Function Tests - Edge Operations
// ============================================================================

// TestGenericDAG_AddEdge tests basic edge addition
func TestGenericDAG_AddEdge(t *testing.T) {
	dag := NewGenericDAG[string]()
	srcID, _ := dag.AddVertex("src")
	dstID, _ := dag.AddVertex("dst")

	err := dag.AddEdge(srcID, dstID)
	if err != nil {
		t.Fatalf("AddEdge failed: %v", err)
	}

	if dag.GetSize() != 1 {
		t.Errorf("GetSize() = %d, want 1", dag.GetSize())
	}

	// Verify edge exists
	isEdge, err := dag.IsEdge(srcID, dstID)
	if err != nil {
		t.Fatalf("IsEdge failed: %v", err)
	}
	if !isEdge {
		t.Error("Expected edge to exist")
	}
}

// TestGenericDAG_AddEdgeUnknownVertex tests error when adding edge with unknown vertex
func TestGenericDAG_AddEdgeUnknownVertex(t *testing.T) {
	dag := NewGenericDAG[string]()
	knownID, _ := dag.AddVertex("known")
	unknownID := "unknown"

	// Test with unknown source
	err := dag.AddEdge(unknownID, knownID)
	if err == nil {
		t.Error("Expected error when adding edge with unknown source")
	}
	if _, ok := err.(IDUnknownError); !ok {
		t.Errorf("Expected IDUnknownError, got %T", err)
	}

	// Test with unknown destination
	err = dag.AddEdge(knownID, unknownID)
	if err == nil {
		t.Error("Expected error when adding edge with unknown destination")
	}
	if _, ok := err.(IDUnknownError); !ok {
		t.Errorf("Expected IDUnknownError, got %T", err)
	}
}

// TestGenericDAG_AddEdgeDuplicate tests error when adding duplicate edge
func TestGenericDAG_AddEdgeDuplicate(t *testing.T) {
	dag := NewGenericDAG[string]()
	srcID, _ := dag.AddVertex("src")
	dstID, _ := dag.AddVertex("dst")

	_ = dag.AddEdge(srcID, dstID)

	err := dag.AddEdge(srcID, dstID)
	if err == nil {
		t.Error("Expected error when adding duplicate edge")
	}
	if _, ok := err.(EdgeDuplicateError); !ok {
		t.Errorf("Expected EdgeDuplicateError, got %T", err)
	}
}

// TestGenericDAG_AddEdgeLoop tests loop detection
func TestGenericDAG_AddEdgeLoop(t *testing.T) {
	dag := NewGenericDAG[string]()
	v1ID, _ := dag.AddVertex("v1")
	v2ID, _ := dag.AddVertex("v2")
	v3ID, _ := dag.AddVertex("v3")

	_ = dag.AddEdge(v1ID, v2ID)
	_ = dag.AddEdge(v2ID, v3ID)

	// Adding v3 -> v1 should create a loop
	err := dag.AddEdge(v3ID, v1ID)
	if err == nil {
		t.Error("Expected error when adding edge that creates loop")
	}
	if _, ok := err.(EdgeLoopError); !ok {
		t.Errorf("Expected EdgeLoopError, got %T", err)
	}
}

// TestGenericDAG_AddEdgeSameSrcDst tests error when source equals destination
func TestGenericDAG_AddEdgeSameSrcDst(t *testing.T) {
	dag := NewGenericDAG[string]()
	id, _ := dag.AddVertex("v1")

	err := dag.AddEdge(id, id)
	if err == nil {
		t.Error("Expected error when source equals destination")
	}
	if _, ok := err.(SrcDstEqualError); !ok {
		t.Errorf("Expected SrcDstEqualError, got %T", err)
	}
}

// TestGenericDAG_IsEdge tests edge existence check
func TestGenericDAG_IsEdge(t *testing.T) {
	dag := NewGenericDAG[string]()
	v1ID, _ := dag.AddVertex("v1")
	v2ID, _ := dag.AddVertex("v2")
	v3ID, _ := dag.AddVertex("v3")

	_ = dag.AddEdge(v1ID, v2ID)

	// Test existing edge
	isEdge, err := dag.IsEdge(v1ID, v2ID)
	if err != nil {
		t.Fatalf("IsEdge failed: %v", err)
	}
	if !isEdge {
		t.Error("Expected edge v1->v2 to exist")
	}

	// Test non-existing edge
	isEdge, err = dag.IsEdge(v1ID, v3ID)
	if err != nil {
		t.Fatalf("IsEdge failed: %v", err)
	}
	if isEdge {
		t.Error("Expected edge v1->v3 to not exist")
	}
}

// TestGenericDAG_DeleteEdge tests edge deletion
func TestGenericDAG_DeleteEdge(t *testing.T) {
	dag := NewGenericDAG[string]()
	srcID, _ := dag.AddVertex("src")
	dstID, _ := dag.AddVertex("dst")

	_ = dag.AddEdge(srcID, dstID)

	err := dag.DeleteEdge(srcID, dstID)
	if err != nil {
		t.Fatalf("DeleteEdge failed: %v", err)
	}

	if dag.GetSize() != 0 {
		t.Errorf("GetSize() = %d, want 0", dag.GetSize())
	}

	// Verify edge no longer exists
	isEdge, _ := dag.IsEdge(srcID, dstID)
	if isEdge {
		t.Error("Expected edge to be deleted")
	}
}

// TestGenericDAG_DeleteEdgeUnknown tests error when deleting unknown edge
func TestGenericDAG_DeleteEdgeUnknown(t *testing.T) {
	dag := NewGenericDAG[string]()
	v1ID, _ := dag.AddVertex("v1")
	v2ID, _ := dag.AddVertex("v2")

	err := dag.DeleteEdge(v1ID, v2ID)
	if err == nil {
		t.Error("Expected error when deleting unknown edge")
	}
	if _, ok := err.(EdgeUnknownError); !ok {
		t.Errorf("Expected EdgeUnknownError, got %T", err)
	}
}

// ============================================================================
// Phase 2: Query Operation Tests
// ============================================================================

// TestGenericDAG_GetOrder tests vertex count
func TestGenericDAG_GetOrder(t *testing.T) {
	dag := NewGenericDAG[string]()

	if dag.GetOrder() != 0 {
		t.Errorf("GetOrder() = %d, want 0", dag.GetOrder())
	}

	_, _ = dag.AddVertex("v1")
	if dag.GetOrder() != 1 {
		t.Errorf("GetOrder() = %d, want 1", dag.GetOrder())
	}

	_, _ = dag.AddVertex("v2")
	if dag.GetOrder() != 2 {
		t.Errorf("GetOrder() = %d, want 2", dag.GetOrder())
	}
}

// TestGenericDAG_GetSize tests edge count
func TestGenericDAG_GetSize(t *testing.T) {
	dag := NewGenericDAG[string]()
	v1ID, _ := dag.AddVertex("v1")
	v2ID, _ := dag.AddVertex("v2")
	v3ID, _ := dag.AddVertex("v3")

	if dag.GetSize() != 0 {
		t.Errorf("GetSize() = %d, want 0", dag.GetSize())
	}

	_ = dag.AddEdge(v1ID, v2ID)
	if dag.GetSize() != 1 {
		t.Errorf("GetSize() = %d, want 1", dag.GetSize())
	}

	_ = dag.AddEdge(v2ID, v3ID)
	if dag.GetSize() != 2 {
		t.Errorf("GetSize() = %d, want 2", dag.GetSize())
	}
}

// TestGenericDAG_GetLeaves tests getting leaf vertices
func TestGenericDAG_GetLeaves(t *testing.T) {
	dag := NewGenericDAG[string]()
	v1ID, _ := dag.AddVertex("v1")
	v2ID, _ := dag.AddVertex("v2")
	v3ID, _ := dag.AddVertex("v3")

	_ = dag.AddEdge(v1ID, v2ID)
	_ = dag.AddEdge(v1ID, v3ID)

	leaves := dag.GetLeaves()
	if len(leaves) != 2 {
		t.Errorf("GetLeaves() = %d, want 2", len(leaves))
	}

	if _, ok := leaves[v2ID]; !ok {
		t.Error("Expected v2 to be a leaf")
	}
	if _, ok := leaves[v3ID]; !ok {
		t.Error("Expected v3 to be a leaf")
	}
	if _, ok := leaves[v1ID]; ok {
		t.Error("Expected v1 to not be a leaf")
	}
}

// TestGenericDAG_GetRoots tests getting root vertices
func TestGenericDAG_GetRoots(t *testing.T) {
	dag := NewGenericDAG[string]()
	v1ID, _ := dag.AddVertex("v1")
	v2ID, _ := dag.AddVertex("v2")
	v3ID, _ := dag.AddVertex("v3")

	_ = dag.AddEdge(v1ID, v2ID)
	_ = dag.AddEdge(v2ID, v3ID)

	roots := dag.GetRoots()
	if len(roots) != 1 {
		t.Errorf("GetRoots() = %d, want 1", len(roots))
	}

	if _, ok := roots[v1ID]; !ok {
		t.Error("Expected v1 to be a root")
	}
	if _, ok := roots[v2ID]; ok {
		t.Error("Expected v2 to not be a root")
	}
	if _, ok := roots[v3ID]; ok {
		t.Error("Expected v3 to not be a root")
	}
}

// TestGenericDAG_IsLeaf tests checking if vertex is a leaf
func TestGenericDAG_IsLeaf(t *testing.T) {
	dag := NewGenericDAG[string]()
	v1ID, _ := dag.AddVertex("v1")
	v2ID, _ := dag.AddVertex("v2")
	v3ID, _ := dag.AddVertex("v3")

	_ = dag.AddEdge(v1ID, v2ID)
	_ = dag.AddEdge(v1ID, v3ID)

	isLeaf, _ := dag.IsLeaf(v1ID)
	if isLeaf {
		t.Error("Expected v1 to not be a leaf")
	}

	isLeaf, _ = dag.IsLeaf(v2ID)
	if !isLeaf {
		t.Error("Expected v2 to be a leaf")
	}

	isLeaf, _ = dag.IsLeaf(v3ID)
	if !isLeaf {
		t.Error("Expected v3 to be a leaf")
	}
}

// TestGenericDAG_IsRoot tests checking if vertex is a root
func TestGenericDAG_IsRoot(t *testing.T) {
	dag := NewGenericDAG[string]()
	v1ID, _ := dag.AddVertex("v1")
	v2ID, _ := dag.AddVertex("v2")
	v3ID, _ := dag.AddVertex("v3")

	_ = dag.AddEdge(v1ID, v2ID)
	_ = dag.AddEdge(v2ID, v3ID)

	isRoot, _ := dag.IsRoot(v1ID)
	if !isRoot {
		t.Error("Expected v1 to be a root")
	}

	isRoot, _ = dag.IsRoot(v2ID)
	if isRoot {
		t.Error("Expected v2 to not be a root")
	}

	isRoot, _ = dag.IsRoot(v3ID)
	if isRoot {
		t.Error("Expected v3 to not be a root")
	}
}

// TestGenericDAG_GetParents tests getting parent vertices
func TestGenericDAG_GetParents(t *testing.T) {
	dag := NewGenericDAG[string]()
	v1ID, _ := dag.AddVertex("v1")
	v2ID, _ := dag.AddVertex("v2")
	v3ID, _ := dag.AddVertex("v3")

	_ = dag.AddEdge(v1ID, v3ID)
	_ = dag.AddEdge(v2ID, v3ID)

	parents, err := dag.GetParents(v3ID)
	if err != nil {
		t.Fatalf("GetParents failed: %v", err)
	}

	if len(parents) != 2 {
		t.Errorf("GetParents() = %d, want 2", len(parents))
	}

	if _, ok := parents[v1ID]; !ok {
		t.Error("Expected v1 to be a parent of v3")
	}
	if _, ok := parents[v2ID]; !ok {
		t.Error("Expected v2 to be a parent of v3")
	}
}

// TestGenericDAG_GetChildren tests getting child vertices
func TestGenericDAG_GetChildren(t *testing.T) {
	dag := NewGenericDAG[string]()
	v1ID, _ := dag.AddVertex("v1")
	v2ID, _ := dag.AddVertex("v2")
	v3ID, _ := dag.AddVertex("v3")

	_ = dag.AddEdge(v1ID, v2ID)
	_ = dag.AddEdge(v1ID, v3ID)

	children, err := dag.GetChildren(v1ID)
	if err != nil {
		t.Fatalf("GetChildren failed: %v", err)
	}

	if len(children) != 2 {
		t.Errorf("GetChildren() = %d, want 2", len(children))
	}

	if _, ok := children[v2ID]; !ok {
		t.Error("Expected v2 to be a child of v1")
	}
	if _, ok := children[v3ID]; !ok {
		t.Error("Expected v3 to be a child of v1")
	}
}

// TestGenericDAG_GetAncestors tests getting ancestor vertices
func TestGenericDAG_GetAncestors(t *testing.T) {
	dag := NewGenericDAG[string]()
	v1ID, _ := dag.AddVertex("v1")
	v2ID, _ := dag.AddVertex("v2")
	v3ID, _ := dag.AddVertex("v3")

	_ = dag.AddEdge(v1ID, v2ID)
	_ = dag.AddEdge(v2ID, v3ID)

	ancestors, err := dag.GetAncestors(v3ID)
	if err != nil {
		t.Fatalf("GetAncestors failed: %v", err)
	}

	if len(ancestors) != 2 {
		t.Errorf("GetAncestors() = %d, want 2", len(ancestors))
	}

	if _, ok := ancestors[v1ID]; !ok {
		t.Error("Expected v1 to be an ancestor of v3")
	}
	if _, ok := ancestors[v2ID]; !ok {
		t.Error("Expected v2 to be an ancestor of v3")
	}
}

// TestGenericDAG_GetDescendants tests getting descendant vertices
func TestGenericDAG_GetDescendants(t *testing.T) {
	dag := NewGenericDAG[string]()
	v1ID, _ := dag.AddVertex("v1")
	v2ID, _ := dag.AddVertex("v2")
	v3ID, _ := dag.AddVertex("v3")

	_ = dag.AddEdge(v1ID, v2ID)
	_ = dag.AddEdge(v2ID, v3ID)

	descendants, err := dag.GetDescendants(v1ID)
	if err != nil {
		t.Fatalf("GetDescendants failed: %v", err)
	}

	if len(descendants) != 2 {
		t.Errorf("GetDescendants() = %d, want 2", len(descendants))
	}

	if _, ok := descendants[v2ID]; !ok {
		t.Error("Expected v2 to be a descendant of v1")
	}
	if _, ok := descendants[v3ID]; !ok {
		t.Error("Expected v3 to be a descendant of v1")
	}
}

// TestGenericDAG_GetOrderedAncestors tests getting ordered ancestor vertices
func TestGenericDAG_GetOrderedAncestors(t *testing.T) {
	dag := NewGenericDAG[string]()
	v1ID, _ := dag.AddVertex("v1")
	v2ID, _ := dag.AddVertex("v2")
	v3ID, _ := dag.AddVertex("v3")

	_ = dag.AddEdge(v1ID, v2ID)
	_ = dag.AddEdge(v2ID, v3ID)

	ancestors, err := dag.GetOrderedAncestors(v3ID)
	if err != nil {
		t.Fatalf("GetOrderedAncestors failed: %v", err)
	}

	if len(ancestors) != 2 {
		t.Errorf("GetOrderedAncestors() = %d, want 2", len(ancestors))
	}

	// Check that v2 comes before v1 (BFS order)
	if ancestors[0] != v2ID || ancestors[1] != v1ID {
		t.Errorf("GetOrderedAncestors() = %v, want [%s, %s]", ancestors, v2ID, v1ID)
	}
}

// TestGenericDAG_GetOrderedDescendants tests getting ordered descendant vertices
func TestGenericDAG_GetOrderedDescendants(t *testing.T) {
	dag := NewGenericDAG[string]()
	v1ID, _ := dag.AddVertex("v1")
	v2ID, _ := dag.AddVertex("v2")
	v3ID, _ := dag.AddVertex("v3")

	_ = dag.AddEdge(v1ID, v2ID)
	_ = dag.AddEdge(v2ID, v3ID)

	descendants, err := dag.GetOrderedDescendants(v1ID)
	if err != nil {
		t.Fatalf("GetOrderedDescendants failed: %v", err)
	}

	if len(descendants) != 2 {
		t.Errorf("GetOrderedDescendants() = %d, want 2", len(descendants))
	}

	// Check that v2 comes before v3 (BFS order)
	if descendants[0] != v2ID || descendants[1] != v3ID {
		t.Errorf("GetOrderedDescendants() = %v, want [%s, %s]", descendants, v2ID, v3ID)
	}
}

// ============================================================================
// Phase 3: Traversal and Subgraph Tests
// ============================================================================

// TestGenericDAG_AncestorsWalker tests ancestor walker
func TestGenericDAG_AncestorsWalker(t *testing.T) {
	dag := NewGenericDAG[string]()
	v1ID, _ := dag.AddVertex("v1")
	v2ID, _ := dag.AddVertex("v2")
	v3ID, _ := dag.AddVertex("v3")
	v4ID, _ := dag.AddVertex("v4")

	_ = dag.AddEdge(v1ID, v2ID)
	_ = dag.AddEdge(v2ID, v3ID)
	_ = dag.AddEdge(v3ID, v4ID)

	ancestorsChan, _, err := dag.AncestorsWalker(v4ID)
	if err != nil {
		t.Fatalf("AncestorsWalker failed: %v", err)
	}

	var ancestors []string
	for id := range ancestorsChan {
		ancestors = append(ancestors, id)
	}

	if len(ancestors) != 3 {
		t.Errorf("AncestorsWalker returned %d ancestors, want 3", len(ancestors))
	}
}

// TestGenericDAG_DescendantsWalker tests descendant walker
func TestGenericDAG_DescendantsWalker(t *testing.T) {
	dag := NewGenericDAG[string]()
	v1ID, _ := dag.AddVertex("v1")
	v2ID, _ := dag.AddVertex("v2")
	v3ID, _ := dag.AddVertex("v3")
	v4ID, _ := dag.AddVertex("v4")

	_ = dag.AddEdge(v1ID, v2ID)
	_ = dag.AddEdge(v2ID, v3ID)
	_ = dag.AddEdge(v3ID, v4ID)

	descendantsChan, _, err := dag.DescendantsWalker(v1ID)
	if err != nil {
		t.Fatalf("DescendantsWalker failed: %v", err)
	}

	var descendants []string
	for id := range descendantsChan {
		descendants = append(descendants, id)
	}

	if len(descendants) != 3 {
		t.Errorf("DescendantsWalker returned %d descendants, want 3", len(descendants))
	}
}

// TestGenericDAG_GetDescendantsGraph tests getting descendants subgraph
func TestGenericDAG_GetDescendantsGraph(t *testing.T) {
	dag := NewGenericDAG[string]()
	v1ID, _ := dag.AddVertex("v1")
	v2ID, _ := dag.AddVertex("v2")
	v3ID, _ := dag.AddVertex("v3")
	v4ID, _ := dag.AddVertex("v4")

	_ = dag.AddEdge(v1ID, v2ID)
	_ = dag.AddEdge(v2ID, v3ID)
	_ = dag.AddEdge(v3ID, v4ID)

	subgraph, rootID, err := dag.GetDescendantsGraph(v1ID)
	if err != nil {
		t.Fatalf("GetDescendantsGraph failed: %v", err)
	}

	if subgraph.GetOrder() != 4 {
		t.Errorf("Subgraph order = %d, want 4", subgraph.GetOrder())
	}

	if subgraph.GetSize() != 3 {
		t.Errorf("Subgraph size = %d, want 3", subgraph.GetSize())
	}

	if rootID != v1ID {
		t.Errorf("Root ID = %s, want %s", rootID, v1ID)
	}

	// Verify vertex values are preserved
	v1, err := subgraph.GetVertex(v1ID)
	if err != nil || v1 != "v1" {
		t.Errorf("Vertex v1 not preserved correctly")
	}
}

// TestGenericDAG_GetAncestorsGraph tests getting ancestors subgraph
func TestGenericDAG_GetAncestorsGraph(t *testing.T) {
	dag := NewGenericDAG[string]()
	v1ID, _ := dag.AddVertex("v1")
	v2ID, _ := dag.AddVertex("v2")
	v3ID, _ := dag.AddVertex("v3")
	v4ID, _ := dag.AddVertex("v4")

	_ = dag.AddEdge(v1ID, v2ID)
	_ = dag.AddEdge(v2ID, v3ID)
	_ = dag.AddEdge(v3ID, v4ID)

	subgraph, leafID, err := dag.GetAncestorsGraph(v4ID)
	if err != nil {
		t.Fatalf("GetAncestorsGraph failed: %v", err)
	}

	if subgraph.GetOrder() != 4 {
		t.Errorf("Subgraph order = %d, want 4", subgraph.GetOrder())
	}

	if subgraph.GetSize() != 3 {
		t.Errorf("Subgraph size = %d, want 3", subgraph.GetSize())
	}

	if leafID != v4ID {
		t.Errorf("Leaf ID = %s, want %s", leafID, v4ID)
	}
}

// ============================================================================
// Phase 4: Generic Serialization Tests
// ============================================================================

// TestGenericDAG_MarshalJSON tests JSON serialization
func TestGenericDAG_MarshalJSON(t *testing.T) {
	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	dag := NewGenericDAG[Person]()
	alice := Person{Name: "Alice", Age: 30}
	bob := Person{Name: "Bob", Age: 25}

	err := dag.AddVertexByID("alice", alice)
	if err != nil {
		t.Fatalf("AddVertexByID failed: %v", err)
	}
	_ = dag.AddVertexByID("bob", bob)
	_ = dag.AddEdge("alice", "bob")

	// Use the GenericDAG MarshalJSON method
	data, err := dag.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	// Verify JSON structure
	var result map[string]json.RawMessage
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if _, ok := result["vs"]; !ok {
		t.Error("Missing 'vs' field in JSON")
	}
	if _, ok := result["es"]; !ok {
		t.Error("Missing 'es' field in JSON")
	}
}

// TestGenericDAG_MarshalJSON_ComplexType tests JSON serialization with complex types
func TestGenericDAG_MarshalJSON_ComplexType(t *testing.T) {
	type Task struct {
		Name     string  `json:"name"`
		Duration int     `json:"duration"`
		Priority float64 `json:"priority"`
	}

	dag := NewGenericDAG[Task]()
	task1 := Task{Name: "Task 1", Duration: 10, Priority: 1.5}
	task2 := Task{Name: "Task 2", Duration: 20, Priority: 2.0}

	_ = dag.AddVertexByID("t1", task1)
	_ = dag.AddVertexByID("t2", task2)
	_ = dag.AddEdge("t1", "t2")

	data, err := dag.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	// Verify unmarshaling works
	var result map[string]json.RawMessage
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if _, ok := result["vs"]; !ok {
		t.Error("Missing 'vs' field in JSON")
	}
}

// TestGenericDAG_UnmarshalJSON tests JSON deserialization
func TestGenericDAG_UnmarshalJSON(t *testing.T) {
	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	data := []byte(`{"vs":[{"i":"alice","v":{"name":"Alice","age":30}},{"i":"bob","v":{"name":"Bob","age":25}}],"es":[{"s":"alice","d":"bob"}]}`)

	dag, err := UnmarshalGenericJSON[Person](data, defaultOptions())
	if err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}

	if dag.GetOrder() != 2 {
		t.Errorf("Expected 2 vertices, got %d", dag.GetOrder())
	}

	if dag.GetSize() != 1 {
		t.Errorf("Expected 1 edge, got %d", dag.GetSize())
	}

	alice, err := dag.GetVertex("alice")
	if err != nil {
		t.Fatalf("GetVertex failed: %v", err)
	}

	if alice.Name != "Alice" || alice.Age != 30 {
		t.Errorf("Alice data mismatch: got %+v", alice)
	}
}

// TestGenericDAG_UnmarshalJSON_ComplexType tests JSON deserialization with complex types
func TestGenericDAG_UnmarshalJSON_ComplexType(t *testing.T) {
	type Task struct {
		Name     string  `json:"name"`
		Duration int     `json:"duration"`
		Priority float64 `json:"priority"`
	}

	data := []byte(`{"vs":[{"i":"t1","v":{"name":"Task 1","duration":10,"priority":1.5}},{"i":"t2","v":{"name":"Task 2","duration":20,"priority":2.0}}],"es":[{"s":"t1","d":"t2"}]}`)

	dag, err := UnmarshalGenericJSON[Task](data, defaultOptions())
	if err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}

	task1, err := dag.GetVertex("t1")
	if err != nil {
		t.Fatalf("GetVertex failed: %v", err)
	}

	if task1.Name != "Task 1" || task1.Duration != 10 || task1.Priority != 1.5 {
		t.Errorf("Task1 data mismatch: got %+v", task1)
	}
}

// TestGenericDAG_MarshalUnmarshalRoundtrip tests serialization roundtrip
func TestGenericDAG_MarshalUnmarshalRoundtrip(t *testing.T) {
	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	original := NewGenericDAG[Person]()
	_ = original.AddVertexByID("alice", Person{Name: "Alice", Age: 30})
	_ = original.AddVertexByID("bob", Person{Name: "Bob", Age: 25})
	_ = original.AddVertexByID("charlie", Person{Name: "Charlie", Age: 35})
	_ = original.AddEdge("alice", "bob")
	_ = original.AddEdge("bob", "charlie")

	// Serialize
	data, err := original.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	// Deserialize
	restored, err := UnmarshalGenericJSON[Person](data, defaultOptions())
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

// ============================================================================
// Phase 5: Boundary Case and Error Handling Tests
// ============================================================================

// TestGenericDAG_EmptyGraph tests operations on empty graph
func TestGenericDAG_EmptyGraph(t *testing.T) {
	dag := NewGenericDAG[string]()

	if dag.GetOrder() != 0 {
		t.Errorf("GetOrder() = %d, want 0", dag.GetOrder())
	}

	if dag.GetSize() != 0 {
		t.Errorf("GetSize() = %d, want 0", dag.GetSize())
	}

	if len(dag.GetVertices()) != 0 {
		t.Errorf("GetVertices() = %d, want 0", len(dag.GetVertices()))
	}

	if len(dag.GetLeaves()) != 0 {
		t.Errorf("GetLeaves() = %d, want 0", len(dag.GetLeaves()))
	}

	if len(dag.GetRoots()) != 0 {
		t.Errorf("GetRoots() = %d, want 0", len(dag.GetRoots()))
	}
}

// TestGenericDAG_SingleVertex tests graph with single vertex
func TestGenericDAG_SingleVertex(t *testing.T) {
	dag := NewGenericDAG[string]()
	id, _ := dag.AddVertex("value")

	if dag.GetOrder() != 1 {
		t.Errorf("GetOrder() = %d, want 1", dag.GetOrder())
	}

	if dag.GetSize() != 0 {
		t.Errorf("GetSize() = %d, want 0", dag.GetSize())
	}

	// Single vertex should be both root and leaf
	roots := dag.GetRoots()
	leaves := dag.GetLeaves()

	if len(roots) != 1 {
		t.Errorf("GetRoots() = %d, want 1", len(roots))
	}

	if len(leaves) != 1 {
		t.Errorf("GetLeaves() = %d, want 1", len(leaves))
	}

	isRoot, _ := dag.IsRoot(id)
	if !isRoot {
		t.Error("Expected single vertex to be a root")
	}

	isLeaf, _ := dag.IsLeaf(id)
	if !isLeaf {
		t.Error("Expected single vertex to be a leaf")
	}
}

// TestGenericDAG_LargeGraph tests large graph performance
func TestGenericDAG_LargeGraph(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large graph test in short mode")
	}

	dag := NewGenericDAG[string]()
	size := 1000

	// Create a linear DAG using explicit IDs to avoid hash collisions
	var prevID string
	for i := 0; i < size; i++ {
		id := fmt.Sprintf("node_%08d", i) // Pad with zeros for consistent sorting
		err := dag.AddVertexByID(id, fmt.Sprintf("value_%d", i))
		if err != nil {
			t.Fatalf("AddVertexByID failed: %v", err)
		}

		if prevID != "" {
			if err := dag.AddEdge(prevID, id); err != nil {
				t.Fatalf("AddEdge failed: %v", err)
			}
		}
		prevID = id
	}

	if dag.GetOrder() != size {
		t.Errorf("GetOrder() = %d, want %d", dag.GetOrder(), size)
	}

	if dag.GetSize() != size-1 {
		t.Errorf("GetSize() = %d, want %d", dag.GetSize(), size-1)
	}

	// Test GetDescendants on root
	rootID := fmt.Sprintf("node_%08d", 0)
	descendants, err := dag.GetDescendants(rootID)
	if err != nil {
		t.Fatalf("GetDescendants failed: %v", err)
	}

	if len(descendants) != size-1 {
		t.Errorf("GetDescendants() = %d, want %d", len(descendants), size-1)
	}
}

// TestGenericDAG_DiamondPattern tests diamond DAG pattern
func TestGenericDAG_DiamondPattern(t *testing.T) {
	dag := NewGenericDAG[string]()
	aID, _ := dag.AddVertex("A")
	bID, _ := dag.AddVertex("B")
	cID, _ := dag.AddVertex("C")
	dID, _ := dag.AddVertex("D")

	_ = dag.AddEdge(aID, bID)
	_ = dag.AddEdge(aID, cID)
	_ = dag.AddEdge(bID, dID)
	_ = dag.AddEdge(cID, dID)

	// Verify structure
	if dag.GetOrder() != 4 {
		t.Errorf("GetOrder() = %d, want 4", dag.GetOrder())
	}

	if dag.GetSize() != 4 {
		t.Errorf("GetSize() = %d, want 4", dag.GetSize())
	}

	// A should have 2 descendants
	descendants, _ := dag.GetDescendants(aID)
	if len(descendants) != 3 {
		t.Errorf("A should have 3 descendants, got %d", len(descendants))
	}

	// D should have 2 ancestors
	ancestors, _ := dag.GetAncestors(dID)
	if len(ancestors) != 3 {
		t.Errorf("D should have 3 ancestors, got %d", len(ancestors))
	}
}

// TestGenericDAG_ComplexDAG tests complex DAG structure
func TestGenericDAG_ComplexDAG(t *testing.T) {
	dag := NewGenericDAG[string]()
	v1ID, _ := dag.AddVertex("v1")
	v2ID, _ := dag.AddVertex("v2")
	v3ID, _ := dag.AddVertex("v3")
	v4ID, _ := dag.AddVertex("v4")
	v5ID, _ := dag.AddVertex("v5")

	// Create a complex structure
	_ = dag.AddEdge(v1ID, v2ID)
	_ = dag.AddEdge(v1ID, v3ID)
	_ = dag.AddEdge(v2ID, v4ID)
	_ = dag.AddEdge(v3ID, v4ID)
	_ = dag.AddEdge(v4ID, v5ID)

	// Verify counts
	if dag.GetOrder() != 5 {
		t.Errorf("GetOrder() = %d, want 5", dag.GetOrder())
	}

	if dag.GetSize() != 5 {
		t.Errorf("GetSize() = %d, want 5", dag.GetSize())
	}

	// Verify roots and leaves
	if len(dag.GetRoots()) != 1 {
		t.Errorf("Expected 1 root, got %d", len(dag.GetRoots()))
	}

	if len(dag.GetLeaves()) != 1 {
		t.Errorf("Expected 1 leaf, got %d", len(dag.GetLeaves()))
	}
}

// ============================================================================
// Phase 6: Concurrency and Memory Tests
// ============================================================================

// TestGenericDAG_ConcurrentReadWrite tests concurrent read and write
func TestGenericDAG_ConcurrentReadWrite(t *testing.T) {
	dag := NewGenericDAG[string]()
	v1ID, _ := dag.AddVertex("v1")
	v2ID, _ := dag.AddVertex("v2")
	_ = dag.AddEdge(v1ID, v2ID)

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent readers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				_ = dag.GetOrder()
				_ = dag.GetSize()
				_ = dag.GetVertices()
				_, _ = dag.GetVertex(v1ID)
			}
		}()
	}

	// Concurrent writers
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			id, _ := dag.AddVertex("")
			_ = dag.AddEdge(v1ID, id)
			_ = dag.DeleteEdge(v1ID, id)
			_ = dag.DeleteVertex(id)
		}(i)
	}

	wg.Wait()
}

// TestGenericDAG_ConcurrentAddVertex tests concurrent vertex addition
func TestGenericDAG_ConcurrentAddVertex(t *testing.T) {
	dag := NewGenericDAG[string]()
	var wg sync.WaitGroup
	count := 100

	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			_, _ = dag.AddVertex(fmt.Sprintf("vertex_%d", n))
		}(i)
	}

	wg.Wait()

	if dag.GetOrder() != count {
		t.Errorf("GetOrder() = %d, want %d", dag.GetOrder(), count)
	}
}

// TestGenericDAG_ConcurrentAddEdge tests concurrent edge addition
func TestGenericDAG_ConcurrentAddEdge(t *testing.T) {
	dag := NewGenericDAG[string]()
	var vertexIDs []string
	count := 50

	// Add vertices first
	for i := 0; i < count; i++ {
		id, _ := dag.AddVertex(fmt.Sprintf("v%d", i))
		vertexIDs = append(vertexIDs, id)
	}

	var wg sync.WaitGroup

	// Add edges concurrently
	for i := 0; i < count-1; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			_ = dag.AddEdge(vertexIDs[n], vertexIDs[n+1])
		}(i)
	}

	wg.Wait()

	expectedEdges := count - 1
	if dag.GetSize() != expectedEdges {
		t.Errorf("GetSize() = %d, want %d", dag.GetSize(), expectedEdges)
	}
}

// TestGenericDAG_CacheFlush tests cache flushing
func TestGenericDAG_CacheFlush(t *testing.T) {
	dag := NewGenericDAG[string]()
	v1ID, _ := dag.AddVertex("v1")
	v2ID, _ := dag.AddVertex("v2")
	v3ID, _ := dag.AddVertex("v3")

	_ = dag.AddEdge(v1ID, v2ID)
	_ = dag.AddEdge(v2ID, v3ID)

	// Populate caches
	_, _ = dag.GetDescendants(v1ID)
	_, _ = dag.GetAncestors(v3ID)

	// Flush caches
	dag.FlushCaches()

	// Verify operations still work after flush
	descendants, err := dag.GetDescendants(v1ID)
	if err != nil {
		t.Fatalf("GetDescendants failed after flush: %v", err)
	}

	if len(descendants) != 2 {
		t.Errorf("GetDescendants() = %d, want 2", len(descendants))
	}
}

// TestGenericDAG_MemoryNoLeaks tests for memory leaks
func TestGenericDAG_MemoryNoLeaks(t *testing.T) {
	// Create a large graph and then delete everything
	dag := NewGenericDAG[string]()
	var ids []string

	// Add vertices
	for i := 0; i < 100; i++ {
		id, _ := dag.AddVertex("")
		ids = append(ids, id)
	}

	// Add edges
	for i := 0; i < len(ids)-1; i++ {
		_ = dag.AddEdge(ids[i], ids[i+1])
	}

	// Delete all vertices
	for _, id := range ids {
		_ = dag.DeleteVertex(id)
	}

	// Verify empty
	if dag.GetOrder() != 0 {
		t.Errorf("GetOrder() = %d, want 0 after deleting all vertices", dag.GetOrder())
	}

	if dag.GetSize() != 0 {
		t.Errorf("GetSize() = %d, want 0 after deleting all vertices", dag.GetSize())
	}
}

// ============================================================================
// Phase 7: Type Safety Tests
// ============================================================================

// TestGenericDAG_WithStringType tests with string type
func TestGenericDAG_WithStringType(t *testing.T) {
	dag := NewGenericDAG[string]()
	id1, _ := dag.AddVertex("value1")
	id2, _ := dag.AddVertex("value2")
	_ = dag.AddEdge(id1, id2)

	v1, err := dag.GetVertex(id1)
	if err != nil {
		t.Fatalf("GetVertex failed: %v", err)
	}

	if v1 != "value1" {
		t.Errorf("Expected 'value1', got '%s'", v1)
	}

	// Type checking is enforced at compile time
	var s string = v1 // This should compile
	_ = s
}

// TestGenericDAG_WithIntType tests with int type
func TestGenericDAG_WithIntType(t *testing.T) {
	dag := NewGenericDAG[int]()
	id1, _ := dag.AddVertex(42)
	id2, _ := dag.AddVertex(100)
	_ = dag.AddEdge(id1, id2)

	v1, err := dag.GetVertex(id1)
	if err != nil {
		t.Fatalf("GetVertex failed: %v", err)
	}

	if v1 != 42 {
		t.Errorf("Expected 42, got %d", v1)
	}

	// Type checking is enforced at compile time
	var n int = v1 // This should compile
	_ = n
}

// TestGenericDAG_WithStructType tests with struct type
func TestGenericDAG_WithStructType(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	dag := NewGenericDAG[Person]()
	alice := Person{Name: "Alice", Age: 30}
	id1, _ := dag.AddVertex(alice)

	v1, err := dag.GetVertex(id1)
	if err != nil {
		t.Fatalf("GetVertex failed: %v", err)
	}

	if v1.Name != "Alice" || v1.Age != 30 {
		t.Errorf("Expected {Name:Alice, Age:30}, got %+v", v1)
	}

	// Type checking is enforced at compile time
	var p Person = v1 // This should compile
	_ = p
}

// TestGenericDAG_WithPointerType tests with pointer type
func TestGenericDAG_WithPointerType(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	dag := NewGenericDAG[*Person]()
	alice := &Person{Name: "Alice", Age: 30}
	id1, _ := dag.AddVertex(alice)

	v1, err := dag.GetVertex(id1)
	if err != nil {
		t.Fatalf("GetVertex failed: %v", err)
	}

	if v1.Name != "Alice" || v1.Age != 30 {
		t.Errorf("Expected {Name:Alice, Age:30}, got %+v", v1)
	}

	// Pointer should be the same
	if v1 != alice {
		t.Error("Expected same pointer instance")
	}

	// Type checking is enforced at compile time
	var p *Person = v1 // This should compile
	_ = p
}

// ============================================================================
// Phase 8: Copy and Optimization Tests
// ============================================================================

// TestGenericDAG_Copy tests graph copying
func TestGenericDAG_Copy(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	original := NewGenericDAG[Person]()
	_ = original.AddVertexByID("alice", Person{Name: "Alice", Age: 30})
	_ = original.AddVertexByID("bob", Person{Name: "Bob", Age: 25})
	_ = original.AddEdge("alice", "bob")

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

	// Verify vertices are preserved
	alice, err := copy.GetVertex("alice")
	if err != nil {
		t.Fatalf("GetVertex failed: %v", err)
	}

	if alice.Name != "Alice" || alice.Age != 30 {
		t.Errorf("Alice data mismatch: got %+v", alice)
	}

	// Verify copy is independent
	_ = copy.AddVertexByID("charlie", Person{Name: "Charlie", Age: 35})
	if original.GetOrder() == copy.GetOrder() {
		t.Error("Copy should be independent of original")
	}
}

// TestGenericDAG_ReduceTransitively tests transitive reduction
func TestGenericDAG_ReduceTransitively(t *testing.T) {
	dag := NewGenericDAG[string]()
	_ = dag.AddVertexByID("account", "AccountCreate")
	_ = dag.AddVertexByID("project", "ProjectCreate")
	_ = dag.AddVertexByID("network", "NetworkCreate")
	_ = dag.AddVertexByID("contact", "ContactCreate")
	_ = dag.AddVertexByID("auth", "AuthCreate")
	_ = dag.AddVertexByID("mail", "MailSend")

	_ = dag.AddEdge("account", "project")
	_ = dag.AddEdge("account", "network")
	_ = dag.AddEdge("account", "contact")
	_ = dag.AddEdge("account", "auth")
	_ = dag.AddEdge("account", "mail")

	_ = dag.AddEdge("project", "mail")
	_ = dag.AddEdge("network", "mail")
	_ = dag.AddEdge("contact", "mail")
	_ = dag.AddEdge("auth", "mail")

	originalSize := dag.GetSize()

	// Reduce transitively
	dag.ReduceTransitively()

	reducedSize := dag.GetSize()

	// Should have removed the direct edges from account to mail
	// since there are indirect paths through the other vertices
	if reducedSize != originalSize-1 {
		t.Errorf("After reduction, size should be %d, got %d", originalSize-1, reducedSize)
	}

	// Verify the edge account->mail was removed
	isEdge, _ := dag.IsEdge("account", "mail")
	if isEdge {
		t.Error("Expected edge account->mail to be removed")
	}

	// Verify other edges still exist
	isEdge, _ = dag.IsEdge("account", "project")
	if !isEdge {
		t.Error("Expected edge account->project to exist")
	}

	isEdge, _ = dag.IsEdge("project", "mail")
	if !isEdge {
		t.Error("Expected edge project->mail to exist")
	}
}

// TestGenericDAG_Options tests custom options
func TestGenericDAG_Options(t *testing.T) {
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

	dag := NewGenericDAG[Person]()
	dag.Options(options)

	_, _ = dag.AddVertex(Person{Name: "Alice"})
	_, _ = dag.AddVertex(Person{Name: "Bob"})

	if dag.GetOrder() != 2 {
		t.Errorf("GetOrder() = %d, want 2", dag.GetOrder())
	}
}

// TestGenericDAG_CustomHashFunc tests custom hash function
func TestGenericDAG_CustomHashFunc(t *testing.T) {
	type Person struct {
		Name string
	}

	options := Options{
		VertexHashFunc: func(v interface{}) interface{} {
			if p, ok := v.(Person); ok {
				return p.Name // Hash by name only
			}
			return v
		},
	}

	dag := NewGenericDAG[Person]()
	dag.Options(options)

	// Add person with name "Alice"
	_, _ = dag.AddVertex(Person{Name: "Alice"})

	// Try to add another person with same name - should fail
	// because they hash to the same value
	_, err := dag.AddVertex(Person{Name: "Alice"})
	if err == nil {
		t.Error("Expected error when adding person with same name")
	}

	if _, ok := err.(VertexDuplicateError); !ok {
		t.Errorf("Expected VertexDuplicateError, got %T", err)
	}
}

// TestGenericDAG_String tests String method
func TestGenericDAG_String(t *testing.T) {
	dag := NewGenericDAG[string]()
	v1ID, _ := dag.AddVertex("v1")
	v2ID, _ := dag.AddVertex("v2")
	v3ID, _ := dag.AddVertex("v3")

	_ = dag.AddEdge(v1ID, v2ID)
	_ = dag.AddEdge(v2ID, v3ID)

	s := dag.String()
	if len(s) == 0 {
		t.Error("String() returned empty string")
	}

	// Verify it contains expected information
	// Note: We don't check exact format since hash values may vary
	if len(s) < 10 {
		t.Error("String() returned too short string")
	}
}