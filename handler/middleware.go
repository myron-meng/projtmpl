package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"

	"projtmpl/pkg/log"

	"github.com/google/uuid"
	loglib "github.com/phuslu/log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

const paramDefaultValue = ""

// confidentialRequests 包含了需要做数据脱敏处理的请求接口名
// 使用请求接口名去辨别接口比使用【HTTP 请求方法 + 路由】组合更加方便
var confidentialRequests = map[string]struct{}{
	"Register": {},
}

// confidentialFields 定义了需要做数据脱敏的请求参数字段名
var confidentialFields = map[string]struct{}{
	"password": {},
}

// formatAny 将 a any 格式化成字符串
// json unmarshal any 可能返回的数据类型：https://pkg.go.dev/encoding/json#Unmarshal
func formatAny(a any) string {
	switch a.(type) {
	case string:
		return a.(string)
	case float64:
		return strconv.FormatFloat(a.(float64), 'f', -1, 64)
	case bool:
		return strconv.FormatBool(a.(bool))
	case []interface{}:
		return ""
	case map[string]interface{}:
		return ""
	default:
		return ""
	}
}

// Logging 中间件打印请求日志
func Logging() fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestLog(c, "Logging")
		return nil
	}
}

// requestLog 打印请求日志
func requestLog(c *fiber.Ctx, from string, errs ...error) {
	start := time.Now().UTC()
	if err := c.Next(); err != nil {
		if err := SimpleCodeResponse(c, http.StatusInternalServerError); err != nil {
			log.Logger.Error().Err(err).Str("request-id", c.GetRespHeader("X-Request-Id")).Msg("write HTTP response failed")
		}
	}
	duration := time.Now().UTC().Sub(start).String()

	var entry *loglib.Entry
	var status int
	if from == "Logging" {
		status = c.Response().StatusCode()
		switch { // 根据 HTTP 状态码设置日志级别
		case status >= http.StatusBadRequest && status < http.StatusInternalServerError:
			entry = log.Logger.Warn()
		case status >= http.StatusInternalServerError:
			entry = log.Logger.Error()
		default:
			entry = log.Logger.Info()
		}
	} else {
		status = http.StatusInternalServerError
		entry = log.Logger.Error()
		if len(errs) > 0 {
			entry.Err(errs[0])
		}
	}
	entry = entry.
		Int("status", status).
		Str("method", c.Method()).
		Str("path", c.Path()).
		Str("name", c.Route().Name).
		Int("bytes", len(c.Response().Body())).
		Str("duration", duration).
		Str("user-agent", string(c.Request().Header.UserAgent())).
		Str("request-id", c.GetRespHeader("X-Request-Id"))

	if len(c.Route().Params) != 0 {
		// 有的话，打印 path param
		for i := range c.Route().Params {
			entry.Str(fmt.Sprintf("param_%s", c.Route().Params[i]), c.Params(c.Route().Params[i], paramDefaultValue))
		}
	}
	if len(c.Request().URI().QueryString()) != 0 {
		// 有的话，打印 query string
		c.Request().URI().QueryArgs().VisitAll(func(key, value []byte) {
			if strings.ToLower(string(key)) == "password" {
				value = []byte(strings.Repeat("*", len(value)))
			}
			entry.Str(fmt.Sprintf("query_%s", string(key)), string(value))
		})
	}

	// 根据 Content-Type 的不同，使用不同的方式检查参数并打印
	// 单个请求 Content-Type 的值是固定的，只能是众多可能值中的一个
	ct := string(c.Request().Header.ContentType())
	if ct == "application/json" {
		// 打印 body
		if _, ok := confidentialRequests[c.Route().Name]; ok {
			// 如果此请求包含有敏感信息，需要做数据脱敏处理
			body := make(map[string]any)
			if err := json.Unmarshal(c.Request().Body(), &body); err != nil {
				// 反序列化失败，无法打印参数，此时把错误信息加到日志条目
				entry.AnErr("body_parse_error", err).Str("body", "unmarshal failed")
			} else {
				// 遍历 confidentialFields, 做数据脱敏处理，依次将请求参数中的隐私字段替换成同等长度的 * 号串
				// 敏感信息字段仅支持 string/number/bool, 如果参数类型不是字符串，那么先将其转换成字符串
				// 暂不支持脱敏对数组或者对象类型的敏感信息字段。也就是说仅支持 JSON 对象的第一层级的简单类型参数做脱敏处理
				// TODO 遍历所有的 []interface{} 和 map[string]interface{}, 直到将请求 body 中的任何位置的字段做数据脱敏处理
				for k := range confidentialFields {
					if _, ok = body[k]; ok {
						body[k] = strings.Repeat("*", len(formatAny(body[k])))
					}
				}
				// 最后将脱敏后的请求 body 序列化并打印
				bodyBytes, _ := json.Marshal(body)
				entry.RawJSON("body", bodyBytes)
			}
		} else {
			// 此请求没有包含敏感信息，直接打印请求 body 即可
			entry.RawJSON("body", c.Request().Body())
		}
	} else if ct == "application/x-www-form-urlencoded" {
		// application/x-www-form-urlencoded 的值跟 query param 差不多，只不过 query param 有长度的限制
		c.Request().PostArgs().VisitAll(func(key, value []byte) {
			k, v := string(key), string(value)
			if _, ok := confidentialFields[strings.ToLower(k)]; ok {
				v = strings.Repeat("*", len(v))
			}
			entry.Str(fmt.Sprintf("post-arg_%s", k), v)
		})
	} else if i := strings.Index(ct, "multipart/form-data"); i != -1 {
		// 打印 form 参数
		// 如果要理解这段代码，需要先熟悉 multipart.Form 的组成
		if form, err := c.Request().MultipartForm(); err == nil {
			for field := range form.Value { // form.Value 的类型是 map[string][]string
				if _, ok := confidentialFields[field]; !ok {
					// 字段没包含有敏感信息字段
					if len(form.Value[field]) == 1 { // form.Value[field] 是一个字符串数组
						entry.Str(fmt.Sprintf("form_%s", field), form.Value[field][0])
					} else {
						for j := range form.Value[field] {
							entry.Str(fmt.Sprintf("form_%s_%d", field, j), form.Value[field][j])
						}
					}
					continue
				}
				// 字段有包含敏感信息字段
				if len(form.Value[field]) == 1 { // form.Value[field] 是一个字符串数组
					// 大多数情况下每个 field 只有一个简单的字符串值
					v := strings.Repeat("*", len(form.Value[field][0]))
					entry.Str(fmt.Sprintf("form_%s", field), v)
				} else {
					vals := make([]string, len(form.Value[field]))
					for j := range form.Value[field] { // form.Value[field] 是一个字符串数组
						vals[j] = strings.Repeat("*", len(form.Value[field][j]))
					}
					entry.Strs(fmt.Sprintf("form_%s", field), vals)
				}
			}
		} else {
			entry.AnErr("body_parse_error", err).Str("body", "parse failed")
		}
	}
	entry.Msg(http.StatusText(status))
}

