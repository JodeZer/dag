package dag

import (
	"fmt"
	"testing"
)

// ============================================================================
// Core Operations Benchmarks
// ============================================================================

func BenchmarkAddVertex(b *testing.B) {
	d := NewDAG()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = d.AddVertex(i)
	}
}

func BenchmarkAddVertexByID(b *testing.B) {
	d := NewDAG()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id := fmt.Sprintf("vertex_%d", i)
		_ = d.AddVertexByID(id, i)
	}
}

func BenchmarkAddEdge(b *testing.B) {
	d := NewDAG()
	numVertices := 10000

	// Pre-populate vertices
	ids := make([]string, numVertices)
	for i := 0; i < numVertices; i++ {
		id := fmt.Sprintf("vertex_%d", i)
		ids[i] = id
		_ = d.AddVertexByID(id, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		src := ids[i%numVertices]
		dst := ids[(i+1)%numVertices]
		_ = d.AddEdge(src, dst)
	}
}

func BenchmarkDeleteVertex(b *testing.B) {
	b.StopTimer()

	// Create a large graph for each iteration
	for i := 0; i < b.N; i++ {
		d := generateLinearDAG(1000)
		ids := make([]string, d.GetOrder())
		j := 0
		for id := range d.GetVertices() {
			ids[j] = id
			j++
		}

		b.StartTimer()
		_ = d.DeleteVertex(ids[i%len(ids)])
		b.StopTimer()
	}
}

func BenchmarkDeleteEdge(b *testing.B) {
	b.StopTimer()

	// Create a large graph for each iteration
	for i := 0; i < b.N; i++ {
		d := generateLinearDAG(1000)

		b.StartTimer()
		// Delete an edge from the middle
		src := "node_500"
		dst := "node_501"
		_ = d.DeleteEdge(src, dst)
		b.StopTimer()
	}
}

func BenchmarkGetVertex(b *testing.B) {
	d := generateLinearDAG(10000)
	vertexID := "node_5000"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = d.GetVertex(vertexID)
	}
}

func BenchmarkGetChildren(b *testing.B) {
	d := generateWideTreeDAG(4, 10)
	rootID := "root_0"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = d.GetChildren(rootID)
	}
}

func BenchmarkGetParents(b *testing.B) {
	d := generateWideTreeDAG(4, 10)
	leafID := "node_3_0"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = d.GetParents(leafID)
	}
}

func BenchmarkIsEdge(b *testing.B) {
	d := generateLinearDAG(1000)
	src := "node_500"
	dst := "node_501"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = d.IsEdge(src, dst)
	}
}

// ============================================================================
// Query Operations Benchmarks (with cache)
// ============================================================================

func BenchmarkGetDescendants(b *testing.B) {
	d := generateWideTreeDAG(4, 10)
	rootID := "root_0"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = d.GetDescendants(rootID)
	}
}

func BenchmarkGetDescendantsCached(b *testing.B) {
	d := generateWideTreeDAG(4, 10)
	rootID := "root_0"

	// Populate cache
	_, _ = d.GetDescendants(rootID)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = d.GetDescendants(rootID)
	}
}

func BenchmarkGetAncestors(b *testing.B) {
	d := generateWideTreeDAG(4, 10)
	leafID := "node_3_0"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = d.GetAncestors(leafID)
	}
}

func BenchmarkGetAncestorsCached(b *testing.B) {
	d := generateWideTreeDAG(4, 10)
	leafID := "node_3_0"

	// Populate cache
	_, _ = d.GetAncestors(leafID)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = d.GetAncestors(leafID)
	}
}

func BenchmarkGetOrderedDescendants(b *testing.B) {
	d := generateLinearDAG(1000)
	rootID := "node_0"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = d.GetOrderedDescendants(rootID)
	}
}

func BenchmarkGetOrderedAncestors(b *testing.B) {
	d := generateLinearDAG(1000)
	leafID := "node_999"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = d.GetOrderedAncestors(leafID)
	}
}

// ============================================================================
// Traversal Operations Benchmarks
// ============================================================================

type benchmarkVisitor struct {
	Count int
}

func (pv *benchmarkVisitor) Visit(v Vertexer) {
	pv.Count++
}

func BenchmarkDFSWalk(b *testing.B) {
	d := getTestWalkDAG() // Use the existing test DAG with string vertices
	visitor := &benchmarkVisitor{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		visitor.Count = 0
		d.DFSWalk(visitor)
	}
}

func BenchmarkBFSWalk(b *testing.B) {
	d := getTestWalkDAG() // Use the existing test DAG with string vertices
	visitor := &benchmarkVisitor{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		visitor.Count = 0
		d.BFSWalk(visitor)
	}
}

