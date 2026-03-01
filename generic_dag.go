package dag

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
)

// GenericDAG implements the data structure of the DAG with typed vertex values.
// This is the new generic implementation that eliminates type conversion overhead
// by storing vertex values directly as type T instead of interface{}.
//
// Example usage:
//
//	type Person struct {
//	    Name string `json:"name"`
//	    Age  int    `json:"age"`
//	}
//
//	// Create a DAG with string vertices
//	dag := NewGenericDAG[string]()
//	dag.AddVertex("value")
//
//	// Create a DAG with custom type vertices
//	personDAG := NewGenericDAG[Person]()
//	personDAG.AddVertex(Person{Name: "Alice", Age: 30})
type GenericDAG[T any] struct {
	muDAG            sync.RWMutex
	vertices         map[interface{}]string
	vertexValues     map[string]T
	inboundEdge      map[interface{}]map[interface{}]struct{}
	outboundEdge     map[interface{}]map[interface{}]struct{}
	muCache          sync.RWMutex
	verticesLocked   *dMutex
	ancestorsCache   map[interface{}]map[interface{}]struct{}
	descendantsCache map[interface{}]map[interface{}]struct{}
	options          Options
}

// NewGenericDAG creates / initializes a new generic DAG.
func NewGenericDAG[T any]() *GenericDAG[T] {
	return &GenericDAG[T]{
		vertices:         make(map[interface{}]string),
		vertexValues:     make(map[string]T),
		inboundEdge:      make(map[interface{}]map[interface{}]struct{}),
		outboundEdge:     make(map[interface{}]map[interface{}]struct{}),
		verticesLocked:   newDMutex(),
		ancestorsCache:   make(map[interface{}]map[interface{}]struct{}),
		descendantsCache: make(map[interface{}]map[interface{}]struct{}),
		options:          defaultOptions(),
	}
}

// AddVertex adds the vertex v to the DAG.
// AddVertex returns the generated id and an error if v is already part of the graph.
func (d *GenericDAG[T]) AddVertex(v T) (string, error) {
	d.muDAG.Lock()
	defer d.muDAG.Unlock()
	return d.addVertex(v)
}

func (d *GenericDAG[T]) addVertex(v T) (string, error) {
	var id string
	// Use interface{} for IDInterface check
	if i, ok := any(v).(IDInterface); ok {
		id = i.ID()
	} else {
		id = uuid.New().String()
	}

	err := d.addVertexByID(id, v)
	return id, err
}

// AddVertexByID adds the vertex v and the specified id to the DAG.
// AddVertexByID returns an error if v is already part of the graph,
// or the specified id is already part of the graph.
func (d *GenericDAG[T]) AddVertexByID(id string, v T) error {
	d.muDAG.Lock()
	defer d.muDAG.Unlock()
	return d.addVertexByID(id, v)
}

func (d *GenericDAG[T]) addVertexByID(id string, v T) error {
	vHash := d.hashVertex(v)

	// Check for duplicate vertex
	if _, exists := d.vertices[vHash]; exists {
		return VertexDuplicateError{v}
	}

	// Check for duplicate ID
	if _, exists := d.vertexValues[id]; exists {
		return IDDuplicateError{id}
	}

	d.vertices[vHash] = id
	d.vertexValues[id] = v
	return nil
}

// GetVertex returns a vertex by its id.
// GetVertex returns an error if id is empty or unknown.
func (d *GenericDAG[T]) GetVertex(id string) (T, error) {
	d.muDAG.RLock()
	defer d.muDAG.RUnlock()

	if id == "" {
		var zero T
		return zero, IDEmptyError{}
	}

	v, exists := d.vertexValues[id]
	if !exists {
		var zero T
		return zero, IDUnknownError{id}
	}
	return v, nil
}

