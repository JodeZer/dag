# 泛型 API 基准测试报告

## 测试环境

```
goos: darwin
goarch: arm64
pkg: github.com/JodeZer/dag
cpu: Apple M4 Pro
```

## 性能对比

### 序列化性能（1365 个顶点的 DAG）

| 基准测试 | 操作耗时 (ns/op) | 内存分配 (B/op) | 分配次数 (allocs/op) | 相对性能 |
|---------|------------------|------------------|---------------------|----------|
| `MarshalJSON` (旧版) | 615,194 | 663,834 | 5,456 | 基准 (1x) |
| `MarshalJSON_Generic_String` | 764,520 | 1,305,565 | 9,564 | 1.24x 慢 |
| `MarshalJSON_Generic_Complex` | 869,437 | 1,350,437 | 9,564 | 1.41x 慢 |

### 反序列化性能（1365 个顶点的 DAG）

| 基准测试 | 操作耗时 (ns/op) | 内存分配 (B/op) | 分配次数 (allocs/op) | 相对性能 |
|---------|------------------|------------------|---------------------|----------|
| `UnmarshalJSON` (旧版) | 526,644 | 297,530 | 2,253 | 基准 (1x) |
| `UnmarshalJSON_Generic_String` | 109,451,295 | 168,363,513 | 66,691 | **208x 慢** |
| `UnmarshalJSON_Generic_Complex` | 147,929,203 | 168,432,276 | 66,692 | **281x 慢** |

### 大规模图性能（100,000 个顶点的 DAG）

| 基准测试 | 操作耗时 (ns/op) | 内存分配 (B/op) | 分配次数 (allocs/op) |
|---------|------------------|------------------|---------------------|
| `UnmarshalJSON_100k_3Branch` | 53,929,360 | 33,847,269 | 200,063 |
| `UnmarshalJSON_100k_4Branch` | 51,617,095 | 33,847,268 | 200,063 |
| `UnmarshalJSON_100k_5Branch` | 51,801,171 | 33,847,265 | 200,063 |

## 性能分析

### 为什么泛型 API 更慢？

泛型版本的主要性能开销来自以下方面：

1. **序列化时的类型转换**
   - `genericMarshalVisitor.Visit` 对每个顶点尝试类型断言
   - 如果类型不匹配，执行 JSON 序列化/反序列化进行类型转换
   - 对于 DAG 中存储的 `interface{}` 值，几乎总是需要类型转换

2. **额外的分配**
   - 泛型版本需要分配 `storableVertexGeneric[T]` 数组
   - 类型转换时的中间 JSON 数据

### 何时使用泛型 API 能获得更好性能？

泛型 API 在以下场景下可以获得更好的性能：

```go
// 场景 1：直接使用类型数据创建 DAG，不经过 interface{}
// 此序列化直接使用 T 类型，无转换开销

d := dag.NewDAG()
d.AddVertexByID("v1", Person{Name: "Alice", Age: 30})
data, _ := dag.MarshalGeneric[Person](d)  // ✅ 快速

// 场景 2：与泛型 API 配对使用（类型一致）
data, _ := dag.MarshalGeneric[Person](d1)
d2, _ := dag.UnmarshalJSON[Person](data, dag.Options{})  // ✅ 直接反序列化
```

### 何时应该避免泛型 API？

```go
// 场景：处理动态类型数据或混合类型数据
// 由于值存储为 interface{}，类型转换开销巨大

d := dag.NewDAG()
var value1 interface{} = Person{Name: "Alice", Age: 30}
var value2 interface{} = "string_value"
d.AddVertexByID("v1", value1)
d.AddVertexByID("v2", value2)

// 使用旧版 API 更快
data, _ := d.MarshalJSON()  // ✅ 推荐
```

## 优化建议

### 方案 1：分类型存储（推荐）

为不同的顶点值类型创建专用的存储方法：

```go
// 存储已知类型的数据时
func (d *DAG) MarshalTyped[T any](data []byte) (*DAG, error)

// 存储混合类型的数据时
func (d *DAG) MarshalJSON() ([]byte, error)
```

### 方案 2：使用 IDInterface

让顶点值实现 `IDInterface` 接口，避免类型转换：

```go
type IDInterface interface {
    ID() string
}

type Person struct {
    Name string
    Age  int
}

func (p Person) ID() string {
    return p.Name
}

d := dag.NewDAG()
d.AddVertex(Person{Name: "Alice", Age: 30})
// 类型信息保留，无需转换
```

### 方案 3：保持混合使用

```go
// 新数据、已知类型 → 使用泛型 API
data, _ := dag.MarshalGeneric[MyType](d)

// 旧数据、混合类型 → 使用旧版 API
data, _ := d.MarshalJSON()
```

## 基准测试命令

```bash
# 运行所有序列化基准测试
go test -bench=BenchmarkMarshal -benchmem -run=^$ ./...

# 运行所有反序列化基准测试
go test -bench=BenchmarkUnmarshal -benchmem -run=^$ ./...

# 运行泛型配对基准测试
go test -bench='BenchmarkMarshalJSON_Generic|BenchmarkUnmarshalJSON_Generic' -benchmem -run=^$ ./...
```

## 结论

| 方面 | 旧版 API | 新泛型 API |
|------|---------|-----------|
| 序列化性能 | 基准 | ~1.3x 慢 |
| 反序列化性能（已知类型） | - | ~200x 慢 |
| 开发复杂度 | 高（需定义结构体） | 低（一行代码） |
| 类型安全 | ❌ | ✅ |

### 最终建议

**泛型 API 的优势在于开发体验，而非性能。** 如果追求最佳性能，请继续使用旧版 API。泛型 API 适合以下场景：

1. 数据类型一致且已知
2. 开发效率优先于运行时性能
3. 需要类型安全保证

对于性能敏感的场景，旧版 API 仍然是更好的选择。