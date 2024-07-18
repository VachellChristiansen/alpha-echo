package main

import (
	"html/template"
	"strings"
)

// Template Renderer
type Templ struct {
	templ *template.Template
}

func Repeat(text string, times int) string {
	return strings.Repeat(text, times)
}

func ReplaceAll(text, old, new string) string {
	return strings.ReplaceAll(text, old, new)
}

func TruncateTime(datetime string) string {
	parts := strings.Split(datetime, "T")
	return parts[0]
}

func newTemplate() *Templ {
	funcMap := template.FuncMap{
		"Repeat":       Repeat,
		"ReplaceAll":   ReplaceAll,
		"TruncateTime": TruncateTime,
	}
	t := template.Must(template.New("").Funcs(funcMap).ParseGlob("views/*.html"))
	t = template.Must(t.ParseGlob("views/components/*.html"))
	t = template.Must(t.ParseGlob("views/components/opus/*.html"))
	t = template.Must(t.ParseGlob("views/components/chrysus/*.html"))

	return &Templ{
		templ: t,
	}
}