// DeleteVertex deletes the vertex with the given id.
// DeleteVertex also deletes all attached edges (inbound and outbound).
// DeleteVertex returns an error if id is empty or unknown.
func (d *GenericDAG[T]) DeleteVertex(id string) error {
	d.muDAG.Lock()
	defer d.muDAG.Unlock()

	if err := d.saneID(id); err != nil {
		return err
	}

	v := d.vertexValues[id]
	vHash := d.hashVertex(v)

	// get descendants and ancestors as they are now
	descendants := copyMap(d.getDescendants(vHash))
	ancestors := copyMap(d.getAncestors(vHash))

	// delete v in outbound edges of parents
	if _, exists := d.inboundEdge[vHash]; exists {
		for parent := range d.inboundEdge[vHash] {
			delete(d.outboundEdge[parent], vHash)
		}
	}

	// delete v in inbound edges of children
	if _, exists := d.outboundEdge[vHash]; exists {
		for child := range d.outboundEdge[vHash] {
			delete(d.inboundEdge[child], vHash)
		}
	}

	// delete in- and outbound of v itself
	delete(d.inboundEdge, vHash)
	delete(d.outboundEdge, vHash)

	// for v and all its descendants delete cached ancestors
	for descendant := range descendants {
		delete(d.ancestorsCache, descendant)
	}
	delete(d.ancestorsCache, vHash)

	// for v and all its ancestors delete cached descendants
	for ancestor := range ancestors {
		delete(d.descendantsCache, ancestor)
	}
	delete(d.descendantsCache, vHash)

	// delete v itself
	delete(d.vertices, vHash)
	delete(d.vertexValues, id)

	return nil
}

// AddEdge adds an edge between srcID and dstID.
// AddEdge returns an error if srcID or dstID are empty strings or unknown,
// if the edge already exists, or if the new edge would create a loop.
func (d *GenericDAG[T]) AddEdge(srcID, dstID string) error {
	d.muDAG.Lock()
	defer d.muDAG.Unlock()

	if err := d.saneID(srcID); err != nil {
		return err
	}
	if err := d.saneID(dstID); err != nil {
		return err
	}
	if srcID == dstID {
		return SrcDstEqualError{srcID, dstID}
	}

	src := d.vertexValues[srcID]
	srcHash := d.hashVertex(src)
	dst := d.vertexValues[dstID]
	dstHash := d.hashVertex(dst)

	// if the edge is already known, there is nothing else to do
	if d.isEdge(srcHash, dstHash) {
		return EdgeDuplicateError{srcID, dstID}
	}

	// check if adding src->dst would create a loop
	if d.wouldCreateLoop(srcHash, dstHash) {
		return EdgeLoopError{srcID, dstID}
	}

	// get descendants and ancestors as they are now
	descendants := copyMap(d.getDescendants(dstHash))
	ancestors := copyMap(d.getAncestors(srcHash))

	// prepare d.outbound[src], iff needed
	if _, exists := d.outboundEdge[srcHash]; !exists {
		d.outboundEdge[srcHash] = make(map[interface{}]struct{})
	}

	// dst is a child of src
	d.outboundEdge[srcHash][dstHash] = struct{}{}

	// prepare d.inboundEdge[dst], iff needed
	if _, exists := d.inboundEdge[dstHash]; !exists {
		d.inboundEdge[dstHash] = make(map[interface{}]struct{})
	}

	// src is a parent of dst
	d.inboundEdge[dstHash][srcHash] = struct{}{}

	// for dst and all its descendants delete cached ancestors
	for descendant := range descendants {
		delete(d.ancestorsCache, descendant)
	}
	delete(d.ancestorsCache, dstHash)

	// for src and all its ancestors delete cached descendants
	for ancestor := range ancestors {
		delete(d.descendantsCache, ancestor)
	}
	delete(d.descendantsCache, srcHash)

	return nil
}

