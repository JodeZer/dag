# 泛型 API 使用指南

## 概述

DAG 库现在支持泛型 API，提供了类型安全且高性能的序列化/反序列化功能。

## 快速开始

### 简单类型（字符串、整数等）

```go
package main

import (
    "encoding/json"
    "fmt"
    "github.com/JodeZer/dag"
)

func main() {
    // 创建并序列化 DAG
    d := dag.NewDAG()
    d.AddVertexByID("v1", "value1")
    d.AddVertexByID("v2", "value2")
    d.AddEdge("v1", "v2")

    data, err := d.MarshalJSON()
    if err != nil {
        panic(err)
    }

    // 反序列化：指定顶点类型为 string
    restored, err := dag.UnmarshalJSON[string](data, dag.DefaultAcyclic())
    if err != nil {
        panic(err)
    }

    // 访问顶点值（类型自动推断为 string）
    vertices := restored.GetVertices()
    fmt.Println(vertices["v1"].(string)) // 输出: value1
}
```

### 复杂类型（结构体）

```go
package main

import (
    "encoding/json"
    "fmt"
    "github.com/JodeZer/dag"
)

type Person struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}

func main() {
    // 创建并序列化 DAG
    d := dag.NewDAG()
    alice := Person{Name: "Alice", Age: 30}
    bob := Person{Name: "Bob", Age: 25}

    d.AddVertexByID("p1", alice)
    d.AddVertexByID("p2", bob)
    d.AddEdge("p1", "p2")

    data, err := d.MarshalJSON()
    if err != nil {
        panic(err)
    }

    // 反序列化：指定顶点类型为 Person
    restored, err := dag.UnmarshalJSON[Person](data, dag.DefaultAcyclic())
    if err != nil {
        panic(err)
    }

    // 访问顶点值（类型自动推断为 Person）
    vertices := restored.GetVertices()
    fmt.Printf("%+v\n", vertices["p1"].(Person)) // 输出: {Name:Alice Age:30}
}
```

### 指针类型

```go
// 反序列化为指针类型
restored, err := dag.UnmarshalJSON[*Person](data, dag.DefaultAcyclic())

// 访问顶点值（类型为 *Person）
vertices := restored.GetVertices()
fmt.Printf("%+v\n", vertices["p1"].(*Person))
```

## API 对比

### 新泛型 API（推荐）

```go
// 一行代码完成反序列化，类型安全，无反射开销
dag, err := dag.UnmarshalJSON[MyType](data, opts)
```

### 旧版 API（保留用于向后兼容）

```go
// 需要定义自定义结构体并实现接口
type MyVertex struct {
    WID string `json:"i"`
    Val MyType `json:"v"`
}

func (v MyVertex) Vertex() (string, interface{}) {
    return v.WID, v.Val
}

type MyStorableDAG struct {
    StorableVertices []MyVertex  `json:"vs"`
    StorableEdges    []storableEdge `json:"es"`
}

func (g MyStorableDAG) Vertices() []Vertexer { /* ... */ }
func (g MyStorableDAG) Edges() []Edger { /* ... */ }

var wd MyStorableDAG
dag, err := dag.UnmarshalJSONLegacy(data, &wd, opts)
```

## 类型推断

反序列化后，顶点值的类型会自动推断为泛型参数 `T`：

| 泛型参数 T | 反序列化后的顶点值类型 |
|----------|---------------------|
| `string` | `string` |
| `int` | `float64` (JSON 数字默认类型) |
| `MyStruct` | `MyStruct` |
| `*MyStruct` | `*MyStruct` |

## JSON 格式

泛型 API 使用与旧版相同的 JSON 格式：

```json
{
  "vs": [
    {"i": "vertex_id_1", "v": "vertex_value_1"},
    {"i": "vertex_id_2", "v": {"name": "Alice", "age": 30}}
  ],
  "es": [
    {"s": "vertex_id_1", "d": "vertex_id_2"}
  ]
}
```

- `"vs"`: 顶点数组
- `"es"`: 边数组
- `"i"`: 顶点 ID
- `"v"`: 顶点值
- `"s"`: 边源顶点
- `"d"`: 边目标顶点

## 性能说明

泛型 API 在编译时确定类型信息，因此：

- ✅ 无反射开销
- ✅ 直接 JSON 反序列化
- ✅ 类型安全
- ✅ 与旧版 API 性能相当（在相同数据格式下）

## 迁移指南

从旧版 API 迁移到泛型 API：

```go
// 旧版
type MyVertex struct { /* ... */ }
type MyStorableDAG struct { /* ... */ }
func (g MyStorableDAG) Vertices() []Vertexer { /* ... */ }
func (g MyStorableDAG) Edges() []Edger { /* ... */ }
var wd MyStorableDAG
dag, err := dag.UnmarshalJSONLegacy(data, &wd, opts)

// 新版
dag, err := dag.UnmarshalJSON[MyType](data, opts)
```

删除所有自定义结构体定义和接口实现，只需一行代码！

## 版本要求

- Go 1.21 或更高版本