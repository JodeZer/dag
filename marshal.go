package dag

import (
	"encoding/json"
	"errors"
)

// MarshalJSON returns the JSON encoding of DAG.
//
// It traverses the DAG using the Depth-First-Search algorithm
// and uses an internal structure to store vertices and edges.
//
// Deprecated: Use MarshalGeneric[T] for better performance with typed data.
func (d *DAG) MarshalJSON() ([]byte, error) {
	mv := newMarshalVisitor(d)
	d.DFSWalk(mv)
	return json.Marshal(mv.storableDAG)
}

// MarshalGeneric returns the JSON encoding of DAG with typed vertex values.
//
// The generic parameter T specifies the type of vertex values.
// This is the recommended method for serialization when using the generic API.
//
// Example usage:
//
//   // Simple type
//   data, err := dag.MarshalGeneric[string](d)
//
//   // Complex custom type
//   type Person struct { Name string; Age int }
//   data, err := dag.MarshalGeneric[Person](d)
func MarshalGeneric[T any](d *DAG) ([]byte, error) {
	mv := newGenericMarshalVisitor[T](d)
	d.DFSWalk(mv)
	return json.Marshal(mv.storableDAGGeneric)
}

// UnmarshalJSON is an informative method. See the UnmarshalJSON function below.
func (d *DAG) UnmarshalJSON(_ []byte) error {
	return errors.New("this method is not supported, request function UnmarshalJSON instead")
}

// UnmarshalJSONGeneric parses JSON-encoded data and returns a new DAG.
// This is the legacy generic function kept for backward compatibility.
//
// The generic parameter T specifies the type of vertex values. It can be any type
// that json.Unmarshal can handle, including complex nested structs.
//
// For new code, consider using the TypedDAG[T] API with UnmarshalJSON[T] instead.
//
// Example usage:
//
//   // Simple type
//   dag, err := dag.UnmarshalJSONGeneric[string](data, opts)
//
//   // Complex custom type
//   type Person struct {
//       Name string `json:"name"`
//       Age  int    `json:"age"`
//   }
//   dag, err := dag.UnmarshalJSONGeneric[Person](data, opts)
//
//   // Pointer to struct type
//   dag, err := dag.UnmarshalJSONGeneric[*Person](data, opts)
func UnmarshalJSONGeneric[T any](data []byte, options Options) (*DAG, error) {
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

	return dag, nil
}

// UnmarshalJSONLegacy parses the JSON-encoded data that defined by StorableDAG.
// This is the legacy function kept for backward compatibility.
// It returns a new DAG defined by the vertices and edges of wd.
// If the internal structure of data and wd do not match,
// then deserialization will fail and return json error.
//
// Because the vertex data passed in by the user is an interface{},
// it does not indicate a specific structure, so it cannot be deserialized.
// And this function needs to pass in a clear DAG structure.
//
// Example:
// dag := NewDAG()
// data, err := json.Marshal(d)
// if err != nil {
//     panic(err)
// }
// var wd YourStorableDAG
// restoredDag, err := UnmarshalJSONLegacy(data, &wd)
// if err != nil {
//     panic(err)
// }
//
// For more specific information please read the test code.
//
// Deprecated: Use the generic UnmarshalJSON[T] function instead.
func UnmarshalJSONLegacy(data []byte, wd StorableDAG, options Options) (*DAG, error) {
	err := json.Unmarshal(data, &wd)
	if err != nil {
		return nil, err
	}
	dag := NewDAG()
	dag.Options(options)

	// Use batch vertex addition for better performance
	vertices := wd.Vertices()
	if len(vertices) > 0 {
		if err := dag.addVerticesBatch(vertices); err != nil {
			return nil, err
		}
	}

	for _, e := range wd.Edges() {
		errEdge := dag.AddEdge(e.Edge())
		if errEdge != nil {
			return nil, errEdge
		}
	}
	return dag, nil
}

type marshalVisitor struct {
	d *DAG
	storableDAG
}

func newMarshalVisitor(d *DAG) *marshalVisitor {
	// Pre-allocate memory based on expected graph size
	// This reduces reallocations during the walk
	order := d.GetOrder()
	size := d.GetSize()
	return &marshalVisitor{
		d: d,
		storableDAG: storableDAG{
			StorableVertices: make([]Vertexer, 0, order),
			StorableEdges:    make([]Edger, 0, size),
		},
	}
}

func (mv *marshalVisitor) Visit(v Vertexer) {
	mv.StorableVertices = append(mv.StorableVertices, v)

	srcID, _ := v.Vertex()
	// Why not use Mutex here?
	// Because at the time of Walk,
	// the read lock has been used to protect the dag.
	children, _ := mv.d.getChildren(srcID)
	// Directly iterate over map keys - no need to sort for serialization
	for dstID := range children {
		e := storableEdge{SrcID: srcID, DstID: dstID}
		mv.StorableEdges = append(mv.StorableEdges, e)
	}
}

// genericMarshalVisitor is a visitor that collects vertices and edges for generic serialization.
type genericMarshalVisitor[T any] struct {
	d                  *DAG
	storableDAGGeneric storableDAGGeneric[T]
}

func newGenericMarshalVisitor[T any](d *DAG) *genericMarshalVisitor[T] {
	// Pre-allocate memory based on expected graph size
	// This reduces reallocations during the walk
	order := d.GetOrder()
	size := d.GetSize()
	return &genericMarshalVisitor[T]{
		d: d,
		storableDAGGeneric: storableDAGGeneric[T]{
			StorableVertices: make([]storableVertexGeneric[T], 0, order),
			StorableEdges:    make([]storableEdge, 0, size),
		},
	}
}

func (mv *genericMarshalVisitor[T]) Visit(v Vertexer) {
	// Extract vertex ID and value
	id, value := v.Vertex()

	// Convert value to type T
	var typedValue T
	if value != nil {
		// Try type assertion first
		if typed, ok := value.(T); ok {
			typedValue = typed
		} else {
			// Fall back to JSON marshaling/unmarshaling for type conversion
			valueJSON, err := json.Marshal(value)
			if err != nil {
				return
			}
			if err := json.Unmarshal(valueJSON, &typedValue); err != nil {
				return
			}
		}
	}

	// Add vertex to storable DAG
	mv.storableDAGGeneric.StorableVertices = append(mv.storableDAGGeneric.StorableVertices, storableVertexGeneric[T]{
		WrappedID: id,
		Value:     typedValue,
	})

	// Add edges
	// Why not use Mutex here?
	// Because at the time of Walk,
	// the read lock has been used to protect the dag.
	children, _ := mv.d.getChildren(id)
	// Directly iterate over map keys - no need to sort for serialization
	for dstID := range children {
		e := storableEdge{SrcID: id, DstID: dstID}
		mv.storableDAGGeneric.StorableEdges = append(mv.storableDAGGeneric.StorableEdges, e)
	}
}