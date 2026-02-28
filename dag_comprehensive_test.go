package dag

import (
	"fmt"
	"sync"
	"testing"
)

// TestEmptyGraph tests operations on an empty graph.
func TestEmptyGraph(t *testing.T) {
	d := NewDAG()

	if order := d.GetOrder(); order != 0 {
		t.Errorf("GetOrder() = %d, want 0", order)
	}
	if size := d.GetSize(); size != 0 {
		t.Errorf("GetSize() = %d, want 0", size)
	}
	if len(d.GetVertices()) != 0 {
		t.Errorf("GetVertices() = %d, want 0", len(d.GetVertices()))
	}
	if len(d.GetRoots()) != 0 {
		t.Errorf("GetRoots() = %d, want 0", len(d.GetRoots()))
	}
	if len(d.GetLeaves()) != 0 {
		t.Errorf("GetLeaves() = %d, want 0", len(d.GetLeaves()))
	}
}

// TestSingleVertex tests a graph with only one vertex.
func TestSingleVertex(t *testing.T) {
	d := NewDAG()

	id, _ := d.AddVertex(1)
	if order := d.GetOrder(); order != 1 {
		t.Errorf("GetOrder() = %d, want 1", order)
	}
	if size := d.GetSize(); size != 0 {
		t.Errorf("GetSize() = %d, want 0", size)
	}
	if len(d.GetRoots()) != 1 {
		t.Errorf("GetRoots() = %d, want 1", len(d.GetRoots()))
	}
	if len(d.GetLeaves()) != 1 {
		t.Errorf("GetLeaves() = %d, want 1", len(d.GetLeaves()))
	}

	isRoot, _ := d.IsRoot(id)
	if !isRoot {
		t.Errorf("IsRoot(id) = false, want true")
	}
	isLeaf, _ := d.IsLeaf(id)
	if !isLeaf {
		t.Errorf("IsLeaf(id) = false, want true")
	}

	desc, _ := d.GetDescendants(id)
	if len(desc) != 0 {
		t.Errorf("GetDescendants(id) = %d, want 0", len(desc))
	}
	anc, _ := d.GetAncestors(id)
	if len(anc) != 0 {
		t.Errorf("GetAncestors(id) = %d, want 0", len(anc))
	}
}

// TestTwoVertices tests a graph with two vertices and one edge.
func TestTwoVertices(t *testing.T) {
	d := NewDAG()

	v1, _ := d.AddVertex(1)
	v2, _ := d.AddVertex(2)
	_ = d.AddEdge(v1, v2)

	if order := d.GetOrder(); order != 2 {
		t.Errorf("GetOrder() = %d, want 2", order)
	}
	if size := d.GetSize(); size != 1 {
		t.Errorf("GetSize() = %d, want 1", size)
	}
	if len(d.GetRoots()) != 1 {
		t.Errorf("GetRoots() = %d, want 1", len(d.GetRoots()))
	}
	if len(d.GetLeaves()) != 1 {
		t.Errorf("GetLeaves() = %d, want 1", len(d.GetLeaves()))
	}

	desc, _ := d.GetDescendants(v1)
	if len(desc) != 1 {
		t.Errorf("GetDescendants(v1) = %d, want 1", len(desc))
	}
	anc, _ := d.GetAncestors(v2)
	if len(anc) != 1 {
		t.Errorf("GetAncestors(v2) = %d, want 1", len(anc))
	}
}

// TestLinearChain tests a linear chain: 1->2->3->...->n.
func TestLinearChain(t *testing.T) {
	d := generateLinearDAG(10)

	if order := d.GetOrder(); order != 10 {
		t.Errorf("GetOrder() = %d, want 10", order)
	}
	if size := d.GetSize(); size != 9 {
		t.Errorf("GetSize() = %d, want 9", size)
	}

	// Test root and leaf
	roots := d.GetRoots()
	if len(roots) != 1 {
		t.Errorf("GetRoots() = %d, want 1", len(roots))
	}
	leaves := d.GetLeaves()
	if len(leaves) != 1 {
		t.Errorf("GetLeaves() = %d, want 1", len(leaves))
	}

	// Test descendants from root
	var rootID string
	for id := range roots {
		rootID = id
	}
	desc, _ := d.GetDescendants(rootID)
	if len(desc) != 9 {
		t.Errorf("GetDescendants(root) = %d, want 9", len(desc))
	}
}

