package utils

import (
	"html/template"
	"strings"
)

// Wraps one template into another. In wrapper target should be named as {{ .Htmlstr }}, and wrapper data named Data
func TemplateWrap(tmpl *template.Template, targetName string, targetData any, wrapperName string, wrapperData any) (*strings.Builder, error) {
	buff := &strings.Builder{}
	err := tmpl.ExecuteTemplate(buff, targetName, targetData)
	if err != nil {
		return buff, err
	}

	wrapperCtx := &struct {
		Htmlstr template.HTML
		Data    any
	}{template.HTML(buff.String()), wrapperData}
	buff.Reset()
	err = tmpl.ExecuteTemplate(buff, "index", wrapperCtx)
	if err != nil {
		return buff, err
	}
	return buff, nil
}
