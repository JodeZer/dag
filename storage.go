package dag

var (
	_ Vertexer    = (*storableVertex)(nil)
	_ Edger       = (*storableEdge)(nil)
	_ StorableDAG = (*storableDAG)(nil)
	_ IDInterface = (*storableVertex)(nil)
	_ StorableDAG = (*storableDAGGeneric[int])(nil)
)

// Vertexer is the interface that wraps the basic Vertex method.
// Vertex returns an id that identifies this vertex and the value of this vertex.
//
// The reason for defining this new structure is that the vertex id may be
// automatically generated when the caller adds a vertex. At this time, the
// vertex structure added by the user does not contain id information.
type Vertexer interface {
	Vertex() (id string, value interface{})
}

// Edger is the interface that wraps the basic Edge method.
// Edge returns the ids of two vertices that connect an edge.
type Edger interface {
	Edge() (srcID, dstID string)
}

// StorableDAG is the interface that defines a DAG that can be stored.
// It provides methods to get all vertices and all edges of a DAG.
type StorableDAG interface {
	Vertices() []Vertexer
	Edges() []Edger
}

// storableVertex implements the Vertexer interface.
// It is implemented as a storable structure.
// And it uses short json tag to reduce the number of bytes after serialization.
type storableVertex struct {
	WrappedID string      `json:"i"`
	Value     interface{} `json:"v"`
}

func (v storableVertex) Vertex() (id string, value interface{}) {
	return v.WrappedID, v.Value
}

func (v storableVertex) ID() string {
	return v.WrappedID
}

// storableVertexGeneric implements the Vertexer interface for generic types.
// It uses short json tag to reduce the number of bytes after serialization.
type storableVertexGeneric[T any] struct {
	WrappedID string `json:"i"`
	Value     T      `json:"v"`
}

func (v storableVertexGeneric[T]) Vertex() (id string, value interface{}) {
	return v.WrappedID, v.Value
}

func (v storableVertexGeneric[T]) ID() string {
	return v.WrappedID
}

// storableEdge implements the Edger interface.
// It is implemented as a storable structure.
// And it uses short json tag to reduce the number of bytes after serialization.
type storableEdge struct {
	SrcID string `json:"s"`
	DstID string `json:"d"`
}

func (e storableEdge) Edge() (srcID, dstID string) {
	return e.SrcID, e.DstID
}

// storableDAG implements the StorableDAG interface.
// It acts as a serializable operable structure.
// And it uses short json tag to reduce the number of bytes after serialization.
type storableDAG struct {
	StorableVertices []Vertexer `json:"vs"`
	StorableEdges    []Edger    `json:"es"`
}

func (g storableDAG) Vertices() []Vertexer {
	return g.StorableVertices
}

func (g storableDAG) Edges() []Edger {
	return g.StorableEdges
}

// storableDAGGeneric implements the StorableDAG interface for generic types.
// It acts as a serializable operable structure.
// And it uses short json tag to reduce the number of bytes after serialization.
type storableDAGGeneric[T any] struct {
	StorableVertices []storableVertexGeneric[T] `json:"vs"`
	StorableEdges    []storableEdge              `json:"es"`
}

func (g storableDAGGeneric[T]) Vertices() []Vertexer {
	vertices := make([]Vertexer, 0, len(g.StorableVertices))
	for _, v := range g.StorableVertices {
		vertices = append(vertices, v)
	}
	return vertices
}

// VerticesGeneric returns the vertices as their generic type T directly,
// avoiding interface conversion overhead.
func (g storableDAGGeneric[T]) VerticesGeneric() []storableVertexGeneric[T] {
	return g.StorableVertices
}

func (g storableDAGGeneric[T]) Edges() []Edger {
	edges := make([]Edger, 0, len(g.StorableEdges))
	for _, e := range g.StorableEdges {
		edges = append(edges, e)
	}
	return edges
}