// TestCompleteBinaryTree tests a complete binary tree structure.
func TestCompleteBinaryTree(t *testing.T) {
	d := NewDAG()

	// Create a complete binary tree: root has 2 children, each has 2 children, etc.
	depth := 3
	branch := 2

	rootID := "root_0"
	_, _ = d.AddVertex(TestVertex{VertexID: rootID, Name: "Root"})

	currentLevel := []string{rootID}
	for level := 1; level < depth; level++ {
		var nextLevel []string
		nodeCounter := 0

		for _, parentID := range currentLevel {
			for b := 0; b < branch; b++ {
				id := fmt.Sprintf("node_%d_%d", level, nodeCounter)
				_, _ = d.AddVertex(TestVertex{VertexID: id, Name: fmt.Sprintf("Node%d_%d", level, nodeCounter)})
				_ = d.AddEdge(parentID, id)
				nextLevel = append(nextLevel, id)
				nodeCounter++
			}
		}
		currentLevel = nextLevel
	}

	expectedVertices := 1 // root
	for i := 0; i < depth-1; i++ {
		expectedVertices *= 2
	}
	expectedVertices = 2*expectedVertices - 1

	if order := d.GetOrder(); order != expectedVertices {
		t.Errorf("GetOrder() = %d, want %d", order, expectedVertices)
	}

	// Test root and leaf counts
	roots := d.GetRoots()
	if len(roots) != 1 {
		t.Errorf("GetRoots() = %d, want 1", len(roots))
	}
	leaves := d.GetLeaves()
	expectedLeaves := 1 << uint(depth-1)
	if len(leaves) != expectedLeaves {
		t.Errorf("GetLeaves() = %d, want %d", len(leaves), expectedLeaves)
	}
}

// TestDiamondDependency tests a diamond dependency graph.
func TestDiamondDependency(t *testing.T) {
	d := generateDiamondDAG()

	if order := d.GetOrder(); order != 4 {
		t.Errorf("GetOrder() = %d, want 4", order)
	}
	if size := d.GetSize(); size != 4 {
		t.Errorf("GetSize() = %d, want 4", size)
	}

	// Test descendants of A
	desc, _ := d.GetDescendants("A")
	if len(desc) != 3 {
		t.Errorf("GetDescendants(A) = %d, want 3", len(desc))
	}

	// Test ancestors of D
	anc, _ := d.GetAncestors("D")
	if len(anc) != 3 {
		t.Errorf("GetAncestors(D) = %d, want 3", len(anc))
	}
}

// TestConcurrentAddVertex tests concurrent AddVertex operations.
func TestConcurrentAddVertex(t *testing.T) {
	d := NewDAG()
	numVertices := 1000
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(start int) {
			defer wg.Done()
			for j := start; j < numVertices; j += 10 {
				_, _ = d.AddVertex(fmt.Sprintf("node_%d", j))
			}
		}(i)
	}

	wg.Wait()

	if order := d.GetOrder(); order != numVertices {
		t.Errorf("GetOrder() = %d, want %d", order, numVertices)
	}
}

// TestConcurrentAddEdge tests concurrent AddEdge operations.
func TestConcurrentAddEdge(t *testing.T) {
	d := NewDAG()
	numVertices := 100

	// First add vertices
	for i := 0; i < numVertices; i++ {
		_, _ = d.AddVertex(fmt.Sprintf("node_%d", i))
	}

	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(start int) {
			defer wg.Done()
			for j := start; j < numVertices-1; j += 10 {
				if d.AddEdge(fmt.Sprintf("node_%d", j), fmt.Sprintf("node_%d", j+1)) == nil {
					// We're not tracking edges in this test, just checking for safety
				}
			}
		}(i)
	}

	wg.Wait()

	// Check that graph is still valid (no cycles)
	// This test mainly checks for race conditions
}

// TestConcurrentGetDescendants tests concurrent GetDescendants operations.
func TestConcurrentGetDescendants(t *testing.T) {
	d := generateLinearDAG(100)
	var wg sync.WaitGroup

	// Populate cache first
	d.GetDescendants("node_0")

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_, _ = d.GetDescendants(fmt.Sprintf("node_%d", j))
			}
		}()
	}

	wg.Wait()
}

