package dag

import (
	"math/rand"
	"strconv"
	"time"
)

// TestVertex is a simple vertex type for testing.
type TestVertex struct {
	VertexID string
	Name     string
}

// ID implements the IDInterface for TestVertex.
func (tv TestVertex) ID() string {
	return tv.VertexID
}

// generateLinearDAG creates a linear chain graph: 1->2->3->...->n
func generateLinearDAG(size int) *DAG {
	d := NewDAG()
	var prevID string

	for i := 0; i < size; i++ {
		id := "node_" + strconv.Itoa(i)
		v := TestVertex{VertexID: id, Name: "Node" + strconv.Itoa(i)}
		_, _ = d.AddVertex(v)
		if prevID != "" {
			_ = d.AddEdge(prevID, id)
		}
		prevID = id
	}
	return d
}

// generateWideTreeDAG creates a tree with specified depth and branching factor.
// The root has 'branches' children, each of which has 'branches' children, etc.
func generateWideTreeDAG(depth, branches int) *DAG {
	d := NewDAG()
	rootID := "root_0"
	_, _ = d.AddVertex(TestVertex{VertexID: rootID, Name: "Root"})

	currentLevel := []string{rootID}

	for level := 1; level < depth; level++ {
		var nextLevel []string
		nodeCounter := 0

		for _, parentID := range currentLevel {
			for b := 0; b < branches; b++ {
				id := "node_" + strconv.Itoa(level) + "_" + strconv.Itoa(nodeCounter)
				_, _ = d.AddVertex(TestVertex{VertexID: id, Name: "Node" + strconv.Itoa(level) + "_" + strconv.Itoa(nodeCounter)})
				_ = d.AddEdge(parentID, id)
				nextLevel = append(nextLevel, id)
				nodeCounter++
			}
		}
		currentLevel = nextLevel
	}

	return d
}

// generateDeepTreeDAG creates a tree with specified depth where each node has only one child.
// This creates a linear-like structure but with vertex names indicating depth.
func generateDeepTreeDAG(depth int) *DAG {
	d := NewDAG()
	rootID := "root_0"
	_, _ = d.AddVertex(TestVertex{VertexID: rootID, Name: "Root"})

	parentID := rootID
	for i := 1; i < depth; i++ {
		id := "node_" + strconv.Itoa(i)
		_, _ = d.AddVertex(TestVertex{VertexID: id, Name: "Node" + strconv.Itoa(i)})
		_ = d.AddEdge(parentID, id)
		parentID = id
	}

	return d
}

// generateDiamondDAG creates a diamond dependency graph:
//     A
//    / \
//   B   C
//    \ /
//     D
func generateDiamondDAG() *DAG {
	d := NewDAG()

	_, _ = d.AddVertex(TestVertex{VertexID: "A", Name: "A"})
	_, _ = d.AddVertex(TestVertex{VertexID: "B", Name: "B"})
	_, _ = d.AddVertex(TestVertex{VertexID: "C", Name: "C"})
	_, _ = d.AddVertex(TestVertex{VertexID: "D", Name: "D"})

	_ = d.AddEdge("A", "B")
	_ = d.AddEdge("A", "C")
	_ = d.AddEdge("B", "D")
	_ = d.AddEdge("C", "D")

	return d
}

// generateMultiDiamondDAG creates a larger diamond-like structure with multiple levels.
// Level 0: A
// Level 1: B, C
// Level 2: D, E
// Level 3: F
func generateMultiDiamondDAG() *DAG {
	d := NewDAG()

	_, _ = d.AddVertex(TestVertex{VertexID: "A", Name: "A"})
	_, _ = d.AddVertex(TestVertex{VertexID: "B", Name: "B"})
	_, _ = d.AddVertex(TestVertex{VertexID: "C", Name: "C"})
	_, _ = d.AddVertex(TestVertex{VertexID: "D", Name: "D"})
	_, _ = d.AddVertex(TestVertex{VertexID: "E", Name: "E"})
	_, _ = d.AddVertex(TestVertex{VertexID: "F", Name: "F"})

	_ = d.AddEdge("A", "B")
	_ = d.AddEdge("A", "C")
	_ = d.AddEdge("B", "D")
	_ = d.AddEdge("B", "E")
	_ = d.AddEdge("C", "D")
	_ = d.AddEdge("C", "E")
	_ = d.AddEdge("D", "F")
	_ = d.AddEdge("E", "F")

	return d
}