func BenchmarkOrderedWalk(b *testing.B) {
	d := getTestWalkDAG() // Use the existing test DAG with string vertices
	visitor := &benchmarkVisitor{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		visitor.Count = 0
		d.OrderedWalk(visitor)
	}
}

func BenchmarkDescendantsWalker(b *testing.B) {
	d := generateWideTreeDAG(4, 10)
	rootID := "root_0"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ids, _, _ := d.DescendantsWalker(rootID)
		for range ids {
			// Consume all elements
		}
	}
}

func BenchmarkAncestorsWalker(b *testing.B) {
	d := generateWideTreeDAG(4, 10)
	leafID := "node_3_0"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ids, _, _ := d.AncestorsWalker(leafID)
		for range ids {
			// Consume all elements
		}
	}
}

// ============================================================================
// Graph Operations Benchmarks
// ============================================================================

func BenchmarkReduceTransitively(b *testing.B) {
	d := generateDenseDAG(50)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.ReduceTransitively()
	}
}

func BenchmarkCopy(b *testing.B) {
	d := generateWideTreeDAG(4, 10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = d.Copy()
	}
}

func BenchmarkFlushCaches(b *testing.B) {
	d := generateWideTreeDAG(4, 10)

	// Populate caches
	d.GetDescendants("root_0")
	d.GetAncestors("node_3_0")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.FlushCaches()
	}
}

func BenchmarkGetDescendantsGraph(b *testing.B) {
	d := generateWideTreeDAG(4, 10)
	rootID := "root_0"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = d.GetDescendantsGraph(rootID)
	}
}

func BenchmarkGetAncestorsGraph(b *testing.B) {
	d := generateWideTreeDAG(4, 10)
	leafID := "node_3_0"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = d.GetAncestorsGraph(leafID)
	}
}

// ============================================================================
// Flow Operations Benchmarks
// ============================================================================

func BenchmarkDescendantsFlow(b *testing.B) {
	d := generateWideTreeDAG(4, 10)
	rootID := "root_0"

	callback := func(d *DAG, id string, parentResults []FlowResult) (interface{}, error) {
		return len(id), nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = d.DescendantsFlow(rootID, nil, callback)
	}
}

// ============================================================================
// Serialization Benchmarks
// ============================================================================

func BenchmarkMarshalJSON(b *testing.B) {
	d := generateWideTreeDAG(4, 10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = d.MarshalJSON()
	}
}

func BenchmarkUnmarshalJSON(b *testing.B) {
	d := generateWideTreeDAG(4, 10)
	data, _ := d.MarshalJSON()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		newDAG := NewDAG()
		_ = newDAG.UnmarshalJSON(data)
	}
}

// ============================================================================
// Large Scale Serialization Benchmarks (100k nodes)
// ============================================================================

func BenchmarkMarshalJSON_100k_3Branch(b *testing.B) {
	d := generateBalancedTreeDAG(100000, 3)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = d.MarshalJSON()
	}
}

func BenchmarkMarshalJSON_100k_4Branch(b *testing.B) {
	d := generateBalancedTreeDAG(100000, 4)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = d.MarshalJSON()
	}
}

func BenchmarkMarshalJSON_100k_5Branch(b *testing.B) {
	d := generateBalancedTreeDAG(100000, 5)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = d.MarshalJSON()
	}
}

func BenchmarkGetDescendants_100k_3Branch(b *testing.B) {
	d := generateBalancedTreeDAG(100000, 3)
	rootID := "root_0"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = d.GetDescendants(rootID)
	}
}

func BenchmarkGetDescendants_100k_4Branch(b *testing.B) {
	d := generateBalancedTreeDAG(100000, 4)
	rootID := "root_0"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = d.GetDescendants(rootID)
	}
}

func BenchmarkGetDescendants_100k_5Branch(b *testing.B) {
	d := generateBalancedTreeDAG(100000, 5)
	rootID := "root_0"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = d.GetDescendants(rootID)
	}
}

// ============================================================================
// Different Graph Structure Benchmarks
// ============================================================================

func BenchmarkLinearGraphGetDescendants(b *testing.B) {
	d := generateLinearDAG(10000)
	rootID := "node_0"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = d.GetDescendants(rootID)
	}
}

func BenchmarkWideTreeGraphGetDescendants(b *testing.B) {
	d := generateWideTreeDAG(4, 100)
	rootID := "root_0"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = d.GetDescendants(rootID)
	}
}