// wouldCreateLoop checks if adding an edge from srcHash to dstHash would create a loop.
func (d *GenericDAG[T]) wouldCreateLoop(srcHash, dstHash interface{}) bool {
	// Use a BFS queue and visited map to search from dstHash
	var fifo []interface{}
	visited := make(map[interface{}]struct{})

	// Start with all children of dstHash
	for child := range d.outboundEdge[dstHash] {
		visited[child] = struct{}{}
		fifo = append(fifo, child)
	}

	// BFS traversal
	for {
		if len(fifo) == 0 {
			break
		}
		top := fifo[0]
		fifo = fifo[1:]

		// If we reached srcHash, adding src->dst would create a loop
		if top == srcHash {
			return true
		}

		// Add all unvisited children to the queue
		for child := range d.outboundEdge[top] {
			if _, exists := visited[child]; !exists {
				visited[child] = struct{}{}
				fifo = append(fifo, child)
			}
		}
	}

	return false
}

// IsEdge returns true if there exists an edge between srcID and dstID.
// IsEdge returns false if there is no such edge.
// IsEdge returns an error if srcID or dstID are empty, unknown, or the same.
func (d *GenericDAG[T]) IsEdge(srcID, dstID string) (bool, error) {
	d.muDAG.RLock()
	defer d.muDAG.RUnlock()

	if err := d.saneID(srcID); err != nil {
		return false, err
	}
	if err := d.saneID(dstID); err != nil {
		return false, err
	}
	if srcID == dstID {
		return false, SrcDstEqualError{srcID, dstID}
	}

	src := d.vertexValues[srcID]
	dst := d.vertexValues[dstID]
	return d.isEdge(d.hashVertex(src), d.hashVertex(dst)), nil
}

func (d *GenericDAG[T]) isEdge(srcHash, dstHash interface{}) bool {
	if _, exists := d.outboundEdge[srcHash]; !exists {
		return false
	}
	if _, exists := d.outboundEdge[srcHash][dstHash]; !exists {
		return false
	}
	if _, exists := d.inboundEdge[dstHash]; !exists {
		return false
	}
	if _, exists := d.inboundEdge[dstHash][srcHash]; !exists {
		return false
	}
	return true
}

// DeleteEdge deletes the edge between srcID and dstID.
// DeleteEdge returns an error if srcID or dstID are empty or unknown,
// or if there is no edge between srcID and dstID.
func (d *GenericDAG[T]) DeleteEdge(srcID, dstID string) error {
	d.muDAG.Lock()
	defer d.muDAG.Unlock()

	if err := d.saneID(srcID); err != nil {
		return err
	}
	if err := d.saneID(dstID); err != nil {
		return err
	}
	if srcID == dstID {
		return SrcDstEqualError{srcID, dstID}
	}

	src := d.vertexValues[srcID]
	srcHash := d.hashVertex(src)
	dst := d.vertexValues[dstID]
	dstHash := d.hashVertex(dst)

	if !d.isEdge(srcHash, dstHash) {
		return EdgeUnknownError{srcID, dstID}
	}

	// get descendants and ancestors as they are now
	descendants := copyMap(d.getDescendants(srcHash))
	ancestors := copyMap(d.getAncestors(dstHash))

	// delete outbound and inbound
	delete(d.outboundEdge[srcHash], dstHash)
	delete(d.inboundEdge[dstHash], srcHash)

	// for src and all its descendants delete cached ancestors
	for descendant := range descendants {
		delete(d.ancestorsCache, descendant)
	}
	delete(d.ancestorsCache, srcHash)

	// for dst and all its ancestors delete cached descendants
	for ancestor := range ancestors {
		delete(d.descendantsCache, ancestor)
	}
	delete(d.descendantsCache, dstHash)

	return nil
}

// GetOrder returns the number of vertices in the graph.
func (d *GenericDAG[T]) GetOrder() int {
	d.muDAG.RLock()
	defer d.muDAG.RUnlock()
	return d.getOrder()
}