// CORS 开启跨域支持
func CORS() fiber.Handler {
	return cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders: "authorization, origin, content-type, accept, content-disposition",
	})
}

// RequestID 给请求响应加上 X-Request-Id 响应头
func RequestID() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		id := uuid.NewString()
		ctx.Set("X-Request-Id", id)
		userContext := context.WithValue(ctx.UserContext(), "request_id", id)
		ctx.SetUserContext(userContext)
		return ctx.Next()
	}
}

// identifyPanic 获取 panic 发生的地方
func identifyPanic() string {
	var name, file string
	var line int
	var pc [16]uintptr

	n := runtime.Callers(3, pc[:])
	for _, pc := range pc[:n] {
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}
		file, line = fn.FileLine(pc)
		name = fn.Name()
		if !strings.HasPrefix(name, "runtime.") {
			break
		}
	}

	switch {
	case name != "":
		return fmt.Sprintf("%v:%v", name, line)
	case file != "":
		return fmt.Sprintf("%v:%v", file, line)
	}

	return fmt.Sprintf("pc:%x", pc)
}

// Recover 中间件在发生 panic 时将其恢复，并打印请求日志
func Recover() fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				err := fmt.Errorf(identifyPanic())
				requestLog(c, "Recover", err)
				if err := SimpleCodeResponse(c, http.StatusInternalServerError); err != nil {
					log.Logger.Error().Err(err).Str("request-id", c.GetRespHeader("X-Request-Id")).Msg("write HTTP response failed")
				}
			}
		}()
		return c.Next()
	}
}
