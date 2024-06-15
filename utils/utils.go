package utils

import (
	"html/template"
	"io"
	"strings"
)

// Wraps one template into another. In wrapper target should be named as {{ .Htmlstr }}, and wrapper data named Data
func TemplateWrap(tmpl *template.Template, w io.Writer, targetName string, targetData any, wrapperName string, wrapperData any) error {
	buff := &strings.Builder{}
	err := tmpl.ExecuteTemplate(buff, targetName, targetData)
	if err != nil {
		return err
	}

	wrapperCtx := &struct {
		Htmlstr template.HTML
		Data    any
	}{template.HTML(buff.String()), wrapperData}
	err = tmpl.ExecuteTemplate(w, "index", wrapperCtx)
	if err != nil {
		return err
	}
	return nil
}