func (d *GenericDAG[T]) getOrder() int {
	return len(d.vertices)
}

// GetSize returns the number of edges in the graph.
func (d *GenericDAG[T]) GetSize() int {
	d.muDAG.RLock()
	defer d.muDAG.RUnlock()
	return d.getSize()
}

func (d *GenericDAG[T]) getSize() int {
	count := 0
	for _, value := range d.outboundEdge {
		count += len(value)
	}
	return count
}

// GetLeaves returns all vertices without children.
func (d *GenericDAG[T]) GetLeaves() map[string]T {
	d.muDAG.RLock()
	defer d.muDAG.RUnlock()
	return d.getLeaves()
}

func (d *GenericDAG[T]) getLeaves() map[string]T {
	leaves := make(map[string]T)
	for vHash := range d.vertices {
		dstIDs, ok := d.outboundEdge[vHash]
		if !ok || len(dstIDs) == 0 {
			id := d.vertices[vHash]
			leaves[id] = d.vertexValues[id]
		}
	}
	return leaves
}

// IsLeaf returns true if the vertex with the given id has no children.
// IsLeaf returns an error if id is empty or unknown.
func (d *GenericDAG[T]) IsLeaf(id string) (bool, error) {
	d.muDAG.RLock()
	defer d.muDAG.RUnlock()
	if err := d.saneID(id); err != nil {
		return false, err
	}
	return d.isLeaf(id), nil
}

func (d *GenericDAG[T]) isLeaf(id string) bool {
	v := d.vertexValues[id]
	vHash := d.hashVertex(v)
	dstIDs, ok := d.outboundEdge[vHash]
	if !ok || len(dstIDs) == 0 {
		return true
	}
	return false
}

// GetRoots returns all vertices without parents.
func (d *GenericDAG[T]) GetRoots() map[string]T {
	d.muDAG.RLock()
	defer d.muDAG.RUnlock()
	return d.getRoots()
}

func (d *GenericDAG[T]) getRoots() map[string]T {
	roots := make(map[string]T)
	for vHash := range d.vertices {
		srcIDs, ok := d.inboundEdge[vHash]
		if !ok || len(srcIDs) == 0 {
			id := d.vertices[vHash]
			roots[id] = d.vertexValues[id]
		}
	}
	return roots
}

// IsRoot returns true if the vertex with the given id has no parents.
// IsRoot returns an error if id is empty or unknown.
func (d *GenericDAG[T]) IsRoot(id string) (bool, error) {
	d.muDAG.RLock()
	defer d.muDAG.RUnlock()
	if err := d.saneID(id); err != nil {
		return false, err
	}
	return d.isRoot(id), nil
}

func (d *GenericDAG[T]) isRoot(id string) bool {
	v := d.vertexValues[id]
	vHash := d.hashVertex(v)
	srcIDs, ok := d.inboundEdge[vHash]
	if !ok || len(srcIDs) == 0 {
		return true
	}
	return false
}

// GetVertices returns all vertices as a map of id to value.
func (d *GenericDAG[T]) GetVertices() map[string]T {
	d.muDAG.RLock()
	defer d.muDAG.RUnlock()
	out := make(map[string]T, len(d.vertexValues))
	for id, value := range d.vertexValues {
		out[id] = value
	}
	return out
}

// GetParents returns all parents of the vertex with the id.
// GetParents returns an error if id is empty or unknown.
func (d *GenericDAG[T]) GetParents(id string) (map[string]T, error) {
	d.muDAG.RLock()
	defer d.muDAG.RUnlock()
	if err := d.saneID(id); err != nil {
		return nil, err
	}
	v := d.vertexValues[id]
	vHash := d.hashVertex(v)
	parents := make(map[string]T)
	for pv := range d.inboundEdge[vHash] {
		pid := d.vertices[pv]
		parents[pid] = d.vertexValues[pid]
	}
	return parents, nil
}

