package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	// 测试环境下静音 Gin 日志输出
	gin.SetMode(gin.TestMode)
}

// TestCode_MsgAndHTTPStatus 测试业务状态码与 HTTP 映射一致性
func TestCode_MsgAndHTTPStatus(t *testing.T) {
	tests := []struct {
		name       string
		code       Code
		wantStatus int
		wantMsg    string
	}{
		{
			name:       "成功状态",
			code:       CodeSuccess,
			wantStatus: http.StatusOK,
			wantMsg:    "操作成功",
		},
		{
			name:       "Token过期401",
			code:       CodeTokenExpired,
			wantStatus: http.StatusUnauthorized,
			wantMsg:    "登录凭据已过期，请刷新令牌",
		},
		{
			name:       "权限不足403",
			code:       CodePermissionDenied,
			wantStatus: http.StatusForbidden,
			wantMsg:    "权限不足，无法执行此操作",
		},
		{
			name:       "文章不存在404",
			code:       CodeArticleNotFound,
			wantStatus: http.StatusNotFound,
			wantMsg:    "知识资产不存在或已被移除",
		},
		{
			name:       "幂等重复提交409",
			code:       CodeIdempotentRepeated,
			wantStatus: http.StatusConflict,
			wantMsg:    "请求正在处理中，请勿重复提交",
		},
		{
			name:       "参数校验失败422",
			code:       CodeValidationFailed,
			wantStatus: http.StatusUnprocessableEntity,
			wantMsg:    "请求参数校验失败，请检查提交内容",
		},
		{
			name:       "未知业务码兜底500",
			code:       Code(99999),
			wantStatus: http.StatusInternalServerError,
			wantMsg:    "未知系统错误",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.code.HTTPStatus(); got != tt.wantStatus {
				t.Errorf("Code.HTTPStatus() = %v, want %v", got, tt.wantStatus)
			}
			if got := tt.code.Msg(); got != tt.wantMsg {
				t.Errorf("Code.Msg() = %v, want %v", got, tt.wantMsg)
			}
		})
	}
}

// TestPageResult_NilSliceGuard 测试空切片序列化为 [] 而非 null
func TestPageResult_NilSliceGuard(t *testing.T) {
	// 传入 nil 切片
	pageRes := NewPageResult[string](nil, 0, 1, 10)
	bytes, err := json.Marshal(pageRes)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	expectedJSON := `{"list":[],"total":0,"page":1,"page_size":10}`
	if string(bytes) != expectedJSON {
		t.Errorf("NewPageResult nil guard failed, got %s, want %s", string(bytes), expectedJSON)
	}
}

// TestCursorResult_NilSliceGuard 测试游标分页空切片序列化
func TestCursorResult_NilSliceGuard(t *testing.T) {
	var cursor int64 = 100
	cursorRes := NewCursorResult[int](nil, &cursor, true)
	bytes, err := json.Marshal(cursorRes)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	expectedJSON := `{"list":[],"next_cursor":100,"has_more":true}`
	if string(bytes) != expectedJSON {
		t.Errorf("NewCursorResult nil guard failed, got %s, want %s", string(bytes), expectedJSON)
	}
}

// TestGinHelpers 测试与 Gin 上下文交互与真实 HTTP 输出
func TestGinHelpers(t *testing.T) {
	// 1. 测试 Ok
	t.Run("Gin_Ok", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		Ok(c, map[string]string{"user": "max"})

		if w.Code != http.StatusOK {
			t.Errorf("Gin Ok HTTP code = %d, want 200", w.Code)
		}

		var resp Response[map[string]string]
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}
		if resp.Code != CodeSuccess || resp.Data["user"] != "max" {
			t.Errorf("Gin Ok payload mismatch: %+v", resp)
		}
	})

	// 2. 测试 Fail 与 HTTP 401 映射
	t.Run("Gin_Fail_TokenExpired", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		Fail(c, CodeTokenExpired)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Gin Fail HTTP code = %d, want 401", w.Code)
		}

		var resp Response[any]
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}
		if resp.Code != CodeTokenExpired {
			t.Errorf("Gin Fail code = %d, want %d", resp.Code, CodeTokenExpired)
		}
	})

	// 3. 测试 Page 响应与空切片安全输出
	t.Run("Gin_Page_EmptySlice", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		Page[string](c, nil, 0, 1, 20)

		if w.Code != http.StatusOK {
			t.Errorf("Gin Page HTTP code = %d, want 200", w.Code)
		}

		var resp Response[PageResult[string]]
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}
		if resp.Data.List == nil || len(resp.Data.List) != 0 {
			t.Errorf("Gin Page list should be empty slice not nil, got: %+v", resp.Data.List)
		}
	})
}
