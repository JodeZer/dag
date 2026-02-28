package dag

import (
	"encoding/json"
	"fmt"
)

// TypedDAG is a type-safe directed acyclic graph with vertex values of type T.
// It provides compile-time type checking for vertex values and eliminates the need
// for type assertions when working with vertices.
//
// Example usage:
//
//	type Person struct {
//	    Name string `json:"name"`
//	    Age  int    `json:"age"`
//	}
//
//	dag := dag.New[Person]()
//	dag.AddVertexByID("p1", Person{Name: "Alice", Age: 30})
//	dag.AddVertexByID("p2", Person{Name: "Bob", Age: 25})
//	dag.AddEdge("p1", "p2")
//
//	// Type-safe vertex access
//	person, err := dag.GetVertex("p1")
//	if err == nil {
//	    fmt.Printf("%s is %d years old\n", person.Name, person.Age)
//	}
//
//	// Auto-infer serialization - no generic parameter needed
//	data, err := dag.MarshalJSON()
//
//	// Deserialize with type parameter (reasonable - need to know the type)
//	restored, err := dag.UnmarshalJSON[Person](data, dag.Options{})
type TypedDAG[T any] struct {
	inner *DAG
}

// New creates a new type-safe DAG with vertex values of type T.
func New[T any]() *TypedDAG[T] {
	return &TypedDAG[T]{
		inner: NewDAG(),
	}
}

// NewWithOptions creates a new type-safe DAG with vertex values of type T
// and custom options.
func NewWithOptions[T any](options Options) *TypedDAG[T] {
	dag := &TypedDAG[T]{
		inner: NewDAG(),
	}
	dag.inner.Options(options)
	return dag
}

// AddVertex adds the vertex v to the DAG.
// AddVertex returns the generated id and an error if v is nil or already part of the graph.
func (d *TypedDAG[T]) AddVertex(v T) (string, error) {
	return d.inner.AddVertex(v)
}

// AddVertexByID adds the vertex v and the specified id to the DAG.
// AddVertexByID returns an error if v is nil, v is already part of the graph,
// or the specified id is already part of the graph.
func (d *TypedDAG[T]) AddVertexByID(id string, v T) error {
	return d.inner.AddVertexByID(id, v)
}

// GetVertex returns a vertex by its id.
// GetVertex returns an error if id is empty or unknown.
func (d *TypedDAG[T]) GetVertex(id string) (T, error) {
	v, err := d.inner.GetVertex(id)
	if err != nil {
		var zero T
		return zero, err
	}
	typed, ok := v.(T)
	if !ok {
		var zero T
		return zero, fmt.Errorf("vertex %s is not of expected type %T", id, zero)
	}
	return typed, nil
}

// GetVertices returns all vertices as a map of id to value.
func (d *TypedDAG[T]) GetVertices() map[string]T {
	result := make(map[string]T)
	vertices := d.inner.GetVertices()
	for id, v := range vertices {
		if typed, ok := v.(T); ok {
			result[id] = typed
		}
	}
	return result
}

// DeleteVertex deletes the vertex with the given id.
// DeleteVertex also deletes all attached edges (inbound and outbound).
// DeleteVertex returns an error if id is empty or unknown.
func (d *TypedDAG[T]) DeleteVertex(id string) error {
	return d.inner.DeleteVertex(id)
}

// AddEdge adds an edge between srcID and dstID.
// AddEdge returns an error if srcID or dstID are empty strings or unknown,
// if the edge already exists, or if the new edge would create a loop.
func (d *TypedDAG[T]) AddEdge(srcID, dstID string) error {
	return d.inner.AddEdge(srcID, dstID)
}

// IsEdge returns true if there exists an edge between srcID and dstID.
// IsEdge returns false if there is no such edge.
// IsEdge returns an error if srcID or dstID are empty, unknown, or the same.
func (d *TypedDAG[T]) IsEdge(srcID, dstID string) (bool, error) {
	return d.inner.IsEdge(srcID, dstID)
}

// DeleteEdge deletes the edge between srcID and dstID.
// DeleteEdge returns an error if srcID or dstID are empty or unknown,
// or if there is no edge between srcID and dstID.
func (d *TypedDAG[T]) DeleteEdge(srcID, dstID string) error {
	return d.inner.DeleteEdge(srcID, dstID)
}

// GetOrder returns the number of vertices in the graph.
func (d *TypedDAG[T]) GetOrder() int {
	return d.inner.GetOrder()
}

// GetSize returns the number of edges in the graph.
func (d *TypedDAG[T]) GetSize() int {
	return d.inner.GetSize()
}

// IsEmpty returns true if the graph has no vertices.
func (d *TypedDAG[T]) IsEmpty() bool {
	return d.inner.GetOrder() == 0
}

