# dag

[![run tests](https://github.com/JodeZer/dag/workflows/Run%20Tests/badge.svg?branch=master)](https://github.com/JodeZer/dag/actions?query=branch%3Amaster)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/JodeZer/dag)](https://pkg.go.dev/github.com/JodeZer/dag)
[![Go Report Card](https://goreportcard.com/badge/github.com/JodeZer/dag)](https://goreportcard.com/report/github.com/JodeZer/dag)
[![Gitpod ready-to-code](https://img.shields.io/badge/Gitpod-ready--to--code-blue?logo=gitpod)](https://gitpod.io/#https://github.com/JodeZer/dag)
[![CodeQL](https://github.com/JodeZer/dag/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/JodeZer/dag/actions/workflows/codeql-analysis.yml)
<!--[![Scorecards supply-chain security](https://github.com/JodeZer/dag/actions/workflows/scorecards.yml/badge.svg)](https://github.com/JodeZer/dag/actions/workflows/scorecards.yml)-->
[![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/6402/badge)](https://bestpractices.coreinfrastructure.org/projects/6402)


Implementation of directed acyclic graphs (DAGs).

The implementation is fast and thread-safe. It prevents adding cycles or 
duplicates and thereby always maintains a valid DAG. The implementation caches
descendants and ancestors to speed up subsequent calls. 

<!--
github.com/JodeZer/dag:

3.770388s to add 597871 vertices and 597870 edges
1.578741s to get descendants
0.143887s to get descendants 2nd time
0.444065s to get descendants ordered
0.000008s to get children
1.301297s to transitively reduce the graph with caches poupulated
2.723708s to transitively reduce the graph without caches poupulated
0.168572s to delete an edge from the root


"github.com/hashicorp/terraform/dag":

3.195338s to add 597871 vertices and 597870 edges
1.121812s to get descendants
1.803096s to get descendants 2nd time
3.056972s to transitively reduce the graph
-->




## Quickstart

Running: 

``` go
package main

import (
	"fmt"
	"github.com/JodeZer/dag"
)

func main() {

	// initialize a new graph
	d := NewDAG()

	// init three vertices
	v1, _ := d.AddVertex(1)
	v2, _ := d.AddVertex(2)
	v3, _ := d.AddVertex(struct{a string; b string}{a: "foo", b: "bar"})

	// add the above vertices and connect them with two edges
	_ = d.AddEdge(v1, v2)
	_ = d.AddEdge(v1, v3)

	// describe the graph
	fmt.Print(d.String())
}
```

will result in something like:

```
DAG Vertices: 3 - Edges: 2
Vertices:
  1
  2
  {foo bar}
Edges:
  1 -> 2
  1 -> {foo bar}
```

## Serialization

DAGs can be serialized to and deserialized from JSON.

### Simple Deserialization (Recommended)

For most use cases, use `UnmarshalFromJSON` which provides a simple API:

```go
// Simple string type
var vertexType string
dag, err := dag.UnmarshalFromJSON(data, &vertexType, dag.DefaultAcyclic())

// Complex custom type
type Person struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}
var vertexType Person
dag, err := dag.UnmarshalFromJSON(data, &vertexType, dag.DefaultAcyclic())

// Integer type
var vertexType int
dag, err := dag.UnmarshalFromJSON(data, &vertexType, dag.DefaultAcyclic())
```

### Advanced Deserialization

For complete control over vertex types, implement the `StorableDAG` interface and use `UnmarshalJSON`:

```go
type MyVertex struct {
    WID string `json:"i"`
    Val string `json:"v"`
}

func (v MyVertex) Vertex() (string, interface{}) {
    return v.WID, v.Val
}

type MyStorableDAG struct {
    StorableVertices []MyVertex     `json:"vs"`
    StorableEdges    []storableEdge `json:"es"`
}

func (g MyStorableDAG) Vertices() []Vertexer {
    l := make([]Vertexer, 0, len(g.StorableVertices))
    for _, v := range g.StorableVertices {
        l = append(l, v)
    }
    return l
}

func (g MyStorableDAG) Edges() []Edger {
    l := make([]Edger, 0, len(g.StorableEdges))
    for _, e := range g.StorableEdges {
        l = append(l, e)
    }
    return l
}

// Usage
var wd MyStorableDAG
dag, err := dag.UnmarshalJSON(data, &wd, dag.DefaultAcyclic())
```

### Serialization

To serialize a DAG to JSON:

```go
d := NewDAG()
// ... add vertices and edges ...

data, err := d.MarshalJSON()
if err != nil {
    panic(err)
}
```