// GetChildren returns all children of the vertex with the id.
// GetChildren returns an error if id is empty or unknown.
func (d *GenericDAG[T]) GetChildren(id string) (map[string]T, error) {
	d.muDAG.RLock()
	defer d.muDAG.RUnlock()
	return d.getChildren(id)
}

func (d *GenericDAG[T]) getChildren(id string) (map[string]T, error) {
	if err := d.saneID(id); err != nil {
		return nil, err
	}
	v := d.vertexValues[id]
	vHash := d.hashVertex(v)
	children := make(map[string]T)
	for cv := range d.outboundEdge[vHash] {
		cid := d.vertices[cv]
		children[cid] = d.vertexValues[cid]
	}
	return children, nil
}

// GetAncestors returns all ancestors of the vertex with the id.
// GetAncestors returns an error if id is empty or unknown.
func (d *GenericDAG[T]) GetAncestors(id string) (map[string]T, error) {
	d.muDAG.RLock()
	defer d.muDAG.RUnlock()
	if err := d.saneID(id); err != nil {
		return nil, err
	}
	v := d.vertexValues[id]
	vHash := d.hashVertex(v)
	ancestors := make(map[string]T)
	for av := range d.getAncestors(vHash) {
		aid := d.vertices[av]
		ancestors[aid] = d.vertexValues[aid]
	}
	return ancestors, nil
}

func (d *GenericDAG[T]) getAncestors(vHash interface{}) map[interface{}]struct{} {
	// in the best case we have already a populated cache
	d.muCache.RLock()
	cache, exists := d.ancestorsCache[vHash]
	d.muCache.RUnlock()
	if exists {
		return cache
	}

	// lock this vertex to work on it exclusively
	d.verticesLocked.lock(vHash)
	defer d.verticesLocked.unlock(vHash)

	// now as we have locked this vertex, check (again) that no one has
	// meanwhile populated the cache
	d.muCache.RLock()
	cache, exists = d.ancestorsCache[vHash]
	d.muCache.RUnlock()
	if exists {
		return cache
	}

	// as there is no cache, we start from scratch and collect all ancestors locally
	cache = make(map[interface{}]struct{})
	var mu sync.Mutex
	if parents, ok := d.inboundEdge[vHash]; ok {
		// for each parent collect its ancestors
		for parent := range parents {
			parentAncestors := d.getAncestors(parent)
			mu.Lock()
			for ancestor := range parentAncestors {
				cache[ancestor] = struct{}{}
			}
			cache[parent] = struct{}{}
			mu.Unlock()
		}
	}

	// remember the collected ancestors
	d.muCache.Lock()
	d.ancestorsCache[vHash] = cache
	d.muCache.Unlock()
	return cache
}

// GetOrderedAncestors returns all ancestors of the vertex with id
// in a breath-first order. Only the first occurrence of each vertex is returned.
// GetOrderedAncestors returns an error if id is empty or unknown.
func (d *GenericDAG[T]) GetOrderedAncestors(id string) ([]string, error) {
	d.muDAG.RLock()
	defer d.muDAG.RUnlock()
	ids, _, err := d.AncestorsWalker(id)
	if err != nil {
		return nil, err
	}
	var ancestors []string
	for aid := range ids {
		ancestors = append(ancestors, aid)
	}
	return ancestors, nil
}

// AncestorsWalker returns a channel and subsequently walks all ancestors of
// the vertex with id in a breath first order. The second channel returned may
// be used to stop further walking. AncestorsWalker returns an error if id is
// empty or unknown.
func (d *GenericDAG[T]) AncestorsWalker(id string) (chan string, chan bool, error) {
	d.muDAG.RLock()
	defer d.muDAG.RUnlock()
	if err := d.saneID(id); err != nil {
		return nil, nil, err
	}
	ids := make(chan string)
	signal := make(chan bool, 1)
	go func() {
		d.muDAG.RLock()
		v := d.vertexValues[id]
		vHash := d.hashVertex(v)
		d.walkAncestors(vHash, ids, signal)
		d.muDAG.RUnlock()
		close(ids)
		close(signal)
	}()
	return ids, signal, nil
}