// GetLeaves returns all vertices without children.
func (d *TypedDAG[T]) GetLeaves() map[string]T {
	leaves := d.inner.GetLeaves()
	result := make(map[string]T)
	for id, v := range leaves {
		if typed, ok := v.(T); ok {
			result[id] = typed
		}
	}
	return result
}

// IsLeaf returns true if the vertex with the given id has no children.
// IsLeaf returns an error if id is empty or unknown.
func (d *TypedDAG[T]) IsLeaf(id string) (bool, error) {
	return d.inner.IsLeaf(id)
}

// GetRoots returns all vertices without parents.
func (d *TypedDAG[T]) GetRoots() map[string]T {
	roots := d.inner.GetRoots()
	result := make(map[string]T)
	for id, v := range roots {
		if typed, ok := v.(T); ok {
			result[id] = typed
		}
	}
	return result
}

// IsRoot returns true if the vertex with the given id has no parents.
// IsRoot returns an error if id is empty or unknown.
func (d *TypedDAG[T]) IsRoot(id string) (bool, error) {
	return d.inner.IsRoot(id)
}

// GetParents returns all parents of the vertex with the id.
// GetParents returns an error if id is empty or unknown.
func (d *TypedDAG[T]) GetParents(id string) (map[string]T, error) {
	parents, err := d.inner.GetParents(id)
	if err != nil {
		return nil, err
	}
	result := make(map[string]T)
	for id, v := range parents {
		if typed, ok := v.(T); ok {
			result[id] = typed
		}
	}
	return result, nil
}

// GetChildren returns all children of the vertex with the id.
// GetChildren returns an error if id is empty or unknown.
func (d *TypedDAG[T]) GetChildren(id string) (map[string]T, error) {
	children, err := d.inner.GetChildren(id)
	if err != nil {
		return nil, err
	}
	result := make(map[string]T)
	for id, v := range children {
		if typed, ok := v.(T); ok {
			result[id] = typed
		}
	}
	return result, nil
}

// GetAncestors returns all ancestors of the vertex with the id.
// GetAncestors returns an error if id is empty or unknown.
func (d *TypedDAG[T]) GetAncestors(id string) (map[string]T, error) {
	ancestors, err := d.inner.GetAncestors(id)
	if err != nil {
		return nil, err
	}
	result := make(map[string]T)
	for id, v := range ancestors {
		if typed, ok := v.(T); ok {
			result[id] = typed
		}
	}
	return result, nil
}

// GetOrderedAncestors returns all ancestors of the vertex with id
// in a breath-first order. Only the first occurrence of each vertex is returned.
// GetOrderedAncestors returns an error if id is empty or unknown.
func (d *TypedDAG[T]) GetOrderedAncestors(id string) ([]string, error) {
	return d.inner.GetOrderedAncestors(id)
}

// GetDescendants returns all descendants of the vertex with the id.
// GetDescendants returns an error if id is empty or unknown.
func (d *TypedDAG[T]) GetDescendants(id string) (map[string]T, error) {
	descendants, err := d.inner.GetDescendants(id)
	if err != nil {
		return nil, err
	}
	result := make(map[string]T)
	for id, v := range descendants {
		if typed, ok := v.(T); ok {
			result[id] = typed
		}
	}
	return result, nil
}

// GetOrderedDescendants returns all descendants of the vertex with id
// in a breath-first order. Only the first occurrence of each vertex is returned.
// GetOrderedDescendants returns an error if id is empty or unknown.
func (d *TypedDAG[T]) GetOrderedDescendants(id string) ([]string, error) {
	return d.inner.GetOrderedDescendants(id)
}

// GetDescendantsGraph returns a new TypedDAG consisting of the vertex with id
// and all its descendants (i.e. the subgraph). GetDescendantsGraph also returns
// the id of the (copy of the) given vertex within the new graph (i.e. the id of
// the single root of the new graph). GetDescendantsGraph returns an error if id
// is empty or unknown.
func (d *TypedDAG[T]) GetDescendantsGraph(id string) (*TypedDAG[T], string, error) {
	inner, newId, err := d.inner.GetDescendantsGraph(id)
	if err != nil {
		return nil, "", err
	}
	return &TypedDAG[T]{inner: inner}, newId, nil
}

// GetAncestorsGraph returns a new TypedDAG consisting of the vertex with id
// and all its ancestors (i.e. the subgraph). GetAncestorsGraph also returns
// the id of the (copy of the) given vertex within the new graph (i.e. the id of
// the single leaf of the new graph). GetAncestorsGraph returns an error if id
// is empty or unknown.
func (d *TypedDAG[T]) GetAncestorsGraph(id string) (*TypedDAG[T], string, error) {
	inner, newId, err := d.inner.GetAncestorsGraph(id)
	if err != nil {
		return nil, "", err
	}
	return &TypedDAG[T]{inner: inner}, newId, nil
}

