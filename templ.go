package main

import (
	"html/template"
	"strings"
	"unicode"
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

func VacuusAnimationNameConverter(name string) string {
	var (
		strSliceCopy []string
	)

	strSlice := strings.Split(name, "-")
	for _, str := range strSlice {
		runes := []rune(str)
		runes[0] = unicode.ToUpper(runes[0])
		strSliceCopy = append(strSliceCopy, string(runes))
	}

	return strings.Join(strSliceCopy, " ")
}

func newTemplate() *Templ {
	funcMap := template.FuncMap{
		"Repeat":                       Repeat,
		"ReplaceAll":                   ReplaceAll,
		"TruncateTime":                 TruncateTime,
		"VacuusAnimationNameConverter": VacuusAnimationNameConverter,
	}
	t := template.Must(template.New("").Funcs(funcMap).ParseGlob("views/*.html"))
	t = template.Must(t.ParseGlob("views/components/*.html"))
	t = template.Must(t.ParseGlob("views/components/opus/*.html"))
	t = template.Must(t.ParseGlob("views/components/chrysus/*.html"))
	t = template.Must(t.ParseGlob("views/components/vacuus/*.html"))
	t = template.Must(t.ParseGlob("views/components/nuntius/*.html"))

	return &Templ{
		templ: t,
	}
}