func (d *GenericDAG[T]) walkAncestors(vHash interface{}, ids chan string, signal chan bool) {
	var fifo []interface{}
	visited := make(map[interface{}]struct{})
	for parent := range d.inboundEdge[vHash] {
		visited[parent] = struct{}{}
		fifo = append(fifo, parent)
	}
	for {
		if len(fifo) == 0 {
			return
		}
		top := fifo[0]
		fifo = fifo[1:]
		for parent := range d.inboundEdge[top] {
			if _, exists := visited[parent]; !exists {
				visited[parent] = struct{}{}
				fifo = append(fifo, parent)
			}
		}
		select {
		case <-signal:
			return
		default:
			ids <- d.vertices[top]
		}
	}
}

// GetDescendants returns all descendants of the vertex with the id.
// GetDescendants returns an error if id is empty or unknown.
func (d *GenericDAG[T]) GetDescendants(id string) (map[string]T, error) {
	d.muDAG.RLock()
	defer d.muDAG.RUnlock()

	if err := d.saneID(id); err != nil {
		return nil, err
	}
	v := d.vertexValues[id]
	vHash := d.hashVertex(v)

	descendants := make(map[string]T)
	for dv := range d.getDescendants(vHash) {
		did := d.vertices[dv]
		descendants[did] = d.vertexValues[did]
	}
	return descendants, nil
}

func (d *GenericDAG[T]) getDescendants(vHash interface{}) map[interface{}]struct{} {
	// in the best case we have already a populated cache
	d.muCache.RLock()
	cache, exists := d.descendantsCache[vHash]
	d.muCache.RUnlock()
	if exists {
		return cache
	}

	// lock this vertex to work on it exclusively
	d.verticesLocked.lock(vHash)
	defer d.verticesLocked.unlock(vHash)

	// now as we have locked this vertex, check (again) that no one has
	// meanwhile populated the cache
	d.muCache.RLock()
	cache, exists = d.descendantsCache[vHash]
	d.muCache.RUnlock()
	if exists {
		return cache
	}

	// as there is no cache, we start from scratch and collect all descendants
	// locally
	cache = make(map[interface{}]struct{})
	var mu sync.Mutex
	if children, ok := d.outboundEdge[vHash]; ok {
		for child := range children {
			childDescendants := d.getDescendants(child)
			mu.Lock()
			for descendant := range childDescendants {
				cache[descendant] = struct{}{}
			}
			cache[child] = struct{}{}
			mu.Unlock()
		}
	}

	// remember the collected descendants
	d.muCache.Lock()
	d.descendantsCache[vHash] = cache
	d.muCache.Unlock()
	return cache
}

// GetOrderedDescendants returns all descendants of the vertex with id
// in a breath-first order. Only the first occurrence of each vertex is returned.
// GetOrderedDescendants returns an error if id is empty or unknown.
func (d *GenericDAG[T]) GetOrderedDescendants(id string) ([]string, error) {
	d.muDAG.RLock()
	defer d.muDAG.RUnlock()
	ids, _, err := d.DescendantsWalker(id)
	if err != nil {
		return nil, err
	}
	var descendants []string
	for did := range ids {
		descendants = append(descendants, did)
	}
	return descendants, nil
}

// DescendantsWalker returns a channel and subsequently walks all descendants
// of the vertex with id in a breath first order. The second channel returned
// may be used to stop further walking. DescendantsWalker returns an error if
// id is empty or unknown.
func (d *GenericDAG[T]) DescendantsWalker(id string) (chan string, chan bool, error) {
	d.muDAG.RLock()
	defer d.muDAG.RUnlock()
	if err := d.saneID(id); err != nil {
		return nil, nil, err
	}
	ids := make(chan string)
	signal := make(chan bool, 1)
	go func() {
		d.muDAG.RLock()
		v := d.vertexValues[id]
		vHash := d.hashVertex(v)
		d.walkDescendants(vHash, ids, signal)
		d.muDAG.RUnlock()
		close(ids)
		close(signal)
	}()
	return ids, signal, nil
}

