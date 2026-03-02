package dag

import (
	"strconv"
	"testing"
)

func BenchmarkGetDescendantsGraphByDepth(b *testing.B) {
	tests := []struct {
		name    string
		size    int
		depth   int
		prepare func() *GenericDAG[string]
	}{
		{
			name: "small_graph_depth_1",
			size:  100,
			depth: 1,
			prepare: func() *GenericDAG[string] {
				return buildLinearDAG(100)
			},
		},
		{
			name: "small_graph_depth_10",
			size:  100,
			depth: 10,
			prepare: func() *GenericDAG[string] {
				return buildLinearDAG(100)
			},
		},
		{
			name: "medium_graph_depth_1",
			size:  1000,
			depth: 1,
			prepare: func() *GenericDAG[string] {
				return buildLinearDAG(1000)
			},
		},
		{
			name: "medium_graph_depth_10",
			size:  1000,
			depth: 10,
			prepare: func() *GenericDAG[string] {
				return buildLinearDAG(1000)
			},
		},
		{
			name: "large_graph_depth_1",
			size:  10000,
			depth: 1,
			prepare: func() *GenericDAG[string] {
				return buildLinearDAG(10000)
			},
		},
		{
			name: "large_graph_depth_100",
			size:  10000,
			depth: 100,
			prepare: func() *GenericDAG[string] {
				return buildLinearDAG(10000)
			},
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			dag := tt.prepare()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _, _ = dag.GetDescendantsGraphByDepth("0", tt.depth)
			}
		})
	}
}

func BenchmarkGetEdges(b *testing.B) {
	tests := []struct {
		name    string
		size    int
		prepare func() *GenericDAG[string]
	}{
		{
			name: "small_graph_100",
			size:  100,
			prepare: func() *GenericDAG[string] {
				return buildLinearDAG(100)
			},
		},
		{
			name: "medium_graph_1000",
			size:  1000,
			prepare: func() *GenericDAG[string] {
				return buildLinearDAG(1000)
			},
		},
		{
			name: "large_graph_10000",
			size:  10000,
			prepare: func() *GenericDAG[string] {
				return buildLinearDAG(10000)
			},
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			dag := tt.prepare()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = dag.GetEdges()
			}
		})
	}
}

func BenchmarkGetEdgesWithOption(b *testing.B) {
	tests := []struct {
		name    string
		size    int
		option  CopyOption
		prepare func() *GenericDAG[string]
	}{
		{
			name:    "share_data_small",
			size:    100,
			option:  ShareData,
			prepare: func() *GenericDAG[string] { return buildLinearDAG(100) },
		},
		{
			name:    "copy_data_small",
			size:    100,
			option:  CopyData,
			prepare: func() *GenericDAG[string] { return buildLinearDAG(100) },
		},
		{
			name:    "share_data_medium",
			size:    1000,
			option:  ShareData,
			prepare: func() *GenericDAG[string] { return buildLinearDAG(1000) },
		},
		{
			name:    "copy_data_medium",
			size:    1000,
			option:  CopyData,
			prepare: func() *GenericDAG[string] { return buildLinearDAG(1000) },
		},
		{
			name:    "share_data_large",
			size:    10000,
			option:  ShareData,
			prepare: func() *GenericDAG[string] { return buildLinearDAG(10000) },
		},
		{
			name:    "copy_data_large",
			size:    10000,
			option:  CopyData,
			prepare: func() *GenericDAG[string] { return buildLinearDAG(10000) },
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			dag := tt.prepare()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = dag.GetEdgesWithOption(tt.option)
			}
		})
	}
}

func BenchmarkGetVerticesList(b *testing.B) {
	tests := []struct {
		name    string
		size    int
		prepare func() *GenericDAG[string]
	}{
		{
			name: "small_graph_100",
			size:  100,
			prepare: func() *GenericDAG[string] {
				return buildLinearDAG(100)
			},
		},
		{
			name: "medium_graph_1000",
			size:  1000,
			prepare: func() *GenericDAG[string] {
				return buildLinearDAG(1000)
			},
		},
		{
			name: "large_graph_10000",
			size:  10000,
			prepare: func() *GenericDAG[string] {
				return buildLinearDAG(10000)
			},
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			dag := tt.prepare()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = dag.GetVerticesList()
			}
		})
	}
}

func BenchmarkGetVerticesListWithOption(b *testing.B) {
	tests := []struct {
		name    string
		size    int
		option  CopyOption
		prepare func() *GenericDAG[string]
	}{
		{
			name:    "share_data_small",
			size:    100,
			option:  ShareData,
			prepare: func() *GenericDAG[string] { return buildLinearDAG(100) },
		},
		{
			name:    "copy_data_small",
			size:    100,
			option:  CopyData,
			prepare: func() *GenericDAG[string] { return buildLinearDAG(100) },
		},
		{
			name:    "share_data_medium",
			size:    1000,
			option:  ShareData,
			prepare: func() *GenericDAG[string] { return buildLinearDAG(1000) },
		},
		{
			name:    "copy_data_medium",
			size:    1000,
			option:  CopyData,
			prepare: func() *GenericDAG[string] { return buildLinearDAG(1000) },
		},
		{
			name:    "share_data_large",
			size:    10000,
			option:  ShareData,
			prepare: func() *GenericDAG[string] { return buildLinearDAG(10000) },
		},
		{
			name:    "copy_data_large",
			size:    10000,
			option:  CopyData,
			prepare: func() *GenericDAG[string] { return buildLinearDAG(10000) },
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			dag := tt.prepare()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = dag.GetVerticesListWithOption(tt.option)
			}
		})
	}
}

func BenchmarkGetEdgesByDepth(b *testing.B) {
	tests := []struct {
		name     string
		size     int
		minDepth int
		maxDepth int
		prepare  func() *GenericDAG[string]
	}{
		{
			name:     "small_graph_depth_0_1",
			size:     100,
			minDepth: 0,
			maxDepth: 1,
			prepare:  func() *GenericDAG[string] { return buildLinearDAG(100) },
		},
		{
			name:     "small_graph_depth_5_10",
			size:     100,
			minDepth: 5,
			maxDepth: 10,
			prepare:  func() *GenericDAG[string] { return buildLinearDAG(100) },
		},
		{
			name:     "medium_graph_depth_0_10",
			size:     1000,
			minDepth: 0,
			maxDepth: 10,
			prepare:  func() *GenericDAG[string] { return buildLinearDAG(1000) },
		},
		{
			name:     "large_graph_depth_0_100",
			size:     10000,
			minDepth: 0,
			maxDepth: 100,
			prepare:  func() *GenericDAG[string] { return buildLinearDAG(10000) },
		},
		{
			name:     "large_graph_unlimited",
			size:     10000,
			minDepth: 0,
			maxDepth: -1,
			prepare:  func() *GenericDAG[string] { return buildLinearDAG(10000) },
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			dag := tt.prepare()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = dag.GetEdgesByDepth("0", tt.minDepth, tt.maxDepth)
			}
		})
	}
}

// buildLinearDAG creates a linear DAG with the specified number of nodes
// node0 -> node1 -> node2 -> ... -> node(n-1)
func buildLinearDAG(n int) *GenericDAG[string] {
	dag := NewGenericDAG[string]()
	for i := 0; i < n; i++ {
		dag.AddVertexByID(strconv.Itoa(i), strconv.Itoa(i))
	}
	for i := 0; i < n-1; i++ {
		dag.AddEdge(strconv.Itoa(i), strconv.Itoa(i+1))
	}
	return dag
}