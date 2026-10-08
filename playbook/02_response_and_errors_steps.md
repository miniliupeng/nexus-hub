# 实操剧本 02：通用响应契约与领域业务错误码 (Response & Errors Steps)

> **文档定位**：
> 本剧本详细记录 `nexus-hub` 在 Phase 2 (Module 02) 阶段从零到一敲下的**每一行命令、代码实现细节、真实避坑排查与最终终端输出**。
> 本手册采用自包含设计，即便脱离项目源码，亦可作为 Go Web 企业级统一响应体与领域错误码字典设计的独立指导手册。

---

## 🎯 一、 核心目标与技术背景

- **核心目标**：
  1. 引入并固化 `github.com/gin-gonic/gin` 核心 Web 框架依赖；
  2. 编写 `pkg/response/code.go`：定义跨领域业务错误码枚举、文本字典及自动映射 HTTP 状态码方法；
  3. 编写 `pkg/response/response.go`：实现纯 Go 泛型响应载荷 `Response[T]`、分页 `PageResult[T]` 与游标深分页 `CursorResult[T]`，实施防 `null` 数组空切片保护；
  4. 编写 `pkg/response/gin.go`：封装面向 Gin 表现层的极简输出助手（`Ok`, `Fail`, `Page`, `Cursor`）；
  5. 编写单元测试 `pkg/response/response_test.go` 并通过 `make test-race` 零数据竞争验收。
- **前置环境**：
  - 已完成 Module 00 与 Module 01。

---

## ⌨️ 二、 分步原生实操命令与技术解析

### 步骤 1：引入 Gin 核心依赖与依赖整理
```bash
go get github.com/gin-gonic/gin@latest
go mod tidy
```
- **技术解析**：引入企业级 Gin Web 引擎，由 `go mod tidy` 自动分析依赖图谱并生成锁定的校验和哈希入 `go.sum`。

---

### 步骤 2：创建领域状态码字典 `pkg/response/code.go`
确立领域区间：
- `200`: 成功；
- `10001 ~ 10999`: 身份认证与用户域；
- `20001 ~ 20999`: 资产与内容域；
- `30001 ~ 30999`: 基础设施与系统域。

```go
// 核心映射：自动根据业务 Code 映射 HTTP 状态码
func (c Code) HTTPStatus() int {
    switch c {
    case CodeSuccess:
        return http.StatusOK
    case CodeTokenExpired, CodeRefreshTokenInvalid:
        return http.StatusUnauthorized
    case CodePermissionDenied, CodeArticleForbiddenDelete:
        return http.StatusForbidden
    case CodeUserNotFound, CodeArticleNotFound:
        return http.StatusNotFound
    case CodeUserAlreadyExists, CodeIdempotentRepeated, CodeArticleAlreadyLiked:
        return http.StatusConflict
    case CodeValidationFailed, CodeArticleStatusInvalid:
        return http.StatusUnprocessableEntity
    case CodeUploadTooLarge:
        return http.StatusRequestEntityTooLarge
    case CodeSystemBusy:
        return http.StatusServiceUnavailable
    default:
        return http.StatusInternalServerError
    }
}
```

---

### 步骤 3：编写通用泛型响应体 `pkg/response/response.go`
核心防崩溃机制：针对 `nil` 切片强制调用 `make([]T, 0)` 转为空数组，防止前端 `list.map()` 抛出空指针异常。

---

### 步骤 4：编写 Gin 控制器响应助手 `pkg/response/gin.go`
提供面向 `*gin.Context` 的极简助手函数：
- `Ok(c, data)` ➔ HTTP 200 + Code 200
- `Fail(c, code)` ➔ 语义化 HTTP Status + 业务 Code
- `Page(c, list, total, page, pageSize)` ➔ HTTP 200 + 分页元数据
- `Cursor(c, list, nextCursor, hasMore)` ➔ HTTP 200 + 游标深分页

---

### 步骤 5：运行并发数据竞争测试
```bash
make test-race
```
- **实际终端输出**：
```text
go test -v -race ./...
?   	nexus-hub/cmd/server	[no test files]
=== RUN   TestLoad_Success
--- PASS: TestLoad_Success (0.00s)
=== RUN   TestValidate_InvalidCases
--- PASS: TestValidate_InvalidCases (0.00s)
PASS
ok  	nexus-hub/configs	(cached)
=== RUN   TestCode_MsgAndHTTPStatus
--- PASS: TestCode_MsgAndHTTPStatus (0.00s)
=== RUN   TestPageResult_NilSliceGuard
--- PASS: TestPageResult_NilSliceGuard (0.00s)
=== RUN   TestCursorResult_NilSliceGuard
--- PASS: TestCursorResult_NilSliceGuard (0.00s)
=== RUN   TestGinHelpers
--- PASS: TestGinHelpers (0.00s)
PASS
ok  	nexus-hub/pkg/response	(cached)
```

---

## 避坑与故障排查记录 (Troubleshooting)

### 踩坑 1：包内函数重名冲突 (Name Redeclared)
- **现象**：在同一个 `package response` 下，`response.go` 中定义了 `Fail(code Code)`，而 `gin.go` 中定义了 `Fail(c *gin.Context, code Code)`，Go 编译器报错：`Fail redeclared in this block`。
- **根因**：Go 语言不支持函数重载。
- **解决**：确立职责分层命名：
  - 纯 Go 构造器：`Success` / `Error`；
  - Gin 表现层助手：`Ok` / `Fail` / `Page` / `Cursor`。
  职责明晰，彻底消除重名冲突。
