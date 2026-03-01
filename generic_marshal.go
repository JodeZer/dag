package dag

import "encoding/json"

// GenericStorableVertex represents a vertex for serialization.
type GenericStorableVertex[T any] struct {
	ID    string `json:"i"`
	Value T      `json:"v"`
}

// GenericStorableDAG represents a DAG for serialization.
type GenericStorableDAG[T any] struct {
	Vertices []GenericStorableVertex[T] `json:"vs"`
	Edges    []GenericEdge              `json:"es"`
}

// GenericEdge represents an edge for serialization.
type GenericEdge struct {
	SrcID string `json:"s"`
	DstID string `json:"d"`
}

// GenericMarshalVisitor implements GenericVisitor for marshaling.
type GenericMarshalVisitor[T any] struct {
	vertices []GenericStorableVertex[T]
	edges    []GenericEdge
	visited  map[string]bool
}

// NewGenericMarshalVisitor creates a new marshal visitor for GenericDAG.
func NewGenericMarshalVisitor[T any](order, size int) *GenericMarshalVisitor[T] {
	return &GenericMarshalVisitor[T]{
		vertices: make([]GenericStorableVertex[T], 0, order),
		edges:    make([]GenericEdge, 0, size),
		visited:  make(map[string]bool),
	}
}

// Visit implements GenericVisitor interface.
func (mv *GenericMarshalVisitor[T]) Visit(value T, id string) {
	if !mv.visited[id] {
		mv.visited[id] = true
		mv.vertices = append(mv.vertices, GenericStorableVertex[T]{
			ID:    id,
			Value: value,
		})
	}
}

// AddEdges adds edges from a parent to its children.
func (mv *GenericMarshalVisitor[T]) AddEdges(parentID string, children map[string]interface{}) {
	for childID := range children {
		mv.edges = append(mv.edges, GenericEdge{
			SrcID: parentID,
			DstID: childID,
		})
	}
}

// MarshalJSON returns the JSON encoding of the GenericDAG.
func (d *GenericDAG[T]) MarshalJSON() ([]byte, error) {
	d.muDAG.RLock()
	defer d.muDAG.RUnlock()

	order := d.getOrder()
	size := d.getSize()
	visitor := NewGenericMarshalVisitor[T](order, size)

	// DFS walk to collect vertices and edges
	stack := make([]string, 0, size)
	vertices := d.getRoots()
	ids := vertexIDsGeneric(vertices)
	for i := len(ids) - 1; i >= 0; i-- {
		stack = append(stack, ids[i])
	}

	visited := make(map[string]bool, order)

	for len(stack) > 0 {
		idx := len(stack) - 1
		id := stack[idx]
		stack = stack[:idx]

		if !visited[id] {
			visited[id] = true
			visitor.Visit(d.vertexValues[id], id)
		}

		children, _ := d.getChildren(id)
		visitor.AddEdges(id, convertToInterfaceMap(children))
		childIDs := vertexIDsGeneric(children)
		for i := len(childIDs) - 1; i >= 0; i-- {
			childID := childIDs[i]
			if !visited[childID] {
				stack = append(stack, childID)
			}
		}
	}

	dag := GenericStorableDAG[T]{
		Vertices: visitor.vertices,
		Edges:    visitor.edges,
	}
	return json.Marshal(dag)
}

// UnmarshalGenericJSON parses JSON-encoded data and returns a new GenericDAG.
// This is the recommended function for unmarshaling GenericDAGs from JSON.
//
// The generic parameter T specifies the type of vertex values. It can be any type
// that json.Unmarshal can handle, including complex nested structs.
//
// Example usage:
//
//	// Simple type
//	dag, err := dag.UnmarshalGenericJSON[string](data, dag.Options{})
//
//	// Complex custom type
//	type Person struct {
//	    Name string `json:"name"`
//	    Age  int    `json:"age"`
//	}
//	dag, err := dag.UnmarshalGenericJSON[Person](data, dag.Options{})
func UnmarshalGenericJSON[T any](data []byte, options Options) (*GenericDAG[T], error) {
	var dag GenericStorableDAG[T]
	if err := json.Unmarshal(data, &dag); err != nil {
		return nil, err
	}

	g := NewGenericDAG[T]()
	if options.VertexHashFunc != nil {
		g.Options(options)
	}

	// Add all vertices
	for _, v := range dag.Vertices {
		if err := g.AddVertexByID(v.ID, v.Value); err != nil {
			return nil, err
		}
	}

	// Add all edges
	for _, e := range dag.Edges {
		if err := g.AddEdge(e.SrcID, e.DstID); err != nil {
			return nil, err
		}
	}

	return g, nil
}

// convertToInterfaceMap is a helper to convert map[string]T to map[string]interface{}
// for compatibility with AddEdges method.
func convertToInterfaceMap[T any](m map[string]T) map[string]interface{} {
	result := make(map[string]interface{}, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}