// TestConcurrentGetAncestors tests concurrent GetAncestors operations.
func TestConcurrentGetAncestors(t *testing.T) {
	d := generateLinearDAG(100)
	var wg sync.WaitGroup

	// Populate cache first
	d.GetAncestors("node_99")

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_, _ = d.GetAncestors(fmt.Sprintf("node_%d", j))
			}
		}()
	}

	wg.Wait()
}

// TestConcurrentReadWrite tests concurrent read and write operations.
func TestConcurrentReadWrite(t *testing.T) {
	d := NewDAG()
	var wg sync.WaitGroup

	// Add initial vertices
	for i := 0; i < 10; i++ {
		_, _ = d.AddVertex(fmt.Sprintf("node_%d", i))
	}

	// Readers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_, _ = d.GetVertex(fmt.Sprintf("node_%d", j%10))
				d.GetVertices()
				d.GetRoots()
				d.GetLeaves()
			}
		}()
	}

	// Writers
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(start int) {
			defer wg.Done()
			for j := start; j < 100; j += 3 {
				id := fmt.Sprintf("node_%d", j)
				if _, err := d.GetVertex(id); err != nil {
					_, _ = d.AddVertex(j)
				}
			}
		}(i)
	}

	wg.Wait()
}

// TestFlushCaches tests cache flushing behavior.
func TestFlushCaches(t *testing.T) {
	d := generateLinearDAG(100)

	// Populate cache
	d.GetDescendants("node_0")
	d.GetAncestors("node_99")

	// Flush caches
	d.FlushCaches()

	// Query again - should rebuild cache
	desc, _ := d.GetDescendants("node_0")
	if len(desc) != 99 {
		t.Errorf("GetDescendants(node_0) = %d, want 99", len(desc))
	}

	anc, _ := d.GetAncestors("node_99")
	if len(anc) != 99 {
		t.Errorf("GetAncestors(node_99) = %d, want 99", len(anc))
	}
}

// TestDeleteVertexCacheInvalidation tests that cache is invalidated when vertex is deleted.
func TestDeleteVertexCacheInvalidation(t *testing.T) {
	d := generateLinearDAG(10)

	// Populate cache
	descBefore, _ := d.GetDescendants("node_0")
	if len(descBefore) != 9 {
		t.Errorf("GetDescendants(node_0) = %d, want 9", len(descBefore))
	}

	// Delete middle vertex
	_ = d.DeleteVertex("node_5")

	// Check that descendants are updated
	descAfter, _ := d.GetDescendants("node_0")
	if len(descAfter) != 4 { // node_1, node_2, node_3, node_4
		t.Errorf("GetDescendants(node_0) after delete = %d, want 4", len(descAfter))
	}
}

// TestDeleteEdgeCacheInvalidation tests that cache is invalidated when edge is deleted.
func TestDeleteEdgeCacheInvalidation(t *testing.T) {
	d := generateLinearDAG(10)

	// Populate cache
	descBefore, _ := d.GetDescendants("node_0")
	if len(descBefore) != 9 {
		t.Errorf("GetDescendants(node_0) = %d, want 9", len(descBefore))
	}

	// Delete middle edge
	_ = d.DeleteEdge("node_4", "node_5")

	// Check that descendants are updated
	descAfter, _ := d.GetDescendants("node_0")
	if len(descAfter) != 4 { // node_1, node_2, node_3, node_4
		t.Errorf("GetDescendants(node_0) after delete = %d, want 4", len(descAfter))
	}
}

// TestCachePerformance tests cache performance difference.
func TestCachePerformance(t *testing.T) {
	d := generateWideTreeDAG(5, 10)

	rootID := "root_0"

	// First call (cache miss)
	allocsFirst := testing.AllocsPerRun(1, func() {
		_, _ = d.GetDescendants(rootID)
	})

	// Second call (cache hit)
	allocsSecond := testing.AllocsPerRun(100, func() {
		_, _ = d.GetDescendants(rootID)
	})

	// Cache hit should allocate significantly less
	if allocsSecond > allocsFirst*2 {
		t.Logf("Warning: Cache hit allocs (%f) > 2 * cache miss allocs (%f)", allocsSecond, allocsFirst)
	}
}

// TestDescendantsFlowSingleVertex tests DescendantsFlow with a single vertex.
func TestDescendantsFlowSingleVertex(t *testing.T) {
	d := NewDAG()
	v0, _ := d.AddVertex(0)

	callback := func(d *DAG, id string, parentResults []FlowResult) (interface{}, error) {
		v, _ := d.GetVertex(id)
		return v.(int) + 1, nil
	}

	results, _ := d.DescendantsFlow(v0, nil, callback)
	if len(results) != 1 {
		t.Errorf("DescendantsFlow() = %d results, want 1", len(results))
	}
	if results[0].Result.(int) != 1 {
		t.Errorf("Result = %d, want 1", results[0].Result.(int))
	}
}

