package response

// Response 全局统一定义的强类型 API 泛型响应体
//
// 1. 【Why 为什么这么设计】:
//    前后端分离架构的核心是契约一致性，泛型 Response[T] 能在编译期锁定 Data 载荷类型；
//    杜绝裸写 map[string]any 导致的字段拼写错误与不可控序列化。
//
// 2. 【Flow 底层执行流程】:
//    第一步：携带 Code 业务码标示具体结果；
//    第二步：Message 提供对端展示用的人类可读文本；
//    第三步：Data 携带具体的强类型业务载荷（实体、DTO、列表或空值）。
//
// 3. 【Gotcha 生产避坑】:
//    Data 字段即使为 nil 也应当明确序列化语义，业务失败时通常 Data 传 nil 或空结构。
type Response[T any] struct {
	Code    Code   `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

// PageResult 传统偏移量分页统一数据结构
//
// 适用于管理后台、标准列表查询等需要明确总页数的场景。
type PageResult[T any] struct {
	List     []T   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

// CursorResult 高性能游标深分页统一数据结构
//
// 适用于移动端流式瀑布流、海量数据翻页场景，避免 MySQL LIMIT offset, count 深分页性能衰减。
type CursorResult[T any] struct {
	List       []T    `json:"list"`
	NextCursor *int64 `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}

// Success 构造通用业务成功响应
func Success[T any](data T) Response[T] {
	return Response[T]{
		Code:    CodeSuccess,
		Message: CodeSuccess.Msg(),
		Data:    data,
	}
}

// SuccessWithMsg 构造携带自定义提示信息的业务成功响应
func SuccessWithMsg[T any](data T, message string) Response[T] {
	return Response[T]{
		Code:    CodeSuccess,
		Message: message,
		Data:    data,
	}
}

// Error 构造通用业务失败响应
func Error(code Code) Response[any] {
	return Response[any]{
		Code:    code,
		Message: code.Msg(),
		Data:    nil,
	}
}

// ErrorWithMsg 构造携带定制错误提示的业务失败响应
func ErrorWithMsg(code Code, message string) Response[any] {
	return Response[any]{
		Code:    code,
		Message: message,
		Data:    nil,
	}
}

// NewPageResult 构造标准分页结果对象
//
// 1. 【Why 为什么这么设计】:
//    Go 切片为 nil 时，默认 JSON 序列化为 null，而前端通常期望 list 永远为 Array；
//    若返回 null，前端执行 data.list.map(...) 会直接触发运行时崩溃。
//
// 2. 【Flow 底层执行流程】:
//    第一步：检查传入切片是否为 nil；
//    第二步：若为 nil，自动重置为非 nil 空切片 make([]T, 0)；
//    第三步：装配总数、页码与分页数据并返回。
func NewPageResult[T any](list []T, total int64, page, pageSize int) PageResult[T] {
	if list == nil {
		list = make([]T, 0)
	}
	return PageResult[T]{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
}

// NewCursorResult 构造高性能游标分页结果对象
//
// 同样对 nil 切片实施防御性空切片转换，保障前端列表迭代安全。
func NewCursorResult[T any](list []T, nextCursor *int64, hasMore bool) CursorResult[T] {
	if list == nil {
		list = make([]T, 0)
	}
	return CursorResult[T]{
		List:       list,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}
}