func BenchmarkDeepTreeGraphGetDescendants(b *testing.B) {
	d := generateDeepTreeDAG(1000)
	rootID := "root_0"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = d.GetDescendants(rootID)
	}
}

func BenchmarkDiamondGraphGetDescendants(b *testing.B) {
	d := generateMultiDiamondDAG()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = d.GetDescendants("A")
	}
}

func BenchmarkRandomGraphGetDescendants(b *testing.B) {
	d := generateRandomDAG(1000, 3000)

	// Find a node with many descendants
	bestID := ""
	maxDescendants := 0
	for id := range d.GetVertices() {
		desc, _ := d.GetDescendants(id)
		if len(desc) > maxDescendants {
			maxDescendants = len(desc)
			bestID = id
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = d.GetDescendants(bestID)
	}
}

func BenchmarkStarGraphGetDescendants(b *testing.B) {
	d := generateStarDAG(1000)
	centerID := "center"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = d.GetDescendants(centerID)
	}
}

func BenchmarkDenseGraphGetDescendants(b *testing.B) {
	d := generateDenseDAG(100)
	rootID := "node_0"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = d.GetDescendants(rootID)
	}
}

// ============================================================================
// Scale Benchmarks
// ============================================================================

func BenchmarkGetDescendants_Scale_100(b *testing.B) {
	d := generateLinearDAG(100)
	rootID := "node_0"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = d.GetDescendants(rootID)
	}
}

func BenchmarkGetDescendants_Scale_1000(b *testing.B) {
	d := generateLinearDAG(1000)
	rootID := "node_0"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = d.GetDescendants(rootID)
	}
}

func BenchmarkGetDescendants_Scale_10000(b *testing.B) {
	d := generateLinearDAG(10000)
	rootID := "node_0"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = d.GetDescendants(rootID)
	}
}

func BenchmarkGetDescendants_Scale_100000(b *testing.B) {
	d := generateLinearDAG(100000)
	rootID := "node_0"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = d.GetDescendants(rootID)
	}
}

func BenchmarkCopy_Scale_100(b *testing.B) {
	d := generateLinearDAG(100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = d.Copy()
	}
}

func BenchmarkCopy_Scale_1000(b *testing.B) {
	d := generateLinearDAG(1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = d.Copy()
	}
}

func BenchmarkCopy_Scale_10000(b *testing.B) {
	d := generateLinearDAG(10000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = d.Copy()
	}
}

// ============================================================================
// Concurrent Benchmarks
// ============================================================================

func BenchmarkConcurrentGetDescendants(b *testing.B) {
	d := generateWideTreeDAG(4, 10)
	rootID := "root_0"

	// Populate cache
	d.GetDescendants(rootID)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = d.GetDescendants(rootID)
		}
	})
}

func BenchmarkConcurrentGetAncestors(b *testing.B) {
	d := generateWideTreeDAG(4, 10)
	leafID := "node_3_0"

	// Populate cache
	d.GetAncestors(leafID)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = d.GetAncestors(leafID)
		}
	})
}

func BenchmarkConcurrentReadWrite(b *testing.B) {
	d := NewDAG()

	// Pre-populate vertices with string IDs
	for i := 0; i < 1000; i++ {
		_, _ = d.AddVertex(fmt.Sprintf("vertex_%d", i))
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i%2 == 0 {
				// Read
				_, _ = d.GetVertex(fmt.Sprintf("vertex_%d", i%1000))
				d.GetVertices()
			} else {
				// Write - add new vertices
				_, _ = d.AddVertex(fmt.Sprintf("vertex_%d", 1000+i))
			}
			i++
		}
	})
}

// ============================================================================
// Special Operation Benchmarks
// ============================================================================

func BenchmarkGetRoots(b *testing.B) {
	d := generateWideTreeDAG(4, 10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = d.GetRoots()
	}
}

func BenchmarkGetLeaves(b *testing.B) {
	d := generateWideTreeDAG(4, 10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = d.GetLeaves()
	}
}

func BenchmarkGetOrder(b *testing.B) {
	d := generateWideTreeDAG(4, 10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = d.GetOrder()
	}
}

func BenchmarkGetSize(b *testing.B) {
	d := generateWideTreeDAG(4, 10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = d.GetSize()
	}
}

func BenchmarkIsLeaf(b *testing.B) {
	d := generateWideTreeDAG(4, 10)
	leafID := "node_3_0"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = d.IsLeaf(leafID)
	}
}

func BenchmarkIsRoot(b *testing.B) {
	d := generateWideTreeDAG(4, 10)
	rootID := "root_0"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = d.IsRoot(rootID)
	}
}