func (d *GenericDAG[T]) walkDescendants(vHash interface{}, ids chan string, signal chan bool) {
	var fifo []interface{}
	visited := make(map[interface{}]struct{})
	for child := range d.outboundEdge[vHash] {
		visited[child] = struct{}{}
		fifo = append(fifo, child)
	}
	for {
		if len(fifo) == 0 {
			return
		}
		top := fifo[0]
		fifo = fifo[1:]
		for child := range d.outboundEdge[top] {
			if _, exists := visited[child]; !exists {
				visited[child] = struct{}{}
				fifo = append(fifo, child)
			}
		}
		select {
		case <-signal:
			return
		default:
			ids <- d.vertices[top]
		}
	}
}

// GetDescendantsGraph returns a new GenericDAG consisting of the vertex with id
// and all its descendants (i.e. the subgraph). GetDescendantsGraph also returns
// the id of the (copy of the) given vertex within the new graph (i.e. the id of
// the single root of the new graph). GetDescendantsGraph returns an error if id
// is empty or unknown.
func (d *GenericDAG[T]) GetDescendantsGraph(id string) (*GenericDAG[T], string, error) {
	return d.getRelativesGraph(id, false)
}

// GetAncestorsGraph returns a new GenericDAG consisting of the vertex with id
// and all its ancestors (i.e. the subgraph). GetAncestorsGraph also returns the id
// of the (copy of the) given vertex within the new graph (i.e. the id of the
// single leaf of the new graph). GetAncestorsGraph returns an error if id is
// empty or unknown.
func (d *GenericDAG[T]) GetAncestorsGraph(id string) (*GenericDAG[T], string, error) {
	return d.getRelativesGraph(id, true)
}

func (d *GenericDAG[T]) getRelativesGraph(id string, asc bool) (*GenericDAG[T], string, error) {
	// sanity checking
	if id == "" {
		return nil, "", IDEmptyError{}
	}
	v, exists := d.vertexValues[id]
	if !exists {
		return nil, "", IDUnknownError{id}
	}
	vHash := d.hashVertex(v)

	// create a new dag
	newDAG := NewGenericDAG[T]()

	// protect the graph from modification
	d.muDAG.RLock()
	defer d.muDAG.RUnlock()

	// recursively add the current vertex and all its relatives
	newId, err := d.getRelativesGraphRec(vHash, newDAG, make(map[interface{}]string), asc)
	return newDAG, newId, err
}

func (d *GenericDAG[T]) getRelativesGraphRec(vHash interface{}, newDAG *GenericDAG[T], visited map[interface{}]string, asc bool) (newId string, err error) {
	// copy this vertex to the new graph
	newId = d.vertices[vHash]
	if err = newDAG.AddVertexByID(newId, d.vertexValues[newId]); err != nil {
		return
	}

	// mark this vertex as visited
	visited[vHash] = newId

	// get the direct relatives (depending on the direction either parents or children)
	var relatives map[interface{}]struct{}
	var ok bool
	if asc {
		relatives, ok = d.inboundEdge[vHash]
	} else {
		relatives, ok = d.outboundEdge[vHash]
	}

	// for all direct relatives in the original graph
	if ok {
		for relative := range relatives {
			// if we haven't seen this relative
			relativeId, exists := visited[relative]
			if !exists {
				// recursively add this relative
				if relativeId, err = d.getRelativesGraphRec(relative, newDAG, visited, asc); err != nil {
					return
				}
			}

			// add edge to this relative (depending on the direction)
			var srcID, dstID string
			if asc {
				srcID, dstID = relativeId, newId
			} else {
				srcID, dstID = newId, relativeId
			}
			if err = newDAG.AddEdge(srcID, dstID); err != nil {
				return
			}
		}
	}
	return
}