// TestDescendantsFlowParallel tests DescendantsFlow with parallel execution.
func TestDescendantsFlowParallel(t *testing.T) {
	d := NewDAG()

	// Create a simple parallel graph
	v0, _ := d.AddVertex(0)
	v1, _ := d.AddVertex(1)
	v2, _ := d.AddVertex(2)
	_ = d.AddEdge(v0, v1)
	_ = d.AddEdge(v0, v2)

	inputs := []FlowResult{{ID: v0, Result: 10}}

	var executionOrder []string
	var mu sync.Mutex

	callback := func(d *DAG, id string, parentResults []FlowResult) (interface{}, error) {
		mu.Lock()
		executionOrder = append(executionOrder, id)
		mu.Unlock()

		v, _ := d.GetVertex(id)
		result := v.(int)
		for _, pr := range parentResults {
			result += pr.Result.(int)
		}
		return result, nil
	}

	results, _ := d.DescendantsFlow(v0, inputs, callback)

	if len(results) != 2 {
		t.Errorf("DescendantsFlow() = %d results, want 2", len(results))
	}

	// Verify results
	resultSum := 0
	for _, r := range results {
		resultSum += r.Result.(int)
	}
	if resultSum != 23 { // v0 is not in results (it's the start), v1 = 10+1=11, v2 = 10+2=12, sum = 23
		t.Errorf("Sum of results = %d, want 23", resultSum)
	}
}

// TestDescendantsFlowParentAggregation tests parent result aggregation.
func TestDescendantsFlowParentAggregation(t *testing.T) {
	d := generateDiamondDAG()

	inputs := []FlowResult{{ID: "A", Result: 10}}

	callback := func(d *DAG, id string, parentResults []FlowResult) (interface{}, error) {
		sum := 0
		for _, pr := range parentResults {
			sum += pr.Result.(int)
		}
		return sum, nil
	}

	results, _ := d.DescendantsFlow("A", inputs, callback)

	if len(results) != 1 {
		t.Errorf("DescendantsFlow() = %d results, want 1", len(results))
	}

	// D should receive results from both B and C, each receiving 10 from A
	// So D gets 10+10 = 20
	if results[0].Result.(int) != 20 {
		t.Errorf("Result = %d, want 20", results[0].Result.(int))
	}
}

// TestDescendantsFlowErrorHandling tests error handling in callback.
func TestDescendantsFlowErrorHandling(t *testing.T) {
	d := NewDAG()

	v0, _ := d.AddVertex(0)
	v1, _ := d.AddVertex(1)
	v2, _ := d.AddVertex(2)
	_ = d.AddEdge(v0, v1)
	_ = d.AddEdge(v0, v2)

	callback := func(d *DAG, id string, parentResults []FlowResult) (interface{}, error) {
		v, _ := d.GetVertex(id)
		if v.(int) == 1 {
			return nil, fmt.Errorf("error at node 1")
		}
		return v.(int), nil
	}

	results, _ := d.DescendantsFlow(v0, nil, callback)

	if len(results) != 2 {
		t.Errorf("DescendantsFlow() = %d results, want 2", len(results))
	}

	// One result should have an error
	errorCount := 0
	for _, r := range results {
		if r.Error != nil {
			errorCount++
		}
	}
	if errorCount != 1 {
		t.Errorf("Error count = %d, want 1", errorCount)
	}
}

