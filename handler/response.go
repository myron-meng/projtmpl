package handler

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
)

const (
	ctxErrKey      string = "error"
	validateErrKey string = "validate_error"
)

// SimpleResponse 仅包含 HTTP status code 和 message 响应
type SimpleResponse struct {
	Code    int           `json:"code"`
	Message string        `json:"message"`
	Errors  []*FieldError `json:"errors,omitempty"`
}

// DataResponse 表示带返回结果的 HTTP 响应
type DataResponse struct {
	*SimpleResponse
	Data any `json:"data,omitempty"`
}

// OK 简单地返回 HTTP 200, OK
func OK(c *fiber.Ctx) error {
	return c.Status(http.StatusOK).JSON(&SimpleResponse{
		Code:    http.StatusOK,
		Message: http.StatusText(http.StatusOK),
	})
}

// Data 返回 HTTP 200, 并带上返回数据
func Data(c *fiber.Ctx, data any) error {
	return c.JSON(&DataResponse{
		SimpleResponse: &SimpleResponse{
			Code:    http.StatusOK,
			Message: http.StatusText(http.StatusOK),
		},
		Data: data,
	})
}

// BadRequest 返回请求解析错误，如果 err != nil, 把 err 加到 ctx 中供后续的日志打印
func BadRequest(c *fiber.Ctx, errs ...error) error {
	var es []*FieldError
	if len(errs) != 0 {
		es = translateError(errs[0])
	}
	resp := DataResponse{
		SimpleResponse: &SimpleResponse{
			Code:    http.StatusBadRequest,
			Message: http.StatusText(http.StatusBadRequest),
			Errors:  es,
		},
	}
	return c.Status(http.StatusBadRequest).JSON(&resp)
}

// SimpleCodeResponse 返回指定的 HTTP 标准 status code 和对应的 message,
// 如果 status code 不是标准定义中的，那么返回的 message 是空字符串
// 可以在这里找到 status code 和 message: https://developer.mozilla.org/zh-CN/docs/Web/HTTP/Status
func SimpleCodeResponse(c *fiber.Ctx, status int, data ...any) error {
	resp := DataResponse{
		SimpleResponse: &SimpleResponse{
			Code:    status,
			Message: http.StatusText(status),
		},
	}
	if len(data) != 0 {
		resp.Data = data[0]
	}
	return c.Status(status).JSON(&resp)
}

// CodeMessageResponse 返回指定的 HTTP 标准 status code 和指定的 message,
// 可以在这里找到 status code 和 message: https://developer.mozilla.org/zh-CN/docs/Web/HTTP/Status
func CodeMessageResponse(c *fiber.Ctx, status int, message string, data ...any) error {
	resp := DataResponse{
		SimpleResponse: &SimpleResponse{
			Code:    status,
			Message: message,
		},
	}
	if len(data) != 0 {
		resp.Data = data[0]
	}
	return c.Status(status).JSON(&resp)
}

// Unauthorized 返回未通过请求身份验证，用户未登录或者登录已过期时返回此状态码
func Unauthorized(c *fiber.Ctx) error {
	return SimpleCodeResponse(c, http.StatusUnauthorized)
}

// UnauthorizedWithReason 返回未通过请求身份验证，并在 data.reason 字段返回用户提交的 authorization 验证未通过的原因
func UnauthorizedWithReason(c *fiber.Ctx, reason string) error {
	return SimpleCodeResponse(
		c,
		http.StatusUnauthorized,
		struct {
			Reason string `json:"reason"`
		}{Reason: reason},
	)
}

// NotFound 返回请求的资源不存在
func NotFound(c *fiber.Ctx, messages ...string) error {
	message := "The resource you requested was not found."
	if len(messages) != 0 {
		message = messages[0]
	}
	return c.Status(http.StatusNotFound).JSON(&SimpleResponse{
		Code:    http.StatusNotFound,
		Message: message,
	})
}

// RequestEntityTooLarge 返回请求体过大
func RequestEntityTooLarge(c *fiber.Ctx) error {
	return SimpleCodeResponse(c, http.StatusRequestEntityTooLarge)
}

// TooManyRequests 返回请求过于频繁错误
func TooManyRequests(c *fiber.Ctx, messages ...string) error {
	message := "Too many requests were made in a short period of time, please try a bit later."
	if len(messages) != 0 {
		message = messages[0]
	}
	return CodeMessageResponse(c, http.StatusTooManyRequests, message)

}

// InternalServerError 返回服务器内部错误，一般是因为业务 bug.
func InternalServerError(c *fiber.Ctx, errs ...error) error {
	return SimpleCodeResponse(c, http.StatusInternalServerError)
}
