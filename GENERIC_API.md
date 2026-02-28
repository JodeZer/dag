# 泛型 API 使用指南

## 概述

DAG 库现在支持泛型 API，提供了类型安全且高性能的序列化/反序列化功能。

**推荐使用类型化 DAG API (`TypedDAG[T]`)**，它提供了最完整的类型安全体验。

## 快速开始

### 类型化 DAG API（推荐）

这是最推荐的方式，提供完整的编译时类型安全：

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
	// 1. 创建类型化的 DAG
	dag := dag.New[Person]()

	// 2. 类型安全的顶点操作 - 无需类型断言
	alice := Person{Name: "Alice", Age: 30}
	bob := Person{Name: "Bob", Age: 25}

	dag.AddVertexByID("p1", alice)
	dag.AddVertexByID("p2", bob)
	dag.AddEdge("p1", "p2")

	// 3. 类型安全的顶点访问 - 返回 Person 而非 interface{}
	person, err := dag.GetVertex("p1")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s is %d years old\n", person.Name, person.Age)
	// 直接访问字段，无需类型断言！

	// 4. 自动推断序列化 - 无需泛型参数
	data, err := dag.MarshalJSON()
	if err != nil {
		panic(err)
	}

	// 5. 反序列化时指定类型（合理 - 需要知道类型）
	restored, err := dag.UnmarshalJSON[Person](data, dag.Options{})
	if err != nil {
		panic(err)
	}

	// 6. 类型安全的顶点访问
	restoredPerson, _ := restored.GetVertex("p1")
	fmt.Printf("Restored: %s is %d years old\n", restoredPerson.Name, restoredPerson.Age)
}
```

### 泛型序列化/反序列化（使用 *DAG）

如果你需要继续使用非类型化的 `*DAG`：

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
	restored, err := dag.UnmarshalJSONGeneric[string](data, dag.Options{})
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
	// 创建类型化的 DAG
	d := dag.New[Person]()
	alice := Person{Name: "Alice", Age: 30}
	bob := Person{Name: "Bob", Age: 25}

	d.AddVertexByID("p1", alice)
	d.AddVertexByID("p2", bob)
	d.AddEdge("p1", "p2")

	// 序列化：自动推断，无需泛型参数
	data, err := d.MarshalJSON()
	if err != nil {
		panic(err)
	}

	// 反序列化
	restored, err := dag.UnmarshalJSON[Person](data, dag.Options{})
	if err != nil {
		panic(err)
	}

	// 类型安全的访问
	person, _ := restored.GetVertex("p1")
	fmt.Printf("%+v\n", person) // 输出: {Name:Alice Age:30}
	// 直接访问字段，无需类型断言！
}
```

## API 对比

### 类型化 DAG API（最推荐）

```go
// 创建 DAG
dag := dag.New[Person]()

// 类型安全的顶点操作
dag.AddVertexByID("p1", Person{Name: "Alice", Age: 30})

// 类型安全的顶点访问 - 无需类型断言
person, _ := dag.GetVertex("p1")
fmt.Println(person.Name) // 直接访问字段

// 自动推断的序列化
data, _ := dag.MarshalJSON()

// 反序列化
restored, _ := dag.UnmarshalJSON[Person](data, dag.Options{})
```

**优势**：
- 完整的编译时类型检查
- 所有方法返回类型化值，无需类型断言
- 更好的 IDE 支持（自动补全、类型提示）
- 简洁的序列化 API（`MarshalJSON()` 无需泛型参数）
- 向后兼容现有 `*DAG` API

### 泛型序列化 API（使用 *DAG）

```go
// 序列化（泛型）
data, err := dag.MarshalGeneric[MyType](d)

// 反序列化（泛型）
dag, err := dag.UnmarshalJSONGeneric[MyType](data, dag.Options{})
```

**优势**：
- 适用于现有的非类型化 `*DAG`
- 类型安全的序列化/反序列化
- 无需样板代码

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

## TypedDAG[T] API 方法

`TypedDAG[T]` 提供了与 `*DAG` 相同的方法，但所有返回值都是类型化的：

### 顶点操作
- `AddVertex(v T) (string, error)` - 添加顶点，返回生成的 ID
- `AddVertexByID(id string, v T) error` - 使用指定 ID 添加顶点
- `GetVertex(id string) (T, error)` - 获取顶点（返回 T，非 interface{}）
- `GetVertices() map[string]T` - 获取所有顶点
- `DeleteVertex(id string) error` - 删除顶点

### 边操作
- `AddEdge(srcID, dstID string) error` - 添加边
- `IsEdge(srcID, dstID string) (bool, error)` - 检查边是否存在
- `DeleteEdge(srcID, dstID string) error` - 删除边

### 图信息
- `GetOrder() int` - 返回顶点数量
- `GetSize() int` - 返回边数量
- `IsEmpty() bool` - 检查图是否为空
- `GetLeaves() map[string]T` - 获取所有叶子节点（返回 T，非 interface{}）
- `GetRoots() map[string]T` - 获取所有根节点（返回 T，非 interface{}）

### 关系查询
- `GetParents(id string) (map[string]T, error)` - 获取父节点（返回 T，非 interface{}）
- `GetChildren(id string) (map[string]T, error)` - 获取子节点（返回 T，非 interface{}）
- `GetAncestors(id string) (map[string]T, error)` - 获取所有祖先（返回 T，非 interface{}）
- `GetDescendants(id string) (map[string]T, error)` - 获取所有后代（返回 T，非 interface{}）

