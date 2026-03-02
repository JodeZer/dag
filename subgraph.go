package dag

// EdgeList represents a list of edges in the DAG.
type EdgeList struct {
	Edges []GenericEdge
}

// NewEdgeList creates a new EdgeList with the given capacity.
func NewEdgeList(capacity int) *EdgeList {
	return &EdgeList{
		Edges: make([]GenericEdge, 0, capacity),
	}
}

// AddEdge adds an edge to the list.
func (el *EdgeList) AddEdge(srcID, dstID string) {
	el.Edges = append(el.Edges, GenericEdge{
		SrcID: srcID,
		DstID: dstID,
	})
}

// Copy creates a deep copy of the EdgeList.
func (el *EdgeList) Copy() *EdgeList {
	copied := NewEdgeList(len(el.Edges))
	copied.Edges = append(copied.Edges, el.Edges...)
	return copied
}

// Count returns the number of edges in the list.
func (el *EdgeList) Count() int {
	return len(el.Edges)
}

// NodeList represents a list of vertices in the DAG.
type NodeList[T any] struct {
	Nodes []GenericStorableVertex[T]
}

// NewNodeList creates a new NodeList with the given capacity.
func NewNodeList[T any](capacity int) *NodeList[T] {
	return &NodeList[T]{
		Nodes: make([]GenericStorableVertex[T], 0, capacity),
	}
}

// AddNode adds a node to the list.
func (nl *NodeList[T]) AddNode(id string, value T) {
	nl.Nodes = append(nl.Nodes, GenericStorableVertex[T]{
		ID:    id,
		Value: value,
	})
}

// Copy creates a deep copy of the NodeList.
// Note: This creates a shallow copy of the vertex values.
func (nl *NodeList[T]) Copy() *NodeList[T] {
	copied := NewNodeList[T](len(nl.Nodes))
	copied.Nodes = append(copied.Nodes, nl.Nodes...)
	return copied
}

// Count returns the number of nodes in the list.
func (nl *NodeList[T]) Count() int {
	return len(nl.Nodes)
}

// CopyOption defines data copying options for edge and node list retrieval.
type CopyOption int

const (
	// ShareData indicates that the returned data should share references
	// with the source DAG. This is more efficient but modifications to the
	// shared data may affect the DAG.
	ShareData CopyOption = iota

	// CopyData indicates that the returned data should be a deep copy.
	// This is safer but uses more memory.
	CopyData
)