func BenchmarkGetVertices(b *testing.B) {
	d := generateWideTreeDAG(4, 10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = d.GetVertices()
	}
}

// ============================================================================
// Cache Performance Benchmarks
// ============================================================================

func BenchmarkCacheMissVsHit(b *testing.B) {
	d := generateWideTreeDAG(4, 10)
	rootID := "root_0"

	// First run - cache miss
	b.Run("CacheMiss", func(b *testing.B) {
		d2, _ := d.Copy()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = d2.GetDescendants(rootID)
		}
	})

	// Second run - cache hit
	b.Run("CacheHit", func(b *testing.B) {
		d2, _ := d.Copy()
		// Populate cache
		d2.GetDescendants(rootID)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = d2.GetDescendants(rootID)
		}
	})
}

func BenchmarkCacheInvalidationOnDeleteVertex(b *testing.B) {
	b.StopTimer()

	for i := 0; i < b.N; i++ {
		d := generateWideTreeDAG(4, 10)
		rootID := "root_0"

		// Populate cache
		d.GetDescendants(rootID)

		b.StartTimer()
		_ = d.DeleteVertex("node_2_0")
		b.StopTimer()
	}
}

func BenchmarkCacheInvalidationOnDeleteEdge(b *testing.B) {
	b.StopTimer()

	for i := 0; i < b.N; i++ {
		d := generateLinearDAG(100)
		rootID := "node_0"

		// Populate cache
		d.GetDescendants(rootID)

		b.StartTimer()
		_ = d.DeleteEdge("node_50", "node_51")
		b.StopTimer()
	}
}

// ============================================================================
// Edge Detection Loop Benchmarks
// ============================================================================

func BenchmarkAddEdgeLoopDetectionLinear(b *testing.B) {
	d := generateLinearDAG(100)
	newVertexID := "new_vertex"
	_, _ = d.AddVertex(newVertexID)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Try to add edge that would create a loop (should fail)
		src := "node_0"
		dst := newVertexID
		_ = d.AddEdge(src, dst)
		_ = d.DeleteEdge(src, dst)
	}
}

func BenchmarkAddEdgeLoopDetectionComplex(b *testing.B) {
	d := generateRandomDAG(100, 300)
	newVertexID := "new_vertex"
	_, _ = d.AddVertex(newVertexID)

	// Add an edge from a random node to new vertex
	for id := range d.GetVertices() {
		_ = d.AddEdge(id, newVertexID)
		break
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Try to add edge that would create a loop (should fail)
		src := "node_0"
		dst := newVertexID
		_ = d.AddEdge(src, dst)
		_ = d.DeleteEdge(src, dst)
	}
}

// ============================================================================
// DescendantsFlow Benchmarks with Different Patterns
// ============================================================================

func BenchmarkDescendantsFlowLinear(b *testing.B) {
	d := generateLinearDAG(1000)
	rootID := "node_0"

	callback := func(d *DAG, id string, parentResults []FlowResult) (interface{}, error) {
		return len(id), nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = d.DescendantsFlow(rootID, nil, callback)
	}
}

func BenchmarkDescendantsFlowParallel(b *testing.B) {
	d := generateStarDAG(100)
	centerID := "center"

	callback := func(d *DAG, id string, parentResults []FlowResult) (interface{}, error) {
		return len(id), nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = d.DescendantsFlow(centerID, nil, callback)
	}
}

// ============================================================================
// Memory Allocation Benchmarks
// ============================================================================

func BenchmarkGetDescendantsAllocs(b *testing.B) {
	d := generateWideTreeDAG(4, 10)
	rootID := "root_0"

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = d.GetDescendants(rootID)
	}
}

func BenchmarkGetDescendantsCachedAllocs(b *testing.B) {
	d := generateWideTreeDAG(4, 10)
	rootID := "root_0"
	d.GetDescendants(rootID)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = d.GetDescendants(rootID)
	}
}

func BenchmarkCopyAllocs(b *testing.B) {
	d := generateWideTreeDAG(4, 10)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = d.Copy()
	}
}

func BenchmarkAddEdgeAllocs(b *testing.B) {
	d := NewDAG()
	numVertices := 1000

	// Pre-populate vertices
	ids := make([]string, numVertices)
	for i := 0; i < numVertices; i++ {
		id := fmt.Sprintf("vertex_%d", i)
		ids[i] = id
		_ = d.AddVertexByID(id, i)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		src := ids[i%numVertices]
		dst := ids[(i+1)%numVertices]
		_ = d.AddEdge(src, dst)
	}
}