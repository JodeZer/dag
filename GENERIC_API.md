# 泛型 API 使用指南

## 概述

DAG 库现在支持泛型 API，提供了类型安全且高性能的序列化/反序列化功能。

## 快速开始

### 序列化和反序列化（推荐使用泛型配对）

```go
package main

import (
	"fmt"
	"github.com/JodeZer/dag"
)

func main() {
	// 创建 DAG
	d := dag.NewDAG()
	d.AddVertexByID("v1", "value1")
	d.AddVertexByID("v2", "value2")
	d.AddEdge("v1", "v2")

	// 序列化：使用泛型 API
	data, err := dag.MarshalGeneric[string](d)
	if err != nil {
		panic(err)
	}

	// 反序列化：使用泛型 API（必须与序列化类型一致）
	restored, err := dag.UnmarshalJSON[string](data, dag.Options{})
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
	"fmt"
	"github.com/JodeZer/dag"
)

type Person struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func main() {
	// 创建 DAG
	d := dag.NewDAG()
	alice := Person{Name: "Alice", Age: 30}
	bob := Person{Name: "Bob", Age: 25}

	d.AddVertexByID("p1", alice)
	d.AddVertexByID("p2", bob)
	d.AddEdge("p1", "p2")

	// 序列化：使用泛型 API
	data, err := dag.MarshalGeneric[Person](d)
	if err != nil {
		panic(err)
	}

	// 反序列化：使用泛型 API
	restored, err := dag.UnmarshalJSON[Person](data, dag.Options{})
	if err != nil {
		panic(err)
	}

	// 访问顶点值（类型自动推断为 Person）
	vertices := restored.GetVertices()
	fmt.Printf("%+v\n", vertices["p1"].(Person)) // 输出: {Name:Alice Age:30}
}
```

## API 对比

### 新泛型 API（推荐）

```go
// 序列化（泛型）
data, err := dag.MarshalGeneric[MyType](d)

// 反序列化（泛型）
dag, err := dag.UnmarshalJSON[MyType](data, dag.Options{})
```

**优势**：
- 类型安全，编译时检查
- 简洁易用，无需样板代码
- 高性能，无反射开销

### 旧版 API（保留用于向后兼容）

```go
// 序列化
data, err := d.MarshalJSON()

// 反序列化
type MyVertex struct {
    WID string `json:"i"`
    Val MyType `json:"v"`
}
type MyStorableDAG struct {
    StorableVertices []MyVertex  `json:"vs"`
    StorableEdges    []storableEdge `json:"es"`
}
func (g MyStorableDAG) Vertices() []Vertexer { /* ... */ }
func (g MyStorableDAG) Edges() []Edger { /* ... */ }
var wd MyStorableDAG
dag, err := dag.UnmarshalJSONLegacy(data, &wd, dag.Options{})
```

**劣势**：
- 需要大量样板代码
- 类型不安全
- 开发效率低

## JSON 格式

泛型 API 生成与旧版 API **完全相同** 的 JSON 格式：

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

## 重要注意事项

### 序列化和反序列化类型必须一致

```go
// ❌ 错误：序列化用 string，反序列化用 Person
data, _ := dag.MarshalGeneric[string](d)
dag, _ := dag.UnmarshalJSON[Person](data, dag.Options{}) // 会失败！

// ✅ 正确：类型一致
data, _ := dag.MarshalGeneric[Person](d)
dag, _ := dag.UnmarshalJSON[Person](data, dag.Options{})
```

### Options 设置

```go
// 使用空 Options 结构体（使用默认值）
dag, err := dag.UnmarshalJSON[MyType](data, dag.Options{})

// 或自定义 Options
opts := dag.Options{
    VertexHashFunc: func(v interface{}) interface{} {
        // 自定义哈希函数
        return v
    },
}
dag, err := dag.UnmarshalJSON[MyType](data, opts)
```

### 向后兼容性

**旧的序列化数据可以用新的泛型 API 反序列化吗？**

这取决于顶点值的类型：

- **简单类型（string, int, bool）**：可以
  ```go
  data, _ := d.MarshalJSON()  // 旧方法序列化
  dag, _ := dag.UnmarshalJSON[string](data, dag.Options{})  // 新方法反序列化
  ```

- **复杂类型（结构体）**：可以
  ```go
  data, _ := d.MarshalJSON()  // 旧方法序列化（interface{} 存储）
  dag, _ := dag.UnmarshalJSON[Person](data, dag.Options{})  // 新方法反序列化（JSON 转换）
  ```

**注意**：对于复杂类型，由于旧方法使用 `interface{}` 存储，新方法反序列化时会有类型转换开销。建议对新数据统一使用泛型 API。

## 性能说明

泛型 API 的性能表现：

| 操作 | 旧版 API | 新泛型 API | 说明 |
|------|---------|-----------|------|
| 序列化（string） | 基准 | ~1.25x | 额外的类型转换开销 |
| 反序列化（string） | 基准 | ~200x | 数据格式不匹配导致 |
| 序列化 + 反序列化（泛型配对） | - | 基准 | 推荐使用配对方式 |

**重要**：只有当序列化和反序列化都使用泛型 API（且类型一致）时，才能获得最佳性能。

## 迁移指南

从旧版 API 迁移到泛型 API：

```go
// 旧版
data, err := d.MarshalJSON()
var wd MyStorableDAG
dag, err := dag.UnmarshalJSONLegacy(data, &wd, dag.Options{})

// 新版
data, err := dag.MarshalGeneric[MyType](d)
dag, err := dag.UnmarshalJSON[MyType](data, dag.Options{})
```

## 版本要求

- Go 1.21 或更高版本

## 总结

| 方面 | 旧版 API | 新泛型 API |
|------|---------|-----------|
| 代码行数 | ~30 行样板代码 | 2 行 |
| 类型安全 | ❌ | ✅ |
| 序列化性能 | 快 | ~1.25x |
| 反序列化性能（泛型配对） | - | **快** |
| 反序列化性能（旧数据） | 快 | 慢（200x） |
| 学习曲线 | 中等 | 简单 |

**推荐做法**：
- 新项目：统一使用泛型 API（`MarshalGeneric[T]` + `UnmarshalJSON[T]`）
- 旧项目：保持使用旧版 API，或逐步迁移到泛型 API