// generateRandomDAG creates a random DAG with the specified number of vertices and edges.
// The vertices are created in a way that ensures no cycles are formed.
func generateRandomDAG(vertices, edges int) *DAG {
	d := NewDAG()
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Add vertices
	vertexIDs := make([]string, vertices)
	for i := 0; i < vertices; i++ {
		id := "node_" + strconv.Itoa(i)
		vertexIDs[i] = id
		_, _ = d.AddVertex(TestVertex{VertexID: id, Name: "Node" + strconv.Itoa(i)})
	}

	// Add edges ensuring no cycles by only adding edges from lower to higher indices
	edgesAdded := 0
	attempts := 0
	maxAttempts := edges * 10

	for edgesAdded < edges && attempts < maxAttempts {
		attempts++

		// Pick two random vertices
		src := vertexIDs[r.Intn(vertices)]
		dst := vertexIDs[r.Intn(vertices)]

		// Extract numeric part for comparison
		srcNum := 0
		dstNum := 0
		for _, c := range src {
			if c >= '0' && c <= '9' {
				srcNum = srcNum*10 + int(c)
			}
		}
		for _, c := range dst {
			if c >= '0' && c <= '9' {
				dstNum = dstNum*10 + int(c)
			}
		}

		// Only add edge if src < dst to avoid cycles
		if srcNum < dstNum && src != dst {
			err := d.AddEdge(src, dst)
			if err == nil {
				edgesAdded++
			}
		}
	}

	return d
}

// generateComplexDAG creates a complex DAG with multiple roots and branches.
// This graph has:
// - 3 root nodes
// - Multiple intermediate nodes with varying connectivity
// - Multiple leaf nodes
func generateComplexDAG() *DAG {
	d := NewDAG()

	// Roots
	_, _ = d.AddVertex(TestVertex{VertexID: "R1", Name: "Root1"})
	_, _ = d.AddVertex(TestVertex{VertexID: "R2", Name: "Root2"})
	_, _ = d.AddVertex(TestVertex{VertexID: "R3", Name: "Root3"})

	// Level 1
	_, _ = d.AddVertex(TestVertex{VertexID: "L1_A", Name: "L1_A"})
	_, _ = d.AddVertex(TestVertex{VertexID: "L1_B", Name: "L1_B"})
	_, _ = d.AddVertex(TestVertex{VertexID: "L1_C", Name: "L1_C"})
	_, _ = d.AddVertex(TestVertex{VertexID: "L1_D", Name: "L1_D"})

	// Level 2
	_, _ = d.AddVertex(TestVertex{VertexID: "L2_A", Name: "L2_A"})
	_, _ = d.AddVertex(TestVertex{VertexID: "L2_B", Name: "L2_B"})
	_, _ = d.AddVertex(TestVertex{VertexID: "L2_C", Name: "L2_C"})

	// Level 3 (leaves)
	_, _ = d.AddVertex(TestVertex{VertexID: "L3_A", Name: "L3_A"})
	_, _ = d.AddVertex(TestVertex{VertexID: "L3_B", Name: "L3_B"})
	_, _ = d.AddVertex(TestVertex{VertexID: "L3_C", Name: "L3_C"})

	// Edges from R1
	_ = d.AddEdge("R1", "L1_A")
	_ = d.AddEdge("R1", "L1_B")

	// Edges from R2
	_ = d.AddEdge("R2", "L1_B")
	_ = d.AddEdge("R2", "L1_C")

	// Edges from R3
	_ = d.AddEdge("R3", "L1_C")
	_ = d.AddEdge("R3", "L1_D")

	// Edges from Level 1
	_ = d.AddEdge("L1_A", "L2_A")
	_ = d.AddEdge("L1_B", "L2_A")
	_ = d.AddEdge("L1_B", "L2_B")
	_ = d.AddEdge("L1_C", "L2_B")
	_ = d.AddEdge("L1_C", "L2_C")
	_ = d.AddEdge("L1_D", "L2_C")

	// Edges from Level 2
	_ = d.AddEdge("L2_A", "L3_A")
	_ = d.AddEdge("L2_A", "L3_B")
	_ = d.AddEdge("L2_B", "L3_B")
	_ = d.AddEdge("L2_B", "L3_C")
	_ = d.AddEdge("L2_C", "L3_C")

	return d
}