### 子图操作
- `GetDescendantsGraph(id string) (*TypedDAG[T], string, error)` - 获取后代子图
- `GetAncestorsGraph(id string) (*TypedDAG[T], string, error)` - 获取祖先子图
- `Copy() (*TypedDAG[T], error)` - 复制图

### 序列化
- `MarshalJSON() ([]byte, error)` - 序列化（自动推断类型，无需泛型参数）

### 其他
- `Options(options Options)` - 设置选项
- `ToDAG() *DAG` - 获取内部的 `*DAG`（用于向后兼容）

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

**注意**：对于复杂类型，由于旧方法使用 `interface{}` 存储，新方法反序列化时会有类型转换开销。建议对新数据统一使用类型化 API。

### 使用 `TypedDAG[T]` 的额外好处

`TypedDAG[T]` 不仅提供了类型安全的序列化，所有图操作都是类型安全的：

```go
dag := dag.New[Person]()
dag.AddVertexByID("p1", Person{Name: "Alice", Age: 30})

// 类型安全的访问 - 无需类型断言
person, _ := dag.GetVertex("p1")
fmt.Println(person.Name) // 直接访问字段

// 类型安全的子图访问
leaves := dag.GetLeaves() // 返回 map[string]Person
for _, p := range leaves {
    fmt.Println(p.Name) // 直接访问字段，无需类型断言
}
```

## 迁移指南

### 从旧版 API 迁移到类型化 DAG API

```go
// === 旧 API ===
d := dag.NewDAG()
d.AddVertexByID("v1", "value1")
d.AddVertexByID("v2", "value2")
d.AddEdge("v1", "v2")
data, _ := d.MarshalJSON()
var wd MyStorableDAG
restored, _ := dag.UnmarshalJSONLegacy(data, &wd, dag.Options{})

// === 新类型化 API ===
d := dag.New[string]()  // 创建时指定类型
d.AddVertexByID("v1", "value1")
d.AddVertexByID("v2", "value2")
d.AddEdge("v1", "v2")
data, _ := d.MarshalJSON()  // 自动推断，无需泛型参数
restored, _ := dag.UnmarshalJSON[string](data, dag.Options{})  // 类型安全
vertex, _ := restored.GetVertex("v1")  // 返回 string，而非 interface{}
```

### 从泛型 API 迁移到类型化 DAG API

```go
// === 泛型 API ===
d := dag.NewDAG()
d.AddVertexByID("p1", Person{Name: "Alice", Age: 30})
data, _ := dag.MarshalGeneric[Person](d)
restored, _ := dag.UnmarshalJSONGeneric[Person](data, dag.Options{})
vertices := restored.GetVertices()
p := vertices["p1"].(Person) // 需要类型断言
fmt.Println(p.Name)

// === 类型化 DAG API ===
d := dag.New[Person]()
d.AddVertexByID("p1", Person{Name: "Alice", Age: 30})
data, _ := d.MarshalJSON()  // 无需泛型参数
restored, _ := dag.UnmarshalJSON[Person](data, dag.Options{})
p, _ := restored.GetVertex("p1")  // 无需类型断言
fmt.Println(p.Name)
```

## 性能说明

泛型 API 的性能表现：

| 操作 | 旧版 API | 新泛型 API | 说明 |
|------|---------|-----------|------|
| 序列化（string） | 基准 | ~1.3x | 额外的类型转换开销 |
| 反序列化（string） | 基准 | ~1.2x | 批量优化后，性能接近旧版 |
| 序列化 + 反序列化（泛型配对） | - | 基准 | 推荐使用配对方式 |

**重要优化**：
- 通过批量边添加（`addEdgesBatch`）优化，反序列化性能从 **200x 慢提升到 1.2x 慢**
- 内存分配减少 99%
- 分配次数减少 85%

**类型化 DAG 的性能**：
- `TypedDAG[T]` 在 `DAG` 之上提供类型安全的包装层
- 核心操作（添加顶点、边等）性能与 `*DAG` 相同
- 类型断言只在访问顶点时发生，开销极小

**建议**：在绝大多数场景下，泛型 API 的性能差异（1.2-1.5x）完全可以接受，而开发体验的提升是显著的。

## 版本要求

- Go 1.21 或更高版本（支持泛型）

## 总结

| 方面 | 旧版 API | 泛型 API | 类型化 DAG API |
|------|---------|---------|---------------|
| 代码行数 | ~30 行样板代码 | 2 行 | 最简洁 |
| 类型安全（序列化） | ❌ | ✅ | ✅ |
| 类型安全（所有操作） | ❌ | ❌ | ✅ |
| 序列化性能 | 快 | ~1.3x | ~1.3x |
| 反序列化性能 | 快 | ~1.2x | ~1.2x |
| 学习曲线 | 中等 | 简单 | 简单 |
| IDE 支持 | 基础 | 好 | 最好 |

**推荐做法**：
- **新项目**：统一使用类型化 DAG API (`dag.New[T]()`)
- **旧项目**：可以继续使用旧版 API，或逐步迁移到类型化 API
- **类型化 DAG API 是推荐的默认选择**，提供最完整的类型安全体验