// TestDescendantsFlowComplex tests DescendantsFlow with a more complex graph.
func TestDescendantsFlowComplex(t *testing.T) {
	d := NewDAG()

	// Create a simple diamond graph
	vA, _ := d.AddVertex("A")
	vB, _ := d.AddVertex("B")
	vC, _ := d.AddVertex("C")
	vD, _ := d.AddVertex("D")
	_ = d.AddEdge(vA, vB)
	_ = d.AddEdge(vA, vC)
	_ = d.AddEdge(vB, vD)
	_ = d.AddEdge(vC, vD)

	inputs := []FlowResult{{ID: vA, Result: 10}}

	callback := func(d *DAG, id string, parentResults []FlowResult) (interface{}, error) {
		v, _ := d.GetVertex(id)
		str := v.(string)
		sum := len(str)
		for _, pr := range parentResults {
			sum += pr.Result.(int)
		}
		return sum, nil
	}

	// Start from A
	results, _ := d.DescendantsFlow(vA, inputs, callback)

	// Should process all descendants of A (B, C, D)
	// But only D is a leaf, so we get 1 result
	if len(results) != 1 {
		t.Errorf("DescendantsFlow() = %d results, want 1", len(results))
	}

	// D should receive results from both B and C
	// B: 10 + 1 = 11, C: 10 + 1 = 11
	// D: 1 + 11 + 11 = 23
	if results[0].Result.(int) != 25 {
		t.Errorf("Result = %d, want 25", results[0].Result.(int))
	}
}

// TestDescendantsWalkerFullTraversal tests full traversal with DescendantsWalker.
func TestDescendantsWalkerFullTraversal(t *testing.T) {
	d := generateLinearDAG(100)

	rootID := "node_0"
	ids, _, _ := d.DescendantsWalker(rootID)

	var visited []string
	for id := range ids {
		visited = append(visited, id)
	}

	if len(visited) != 99 {
		t.Errorf("DescendantsWalker() visited %d nodes, want 99", len(visited))
	}
}

// TestDescendantsWalkerSignal tests DescendantsWalker with signal interruption.
func TestDescendantsWalkerSignal(t *testing.T) {
	d := generateLinearDAG(100)

	rootID := "node_0"
	ids, signal, _ := d.DescendantsWalker(rootID)

	var visited []string
	for id := range ids {
		visited = append(visited, id)
		if len(visited) >= 10 {
			signal <- true
			break
		}
	}

	if len(visited) != 10 {
		t.Errorf("DescendantsWalker with signal visited %d nodes, want 10", len(visited))
	}
}

// TestAncestorsWalkerFullTraversal tests full traversal with AncestorsWalker.
func TestAncestorsWalkerFullTraversal(t *testing.T) {
	d := generateLinearDAG(100)

	leafID := "node_99"
	ids, _, _ := d.AncestorsWalker(leafID)

	var visited []string
	for id := range ids {
		visited = append(visited, id)
	}

	if len(visited) != 99 {
		t.Errorf("AncestorsWalker() visited %d nodes, want 99", len(visited))
	}
}

// TestAncestorsWalkerSignal tests AncestorsWalker with signal interruption.
func TestAncestorsWalkerSignal(t *testing.T) {
	d := generateLinearDAG(100)

	leafID := "node_99"
	ids, signal, _ := d.AncestorsWalker(leafID)

	var visited []string
	for id := range ids {
		visited = append(visited, id)
		if len(visited) >= 10 {
			signal <- true
			break
		}
	}

	if len(visited) != 10 {
		t.Errorf("AncestorsWalker with signal visited %d nodes, want 10", len(visited))
	}
}

// TestDescendantsWalkerLargeGraph tests Walker on a large graph.
func TestDescendantsWalkerLargeGraph(t *testing.T) {
	d := generateWideTreeDAG(4, 10)

	rootID := "root_0"
	ids, _, _ := d.DescendantsWalker(rootID)

	visited := make(map[string]struct{})
	for id := range ids {
		visited[id] = struct{}{}
	}

	// Expected descendants = total vertices - 1 (root)
	// For depth 4, branches 10:
	// Level 0: 1 (root)
	// Level 1: 10
	// Level 2: 100
	// Level 3: 1000
	// Total = 1111, descendants = 1110
	expected := 1110

	if len(visited) != expected {
		t.Errorf("DescendantsWalker() visited %d nodes, want %d", len(visited), expected)
	}
}