// ReduceTransitively transitively reduces the graph.
func (d *GenericDAG[T]) ReduceTransitively() {
	d.muDAG.Lock()
	defer d.muDAG.Unlock()

	graphChanged := false

	// populate the descendants cache for all roots (i.e. the whole graph)
	for _, root := range d.getRoots() {
		_ = d.getDescendants(d.hashVertex(root))
	}

	// for each vertex
	for vHash := range d.vertices {
		// map of descendants of the children of v
		descendantsOfChildrenOfV := make(map[interface{}]struct{})

		// for each child of v
		for childOfV := range d.outboundEdge[vHash] {
			// collect child descendants
			for descendant := range d.descendantsCache[childOfV] {
				descendantsOfChildrenOfV[descendant] = struct{}{}
			}
		}

		// for each child of v
		for childOfV := range d.outboundEdge[vHash] {
			// remove the edge between v and child, iff child is a
			// descendant of any of the children of v
			if _, exists := descendantsOfChildrenOfV[childOfV]; exists {
				delete(d.outboundEdge[vHash], childOfV)
				delete(d.inboundEdge[childOfV], vHash)
				graphChanged = true
			}
		}
	}

	// flush the descendants- and ancestor cache if the graph has changed
	if graphChanged {
		d.flushCaches()
	}
}

// FlushCaches completely flushes the descendants- and ancestor cache.
func (d *GenericDAG[T]) FlushCaches() {
	d.muDAG.Lock()
	defer d.muDAG.Unlock()
	d.flushCaches()
}

func (d *GenericDAG[T]) flushCaches() {
	d.ancestorsCache = make(map[interface{}]map[interface{}]struct{})
	d.descendantsCache = make(map[interface{}]map[interface{}]struct{})
}

// Copy returns a copy of the GenericDAG.
func (d *GenericDAG[T]) Copy() (*GenericDAG[T], error) {
	// create a new dag
	newDAG := NewGenericDAG[T]()

	// create a map of visited vertices
	visited := make(map[interface{}]string)

	// protect the graph from modification
	d.muDAG.RLock()
	defer d.muDAG.RUnlock()

	// add all roots and their descendants to the new DAG
	roots := d.getRoots()
	for id := range roots {
		root := roots[id]
		if _, err := d.getRelativesGraphRec(d.hashVertex(root), newDAG, visited, false); err != nil {
			return nil, err
		}
	}
	return newDAG, nil
}

// String returns a textual representation of the graph.
func (d *GenericDAG[T]) String() string {
	result := fmt.Sprintf("GenericDAG Vertices: %d - Edges: %d\n", d.GetOrder(), d.GetSize())
	result += "Vertices:\n"
	d.muDAG.RLock()
	for k := range d.vertices {
		result += fmt.Sprintf("  %v\n", k)
	}
	result += "Edges:\n"
	for v, children := range d.outboundEdge {
		for child := range children {
			result += fmt.Sprintf("  %v -> %v\n", v, child)
		}
	}
	d.muDAG.RUnlock()
	return result
}

func (d *GenericDAG[T]) saneID(id string) error {
	// sanity checking
	if id == "" {
		return IDEmptyError{}
	}
	_, exists := d.vertexValues[id]
	if !exists {
		return IDUnknownError{id}
	}
	return nil
}

func (d *GenericDAG[T]) hashVertex(v T) interface{} {
	return d.options.VertexHashFunc(v)
}

// Options sets the options for the GenericDAG.
// Options must be called before any other method of the GenericDAG is called.
func (d *GenericDAG[T]) Options(options Options) {
	d.muDAG.Lock()
	defer d.muDAG.Unlock()
	d.options = options
}