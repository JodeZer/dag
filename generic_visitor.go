package dag

// GenericVisitor is the interface for visiting generic DAG vertices.
type GenericVisitor[T any] interface {
	Visit(value T, id string)
}

// GenericDFSVisitor implements the DFS traversal for GenericDAG.
type GenericDFSVisitor[T any] struct {
	visitor GenericVisitor[T]
}

// NewGenericDFSVisitor creates a new DFS visitor.
func NewGenericDFSVisitor[T any](visitor GenericVisitor[T]) *GenericDFSVisitor[T] {
	return &GenericDFSVisitor[T]{visitor: visitor}
}

// Visit implements the Visitor interface for DFS walk compatibility.
func (gv *GenericDFSVisitor[T]) Visit(v Vertexer) {
	// This method is used when the visitor is passed to the original DAG walk
	// It's not intended for use with GenericDAG
}

// GenericDFSWalk implements the Depth-First-Search algorithm to traverse the entire GenericDAG.
// The algorithm starts at the root node and explores as far as possible
// along each branch before backtracking.
func (d *GenericDAG[T]) GenericDFSWalk(visitor GenericVisitor[T]) {
	d.muDAG.RLock()
	defer d.muDAG.RUnlock()

	// Use native slice as stack for better performance
	stack := make([]string, 0, d.GetSize())

	vertices := d.getRoots()
	// Push roots in reverse order to maintain consistent traversal order
	ids := vertexIDsGeneric(vertices)
	for i := len(ids) - 1; i >= 0; i-- {
		id := ids[i]
		stack = append(stack, id)
	}

	visited := make(map[string]bool, d.getSize())

	for len(stack) > 0 {
		// Pop from stack
		idx := len(stack) - 1
		id := stack[idx]
		stack = stack[:idx]

		if !visited[id] {
			visited[id] = true
			visitor.Visit(d.vertexValues[id], id)
		}

		children, _ := d.getChildren(id)
		childIDs := vertexIDsGeneric(children)
		for i := len(childIDs) - 1; i >= 0; i-- {
			childID := childIDs[i]
			if !visited[childID] {
				stack = append(stack, childID)
			}
		}
	}
}

// GenericBFSWalk implements the Breadth-First-Search algorithm to traverse the entire GenericDAG.
// It starts at the tree root and explores all nodes at the present depth prior
// to moving on to the nodes at the next depth level.
func (d *GenericDAG[T]) GenericBFSWalk(visitor GenericVisitor[T]) {
	d.muDAG.RLock()
	defer d.muDAG.RUnlock()

	// Use native slice as queue for better performance
	queue := make([]string, 0, d.GetSize())

	vertices := d.getRoots()
	ids := vertexIDsGeneric(vertices)
	queue = append(queue, ids...)

	visited := make(map[string]bool, d.getOrder())

	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]

		if !visited[id] {
			visited[id] = true
			visitor.Visit(d.vertexValues[id], id)
		}

		children, _ := d.getChildren(id)
		childIDs := vertexIDsGeneric(children)
		for _, childID := range childIDs {
			if !visited[childID] {
				queue = append(queue, childID)
			}
		}
	}
}

// GenericOrderedWalk implements the Topological Sort algorithm to traverse the entire GenericDAG.
// This means that for any edge a -> b, node a will be visited before node b.
func (d *GenericDAG[T]) GenericOrderedWalk(visitor GenericVisitor[T]) {
	d.muDAG.RLock()
	defer d.muDAG.RUnlock()

	queue := make([]string, 0, d.GetSize())
	vertices := d.getRoots()
	ids := vertexIDsGeneric(vertices)
	queue = append(queue, ids...)

	visited := make(map[string]bool, d.getOrder())

	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]

		if visited[id] {
			continue
		}

		// if the current vertex has any parent that hasn't been visited yet,
		// put it back into the queue, and work on the next element
		parents, _ := d.GetParents(id)
		hasUnvisitedParent := false
		for parent := range parents {
			if !visited[parent] {
				queue = append(queue, id)
				hasUnvisitedParent = true
				break
			}
		}
		if hasUnvisitedParent {
			continue
		}

		if !visited[id] {
			visited[id] = true
			visitor.Visit(d.vertexValues[id], id)
		}

		children, _ := d.getChildren(id)
		childIDs := vertexIDsGeneric(children)
		for _, childID := range childIDs {
			if !visited[childID] {
				queue = append(queue, childID)
			}
		}
	}
}

func vertexIDsGeneric[T any](vertices map[string]T) []string {
	ids := make([]string, 0, len(vertices))
	for id := range vertices {
		ids = append(ids, id)
	}
	return ids
}