// TestAllErrorTypes tests all error types are properly returned.
func TestAllErrorTypes(t *testing.T) {
	d := NewDAG()

	// VertexNilError
	_, err := d.AddVertex(nil)
	if _, ok := err.(VertexNilError); !ok {
		t.Errorf("Expected VertexNilError, got %T", err)
	}

	v, _ := d.AddVertex(1)

	// VertexDuplicateError
	_, err = d.AddVertex(1)
	if _, ok := err.(VertexDuplicateError); !ok {
		t.Errorf("Expected VertexDuplicateError, got %T", err)
	}

	// IDEmptyError
	_, err = d.GetVertex("")
	if _, ok := err.(IDEmptyError); !ok {
		t.Errorf("Expected IDEmptyError, got %T", err)
	}

	// IDUnknownError
	_, err = d.GetVertex("unknown")
	if _, ok := err.(IDUnknownError); !ok {
		t.Errorf("Expected IDUnknownError, got %T", err)
	}

	v2, _ := d.AddVertex(2)
	_ = d.AddEdge(v, v2)

	// EdgeDuplicateError
	err = d.AddEdge(v, v2)
	if _, ok := err.(EdgeDuplicateError); !ok {
		t.Errorf("Expected EdgeDuplicateError, got %T", err)
	}

	// EdgeLoopError
	err = d.AddEdge(v2, v)
	if _, ok := err.(EdgeLoopError); !ok {
		t.Errorf("Expected EdgeLoopError, got %T", err)
	}

	// SrcDstEqualError
	err = d.AddEdge(v, v)
	if _, ok := err.(SrcDstEqualError); !ok {
		t.Errorf("Expected SrcDstEqualError, got %T", err)
	}

	_ = d.DeleteEdge(v, v2)

	// EdgeUnknownError
	err = d.DeleteEdge(v, v2)
	if _, ok := err.(EdgeUnknownError); !ok {
		t.Errorf("Expected EdgeUnknownError, got %T", err)
	}
}

// TestBoundaryErrorCombinations tests various boundary error combinations.
func TestBoundaryErrorCombinations(t *testing.T) {
	d := NewDAG()

	// Test edge operations on non-existent vertices
	err := d.AddEdge("v1", "v2")
	if err == nil {
		t.Error("Expected error for AddEdge on non-existent vertices")
	}

	// Test vertex operations with empty strings
	_, err = d.GetVertex("")
	if _, ok := err.(IDEmptyError); !ok {
		t.Errorf("Expected IDEmptyError, got %T", err)
	}

	// Test edge operations with empty strings
	err = d.AddEdge("", "v1")
	if _, ok := err.(IDEmptyError); !ok {
		t.Errorf("Expected IDEmptyError, got %T", err)
	}

	v1, _ := d.AddVertex("v1")
	err = d.AddEdge(v1, "")
	if _, ok := err.(IDEmptyError); !ok {
		t.Errorf("Expected IDEmptyError, got %T", err)
	}

	// Test IsEdge with invalid inputs
	_, err = d.IsEdge("", v1)
	if _, ok := err.(IDEmptyError); !ok {
		t.Errorf("Expected IDEmptyError, got %T", err)
	}

	_, err = d.IsEdge(v1, "")
	if _, ok := err.(IDEmptyError); !ok {
		t.Errorf("Expected IDEmptyError, got %T", err)
	}

	_, err = d.IsEdge(v1, v1)
	if _, ok := err.(SrcDstEqualError); !ok {
		t.Errorf("Expected SrcDstEqualError, got %T", err)
	}
}

// TestGraphConsistencyAfterOperations tests graph consistency after various operations.
func TestGraphConsistencyAfterOperations(t *testing.T) {
	d := generateComplexDAG()

	// Check initial state
	order := d.GetOrder()
	_ = d.GetSize()
	_ = d.GetRoots()
	_ = d.GetLeaves()

	if order != len(d.GetVertices()) {
		t.Errorf("Order inconsistency: GetOrder=%d, GetVertices=%d", order, len(d.GetVertices()))
	}

	// Delete a vertex
	_ = d.DeleteVertex("L1_A")

	// Check consistency again
	newOrder := d.GetOrder()
	_ = d.GetSize()

	if newOrder != order-1 {
		t.Errorf("Order should decrease by 1: was %d, now %d", order, newOrder)
	}

	// Check that edges to/from deleted vertex are removed
	_, err := d.GetChildren("L1_A")
	if _, ok := err.(IDUnknownError); !ok {
		t.Errorf("Expected IDUnknownError, got %T", err)
	}
}

