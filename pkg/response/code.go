package response

import "net/http"

// Code 定义强类型领域业务状态码
type Code int

const (
	// Success 操作成功
	CodeSuccess Code = 200

	// 10000 ~ 10999: 身份认证与用户域 (Auth & User Domain)
	CodeUserNotFound        Code = 10001 // 用户不存在
	CodeUserAlreadyExists   Code = 10002 // 用户名或邮箱已注册
	CodePasswordInvalid     Code = 10003 // 账号密码错误
	CodeTokenExpired        Code = 10004 // Access Token 已过期 (触发前端静默刷新)
	CodeRefreshTokenInvalid Code = 10005 // Refresh Token 无效或过期 (强制重新登录)
	CodePermissionDenied    Code = 10006 // 角色权限不足 (禁止访问受保护资源)

	// 20000 ~ 20999: 资产与内容域 (Article & Asset Domain)
	CodeArticleNotFound        Code = 20001 // 文章资产不存在或已软删除
	CodeArticleStatusInvalid   Code = 20002 // 非法状态机流转
	CodeArticleAuditRejected   Code = 20003 // 内容包含违规敏感词被系统驳回
	CodeArticleForbiddenDelete Code = 20004 // 无权删除该文章资产
	CodeArticleAlreadyLiked    Code = 20005 // 用户已点赞，防刷拦截

	// 30000 ~ 30999: 通用基础设施与网络域 (Infra & Network Domain)
	CodeValidationFailed   Code = 30001 // 请求入参反序列化或格式校验失败
	CodeIdempotentRepeated Code = 30002 // 触发接口幂等性排他锁 (请勿并发重复提交)
	CodeUploadTooLarge     Code = 30003 // 上传文件体积超出最大允许限制
	CodeSystemBusy         Code = 30004 // 服务熔断保护或系统繁忙
)

// codeMsgMap 业务状态码与提示信息映射表
var codeMsgMap = map[Code]string{
	CodeSuccess:                "操作成功",
	CodeUserNotFound:        "用户不存在",
	CodeUserAlreadyExists:   "用户名或邮箱已被占用",
	CodePasswordInvalid:     "用户名或密码错误",
	CodeTokenExpired:        "登录凭据已过期，请刷新令牌",
	CodeRefreshTokenInvalid: "刷新令牌失效，请重新登录",
	CodePermissionDenied:    "权限不足，无法执行此操作",
	CodeArticleNotFound:        "知识资产不存在或已被移除",
	CodeArticleStatusInvalid:   "当前资产状态不允许执行该流转操作",
	CodeArticleAuditRejected:   "内容未能通过安全风控审查",
	CodeArticleForbiddenDelete: "仅作者本人或系统管理员有权执行删除",
	CodeArticleAlreadyLiked:    "您已经点赞过该资产，请勿重复操作",
	CodeValidationFailed:   "请求参数校验失败，请检查提交内容",
	CodeIdempotentRepeated: "请求正在处理中，请勿重复提交",
	CodeUploadTooLarge:     "上传文件体积超出系统阈值",
	CodeSystemBusy:         "系统繁忙，请稍后重试",
}

// Msg 获取业务错误码对应的可读文本信息
//
// 1. 【Why 为什么这么设计】:
//    统一维护业务错误描述，保证服务端对外暴露的人类可读信息一致；
//    杜绝在各 Handler 中手写散装提示文本，便于国际化拓展或集中修订。
//
// 2. 【Flow 底层执行流程】:
//    第一步：在 codeMsgMap 中检索给定业务码对应的文本；
//    第二步：若匹配成功则返回定制文本；
//    第三步：若传入未定义的未知错误码，兜底返回通用提示。
//
// 3. 【Gotcha 生产避坑】:
//    查询 map 时需做好未命中时的默认值兜底，避免未知错误码返回空字符串导致前端 Toast 渲染空白。
func (c Code) Msg() string {
	if msg, ok := codeMsgMap[c]; ok {
		return msg
	}
	return "未知系统错误"
}

// HTTPStatus 将领域业务错误码智能映射为语义化的 HTTP 状态码
//
// 1. 【Why 为什么这么设计】:
//    大厂生产级 API 严格遵循 RESTful 与 HTTP 规范，反对盲目使用“全局 HTTP 200”反模式；
//    语义化的 HTTP 状态码使得网关、反向代理（Nginx）、前端 Axios 拦截器能原生捕获 401/403/404 等网络状态。
//
// 2. 【Flow 底层执行流程】:
//    第一步：根据业务错误码类型匹配对应的 HTTP 响应码（如 10004 -> 401, 10006 -> 403）；
//    第二步：默认未分类的业务异常映射为 500 Internal Server Error，成功操作映射为 200 OK。
//
// 3. 【Gotcha 生产避坑】:
//    必须确保验证失败返回 422 (Unprocessable Entity) 或 400，资源不存在返回 404，
//    否则上层 CDN/监控探针无法真实感知系统异常率。
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
