package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Ok 在 Gin 上下文中输出标准成功响应 (HTTP 200)
//
// 1. 【Why 为什么这么设计】:
//    将 Gin 的 c.JSON 封装收口，屏蔽底层序列化细节，使控制器（Handler）代码极简且统一；
//    直接复用通用 Response 结构体，实现与前端契约严格绑定。
//
// 2. 【Flow 底层执行流程】:
//    第一步：构建 CodeSuccess 的 Response 结构体；
//    第二步：调用 c.JSON(http.StatusOK, resp) 将数据以 application/json 格式写入 ResponseWriter。
//
// 3. 【Gotcha 生产避坑】:
//    调用此方法后 handler 应直接 return，切勿多次向同一个 Context 调用 c.JSON 导致重复写头（Header Re-write 警告）。
func Ok(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Success(data))
}

// OkWithMsg 在 Gin 上下文中输出携带自定义成功提示的响应
func OkWithMsg(c *gin.Context, data any, message string) {
	c.JSON(http.StatusOK, SuccessWithMsg(data, message))
}

// Fail 在 Gin 上下文中输出业务失败响应
//
// 1. 【Why 为什么这么设计】:
//    自动根据业务 Code 转换出规范的 HTTP 状态码（如 401, 403, 404, 422），
//    既满足客户端业务码精确识别，又满足网关、SLB 与 HTTP 标准规范。
//
// 2. 【Flow 底层执行流程】:
//    第一步：调用 code.HTTPStatus() 获取对应的标准 HTTP Status Code；
//    第二步：组装包含业务码与提示文本的通用 Fail 响应载荷；
//    第三步：向客户端输出 JSON。
func Fail(c *gin.Context, code Code) {
	c.JSON(code.HTTPStatus(), Error(code))
}

// FailWithMsg 在 Gin 上下文中输出携带自定义错误提示的业务失败响应
func FailWithMsg(c *gin.Context, code Code, message string) {
	c.JSON(code.HTTPStatus(), ErrorWithMsg(code, message))
}

// Page 在 Gin 上下文中输出标准传统分页响应
func Page[T any](c *gin.Context, list []T, total int64, page, pageSize int) {
	pageData := NewPageResult(list, total, page, pageSize)
	c.JSON(http.StatusOK, Success(pageData))
}

// Cursor 在 Gin 上下文中输出标准游标分页响应
func Cursor[T any](c *gin.Context, list []T, nextCursor *int64, hasMore bool) {
	cursorData := NewCursorResult(list, nextCursor, hasMore)
	c.JSON(http.StatusOK, Success(cursorData))
}