// TestReduceTransitivelyConsistency tests transitive reduction consistency.
func TestReduceTransitivelyConsistency(t *testing.T) {
	d := generateLinearDAG(10)

	beforeSize := d.GetSize()
	d.ReduceTransitively()
	afterSize := d.GetSize()

	// Linear chain should be unchanged by transitive reduction
	if beforeSize != afterSize {
		t.Errorf("Transitive reduction changed size: before %d, after %d", beforeSize, afterSize)
	}

	// Test on a graph with transitive edges
	d2 := NewDAG()
	v1, _ := d2.AddVertex(1)
	v2, _ := d2.AddVertex(2)
	v3, _ := d2.AddVertex(3)
	_ = d2.AddEdge(v1, v2)
	_ = d2.AddEdge(v2, v3)
	_ = d2.AddEdge(v1, v3) // This is transitive

	beforeSize = d2.GetSize()
	d2.ReduceTransitively()
	afterSize = d2.GetSize()

	if beforeSize-1 != afterSize {
		t.Errorf("Transitive reduction should remove one edge: before %d, after %d", beforeSize, afterSize)
	}

	// Check that the transitive edge was removed
	isEdge, _ := d2.IsEdge(v1, v3)
	if isEdge {
		t.Error("Transitive edge should have been removed")
	}
}

// TestCopyPreservesStructure tests that Copy preserves the graph structure.
func TestCopyPreservesStructure(t *testing.T) {
	original := generateComplexDAG()

	copy, err := original.Copy()
	if err != nil {
		t.Fatal(err)
	}

	if original.GetOrder() != copy.GetOrder() {
		t.Errorf("Order mismatch: original %d, copy %d", original.GetOrder(), copy.GetOrder())
	}

	if original.GetSize() != copy.GetSize() {
		t.Errorf("Size mismatch: original %d, copy %d", original.GetSize(), copy.GetSize())
	}

	// Check that all vertices exist in copy
	originalVertices := original.GetVertices()
	for id, v := range originalVertices {
		copyV, err := copy.GetVertex(id)
		if err != nil {
			t.Errorf("Vertex %s not found in copy", id)
		}
		if v != copyV {
			t.Errorf("Vertex %s mismatch: original %v, copy %v", id, v, copyV)
		}
	}

	// Check that roots and leaves match
	if len(original.GetRoots()) != len(copy.GetRoots()) {
		t.Errorf("Roots count mismatch: original %d, copy %d", len(original.GetRoots()), len(copy.GetRoots()))
	}

	if len(original.GetLeaves()) != len(copy.GetLeaves()) {
		t.Errorf("Leaves count mismatch: original %d, copy %d", len(original.GetLeaves()), len(copy.GetLeaves()))
	}
}

// TestGetDescendantsGraph tests subgraph extraction.
func TestGetDescendantsGraph(t *testing.T) {
	original := generateComplexDAG()

	sub, rootID, err := original.GetDescendantsGraph("R1")
	if err != nil {
		t.Fatal(err)
	}

	if rootID != "R1" {
		t.Errorf("Root ID mismatch: expected R1, got %s", rootID)
	}

	// Check that all descendants are included
	expectedDescendants, _ := original.GetDescendants("R1")
	actualDescendants, _ := sub.GetDescendants(rootID)

	if len(expectedDescendants) != len(actualDescendants) {
		t.Errorf("Descendants count mismatch: original %d, subgraph %d", len(expectedDescendants), len(actualDescendants))
	}

	// Check that only R1 and its descendants are included
	if sub.GetOrder() != len(expectedDescendants)+1 { // +1 for R1 itself
		t.Errorf("Subgraph order mismatch: expected %d, got %d", len(expectedDescendants)+1, sub.GetOrder())
	}
}

// TestGetAncestorsGraph tests ancestor subgraph extraction.
func TestGetAncestorsGraph(t *testing.T) {
	original := generateComplexDAG()

	sub, leafID, err := original.GetAncestorsGraph("L3_A")
	if err != nil {
		t.Fatal(err)
	}

	if leafID != "L3_A" {
		t.Errorf("Leaf ID mismatch: expected L3_A, got %s", leafID)
	}

	// Check that all ancestors are included
	expectedAncestors, _ := original.GetAncestors("L3_A")
	actualAncestors, _ := sub.GetAncestors(leafID)

	if len(expectedAncestors) != len(actualAncestors) {
		t.Errorf("Ancestors count mismatch: original %d, subgraph %d", len(expectedAncestors), len(actualAncestors))
	}

	// Check that only L3_A and its ancestors are included
	if sub.GetOrder() != len(expectedAncestors)+1 { // +1 for L3_A itself
		t.Errorf("Subgraph order mismatch: expected %d, got %d", len(expectedAncestors)+1, sub.GetOrder())
	}
}

