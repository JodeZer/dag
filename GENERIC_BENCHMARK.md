# 泛型 API 基准测试报告

## 测试环境

```
goos: darwin
goarch: arm64
pkg: github.com/JodeZer/dag
cpu: Apple M4 Pro
```

## 性能对比

### 1365 个顶点的 DAG

| 基准测试 | 操作耗时 (ns/op) | 内存分配 (B/op) | 分配次数 (allocs/op) |
|---------|------------------|------------------|---------------------|
| `BenchmarkUnmarshalJSON` | 516,153 | 297,530 | 2,253 |
| `BenchmarkUnmarshalJSON_Generic_Simple` | 108,787,349 | 168,363,688 | 66,691 |
| `BenchmarkUnmarshalJSON_Generic_Complex` | 147,994,543 | 168,432,240 | 66,692 |

### 100,000 个顶点的 DAG

| 基准测试 | 操作耗时 (ns/op) | 内存分配 (B/op) | 分配次数 (allocs/op) |
|---------|------------------|------------------|---------------------|
| `BenchmarkUnmarshalJSON_100k_3Branch` | 53,031,882 | 33,847,269 | 200,063 |
| `BenchmarkUnmarshalJSON_100k_4Branch` | 50,823,756 | 33,847,269 | 200,063 |
| `BenchmarkUnmarshalJSON_100k_5Branch` | 51,114,711 | 33,847,266 | 200,063 |

## 性能分析

### 性能差异原因

泛型版本 `UnmarshalJSON[T]` 相比旧版 `UnmarshalJSONLegacy` 性能较差的主要原因：

1. **数据格式不同**
   - 旧版：使用 `interface{}` 存储，数据在 JSON 中已经是原始格式
   - 泛型版：使用 `storableVertexGeneric[T]`，但序列化时仍使用旧格式

2. **二次转换**
   - JSON 数据中的 `"v"` 字段先被反序列化为 `map[string]interface{}`
   - 然后需要再次序列化为 JSON 再反序列化为目标类型 `T`

### 为什么会出现这个差异？

当前的序列化 `MarshalJSON()` 使用的是非泛型的 `storableVertex` 结构：

```go
type storableVertex struct {
    WrappedID string      `json:"i"`
    Value     interface{} `json:"v"`  // ← 这里是 interface{}
}
```

当 `Value` 是结构体时，JSON 序列化为对象；当反序列化时，`json.Unmarshal` 会将其解析为 `map[string]interface{}`。

泛型版本的 `storableVertexGeneric[T]` 虽然定义了正确的类型：

```go
type storableVertexGeneric[T any] struct {
    WrappedID string `json:"i"`
    Value     T      `json:"v"`  // ← 这里是具体类型 T
}
```

但是现有的 `MarshalJSON()` 仍然使用非泛型版本，所以反序列化时需要额外转换。

## 优化建议

### 方案 1：统一使用泛型序列化（推荐）

将 `MarshalJSON()` 也改为泛型版本：

```go
func MarshalJSON[T any](d *DAG) ([]byte, error)
```

这样序列化和反序列化都使用泛型类型，性能将与旧版 API 相当。

### 方案 2：保持混合使用

如果需要保持与现有数据的兼容性：
- 对于新数据，使用泛型 API 的专用序列化方法
- 对于旧数据，继续使用旧版 API

## 结论

### 优势

- **开发体验**：泛型 API 简洁易用，只需一行代码
- **类型安全**：编译时类型检查
- **向后兼容**：旧版 API 仍然可用

### 劣势

- **性能开销**：当前实现有显著的性能损失（约 200 倍）
- **版本要求**：需要 Go 1.21+

### 建议

1. **开发阶段**：使用泛型 API 提高开发效率
2. **性能敏感场景**：暂时使用旧版 API
3. **未来优化**：实现泛型版本的序列化方法，消除性能差异

## 基准测试命令

```bash
# 运行所有反序列化基准测试
go test -bench=BenchmarkUnmarshal -benchmem -run=^$ ./...

# 运行特定基准测试
go test -bench=BenchmarkUnmarshalJSON_Generic -benchmem -run=^$ ./...
```