// AncestorsWalker returns a channel and subsequently walks all ancestors of
// the vertex with id in a breath first order. The second channel returned may
// be used to stop further walking. AncestorsWalker returns an error if id is
// empty or unknown.
func (d *TypedDAG[T]) AncestorsWalker(id string) (chan string, chan bool, error) {
	return d.inner.AncestorsWalker(id)
}

// DescendantsWalker returns a channel and subsequently walks all descendants
// of the vertex with id in a breath first order. The second channel returned
// may be used to stop further walking. DescendantsWalker returns an error if
// id is empty or unknown.
func (d *TypedDAG[T]) DescendantsWalker(id string) (chan string, chan bool, error) {
	return d.inner.DescendantsWalker(id)
}

// DescendantsFlow traverses descendants of the vertex with the ID startID.
// For the vertex itself and each of its descendant it executes the given
// callback function providing it the results of its respective parents.
// The callback function is only executed after all parents have finished their work.
func (d *TypedDAG[T]) DescendantsFlow(startID string, inputs []FlowResult, callback FlowCallback) ([]FlowResult, error) {
	return d.inner.DescendantsFlow(startID, inputs, callback)
}

// ReduceTransitively transitively reduces the graph.
func (d *TypedDAG[T]) ReduceTransitively() {
	d.inner.ReduceTransitively()
}

// FlushCaches completely flushes the descendants- and ancestor cache.
func (d *TypedDAG[T]) FlushCaches() {
	d.inner.FlushCaches()
}

// Copy returns a copy of the TypedDAG.
func (d *TypedDAG[T]) Copy() (*TypedDAG[T], error) {
	inner, err := d.inner.Copy()
	if err != nil {
		return nil, err
	}
	return &TypedDAG[T]{inner: inner}, nil
}

// MarshalJSON returns the JSON encoding of the TypedDAG.
// This method automatically infers the vertex type T from the TypedDAG,
// so no generic parameter is needed.
func (d *TypedDAG[T]) MarshalJSON() ([]byte, error) {
	return MarshalGeneric[T](d.inner)
}

// Options sets the options for the TypedDAG.
// Options must be called before any other method of the TypedDAG is called.
func (d *TypedDAG[T]) Options(options Options) {
	d.inner.Options(options)
}

// UnmarshalJSON parses JSON-encoded data and returns a new TypedDAG[T].
// This is the recommended function for unmarshaling TypedDAGs from JSON.
//
// The generic parameter T specifies the type of vertex values. It can be any type
// that json.Unmarshal can handle, including complex nested structs.
//
// Example usage:
//
//	// Simple type
//	dag, err := dag.UnmarshalJSON[string](data, dag.Options{})
//
//	// Complex custom type
//	type Person struct {
//	    Name string `json:"name"`
//	    Age  int    `json:"age"`
//	}
//	dag, err := dag.UnmarshalJSON[Person](data, dag.Options{})
func UnmarshalJSON[T any](data []byte, options Options) (*TypedDAG[T], error) {
	var sd storableDAGGeneric[T]
	if err := json.Unmarshal(data, &sd); err != nil {
		return nil, err
	}

	dag := NewDAG()

	// Set options only if VertexHashFunc is provided
	if options.VertexHashFunc != nil {
		dag.Options(options)
	}

	// Batch add vertices - direct access to avoid interface boxing
	dag.muDAG.Lock()
	for _, v := range sd.VerticesGeneric() {
		id := v.WrappedID
		value := v.Value

		vHash := dag.hashVertex(value)

		// Check for duplicate vertex
		if _, exists := dag.vertices[vHash]; exists {
			dag.muDAG.Unlock()
			return nil, VertexDuplicateError{value}
		}

		if _, exists := dag.vertexIds[id]; exists {
			dag.muDAG.Unlock()
			return nil, IDDuplicateError{id}
		}

		// Add vertex directly without interface boxing
		dag.vertices[vHash] = id
		dag.vertexIds[id] = value
	}
	dag.muDAG.Unlock()

	// Batch add edges using optimized method
	if len(sd.StorableEdges) > 0 {
		if err := dag.addEdgesBatch(sd.StorableEdges); err != nil {
			return nil, err
		}
	}

	return &TypedDAG[T]{inner: dag}, nil
}

// ToDAG returns the underlying *DAG for compatibility with existing code.
// This is a convenience method for migrating from the non-typed API.
func (d *TypedDAG[T]) ToDAG() *DAG {
	return d.inner
}