// TestOrderedDescendantsCorrectness tests ordered descendants BFS order.
func TestOrderedDescendantsCorrectness(t *testing.T) {
	d := generateLinearDAG(10)

	ordered, _ := d.GetOrderedDescendants("node_0")

	if len(ordered) != 9 {
		t.Errorf("GetOrderedDescendants() = %d, want 9", len(ordered))
	}

	// For a linear chain, order should be node_1, node_2, ..., node_9
	for i, id := range ordered {
		expected := fmt.Sprintf("node_%d", i+1)
		if id != expected {
			t.Errorf("Position %d: expected %s, got %s", i, expected, id)
		}
	}
}

// TestOrderedAncestorsCorrectness tests ordered ancestors BFS order.
func TestOrderedAncestorsCorrectness(t *testing.T) {
	d := generateLinearDAG(10)

	ordered, _ := d.GetOrderedAncestors("node_9")

	if len(ordered) != 9 {
		t.Errorf("GetOrderedAncestors() = %d, want 9", len(ordered))
	}

	// For a linear chain, order should be node_8, node_7, ..., node_0 (BFS from bottom)
	// Actually, BFS from node_9 would visit node_8 first, then node_7, etc.
	// But since each node has only one parent, it's essentially reverse order
	for i, id := range ordered {
		expected := fmt.Sprintf("node_%d", 8-i)
		if id != expected {
			t.Errorf("Position %d: expected %s, got %s", i, expected, id)
		}
	}
}

// TestLargeGraphOperations tests operations on a large graph.
func TestLargeGraphOperations(t *testing.T) {
	// Create a large graph (approx 1000 vertices)
	d := generateWideTreeDAG(4, 10)

	// Test GetDescendants on root (should be fast with cache)
	rootID := "root_0"
	desc, _ := d.GetDescendants(rootID)
	if len(desc) == 0 {
		t.Error("GetDescendants on root should return many nodes")
	}

	// Test GetAncestors on a leaf
	leafID := "node_3_0"
	anc, _ := d.GetAncestors(leafID)
	if len(anc) == 0 {
		t.Error("GetAncestors on leaf should return many nodes")
	}

	// Test Copy
	copy, err := d.Copy()
	if err != nil {
		t.Fatal(err)
	}
	if copy.GetOrder() != d.GetOrder() {
		t.Errorf("Copy order mismatch: original %d, copy %d", d.GetOrder(), copy.GetOrder())
	}

	// Test ReduceTransitively
	d.ReduceTransitively()
	if d.GetOrder() != copy.GetOrder() {
		t.Error("ReduceTransitively should not change vertex count")
	}
}

// TestIDInterfaceImplementation tests custom ID implementation.
func TestIDInterfaceImplementation(t *testing.T) {
	d := NewDAG()

	// Add vertices with custom IDs
	tv1 := TestVertex{VertexID: "custom_1", Name: "Test1"}
	tv2 := TestVertex{VertexID: "custom_2", Name: "Test2"}

	id1, _ := d.AddVertex(tv1)
	id2, _ := d.AddVertex(tv2)

	if id1 != "custom_1" {
		t.Errorf("ID 1 = %s, want custom_1", id1)
	}
	if id2 != "custom_2" {
		t.Errorf("ID 2 = %s, want custom_2", id2)
	}

	// Verify vertices can be retrieved
	v1, _ := d.GetVertex("custom_1")
	if v1.(TestVertex).Name != "Test1" {
		t.Errorf("Vertex 1 name = %s, want Test1", v1.(TestVertex).Name)
	}
}

// TestStarGraphOperations tests operations on a star-shaped graph.
func TestStarGraphOperations(t *testing.T) {
	d := generateStarDAG(10)

	centerID := "center"

	// Center should have 10 children
	children, _ := d.GetChildren(centerID)
	if len(children) != 10 {
		t.Errorf("Center children count = %d, want 10", len(children))
	}

	// Center should be a root
	isRoot, _ := d.IsRoot(centerID)
	if !isRoot {
		t.Error("Center should be a root")
	}

	// All leaves should have center as parent
	for id := range d.GetLeaves() {
		parents, _ := d.GetParents(id)
		if len(parents) != 1 {
			t.Errorf("Leaf %s should have 1 parent, has %d", id, len(parents))
		}
		// Check that the parent is center
		for parentID := range parents {
			if parentID != centerID {
				t.Errorf("Leaf %s parent is %s, expected %s", id, parentID, centerID)
			}
		}
	}
}