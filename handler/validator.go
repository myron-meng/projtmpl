package handler

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	eng "github.com/go-playground/validator/v10/translations/en"
)

var validate *validator.Validate
var translator ut.Translator

// init 初始化请求参数校验的 validator
// Note: 这里遇到错误直接 panic, 而没有返回 error
func init() {
	validate = validator.New()
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return "'" + name + "'"
	})

	english := en.New()
	uni := ut.New(english, english)
	var found bool
	translator, found = uni.GetTranslator("en")
	if !found {
		panic(fmt.Errorf("translator [%s] not found", "en"))
	}
	if err := eng.RegisterDefaultTranslations(validate, translator); err != nil {
		panic(err)
	}
}

// FieldError 表示一个请求参数错误
type FieldError struct {
	Message string `json:"message"`
	Tag     string `json:"tag"`
	Value   string `json:"value"`
}

func translateError(err error) []*FieldError {
	if err == nil {
		return nil
	}
	vldErrs, ok := err.(validator.ValidationErrors)
	if !ok {
		return nil
	}

	var errs []*FieldError
	for _, e := range vldErrs {
		errs = append(errs, &FieldError{
			Message: e.Translate(translator),
			Tag:     e.Tag(),
			Value:   e.Param(),
		})
	}
	return errs
}