// generateStarDAG creates a star-shaped graph where one central node connects to many leaves.
func generateStarDAG(numLeaves int) *DAG {
	d := NewDAG()

	centerID := "center"
	_, _ = d.AddVertex(TestVertex{VertexID: centerID, Name: "Center"})

	for i := 0; i < numLeaves; i++ {
		id := "leaf_" + strconv.Itoa(i)
		_, _ = d.AddVertex(TestVertex{VertexID: id, Name: "Leaf" + strconv.Itoa(i)})
		_ = d.AddEdge(centerID, id)
	}

	return d
}

// generateDenseDAG creates a densely connected DAG where most possible edges exist.
func generateDenseDAG(size int) *DAG {
	d := NewDAG()

	// Add vertices
	vertexIDs := make([]string, size)
	for i := 0; i < size; i++ {
		id := "node_" + strconv.Itoa(i)
		vertexIDs[i] = id
		_, _ = d.AddVertex(TestVertex{VertexID: id, Name: "Node" + strconv.Itoa(i)})
	}

	// Add edges - for each pair (i, j) where i < j, add edge
	for i := 0; i < size; i++ {
		for j := i + 1; j < size; j++ {
			_ = d.AddEdge(vertexIDs[i], vertexIDs[j])
		}
	}

	return d
}

// calculateTreeDepth calculates the depth of a balanced tree given the node count and branching factor.
// For a balanced tree: total_nodes = (branches^depth - 1) / (branches - 1)
// Solving for depth: depth = floor(log(1 + node_count * (branches - 1)) / log(branches))
func calculateTreeDepth(nodeCount, branches int) int {
	if nodeCount <= 1 {
		return 1
	}
	if branches <= 1 {
		return nodeCount
	}

	// Using the formula for geometric series sum
	// N = (b^depth - 1) / (b - 1)
	// N * (b - 1) = b^depth - 1
	// b^depth = N * (b - 1) + 1
	// depth = log(N * (b - 1) + 1) / log(b)

	target := float64(nodeCount*(branches-1) + 1)
	base := float64(branches)

	depth := 1.0
	for {
		if powFloat(base, depth) >= target {
			break
		}
		depth++
	}

	return int(depth)
}

// powFloat calculates base^exp using integer exponentiation
func powFloat(base, exp float64) float64 {
	result := 1.0
	expInt := int(exp)
	for i := 0; i < expInt; i++ {
		result *= base
	}
	return result
}

// generateBalancedTreeDAG generates a balanced tree DAG with approximately the specified number of nodes.
// The tree will have the given branching factor (number of children per node).
// The actual number of nodes will be at least nodeCount, using a complete tree structure.
func generateBalancedTreeDAG(nodeCount, branches int) *DAG {
	d := NewDAG()

	if nodeCount <= 0 {
		return d
	}

	depth := calculateTreeDepth(nodeCount, branches)
	nodesPerLevel := make([][]string, depth)

	// Generate root
	rootID := "root_0"
	_, _ = d.AddVertex(TestVertex{VertexID: rootID, Name: "Root"})
	nodesPerLevel[0] = []string{rootID}

	count := 1
	level := 1

	for count < nodeCount && level < depth {
		var nextLevel []string
		nodeCounter := 0

		for _, parentID := range nodesPerLevel[level-1] {
			for b := 0; b < branches && count < nodeCount; b++ {
				id := "node_" + strconv.Itoa(level) + "_" + strconv.Itoa(nodeCounter)
				_, _ = d.AddVertex(TestVertex{VertexID: id, Name: "Node" + strconv.Itoa(level) + "_" + strconv.Itoa(nodeCounter)})
				_ = d.AddEdge(parentID, id)
				nextLevel = append(nextLevel, id)
				nodeCounter++
				count++
			}
		}

		if len(nextLevel) > 0 {
			nodesPerLevel[level] = nextLevel
		}
		level++
	}

	return d
}