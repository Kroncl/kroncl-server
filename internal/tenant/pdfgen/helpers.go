package pdfgen

import (
	"fmt"
	"html/template"
)

func getFuncMap() template.FuncMap {
	return template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"multiply": func(a, b float64) float64 {
			return a * b
		},
		"formatDate": func(t interface{}) string {
			// можно добавить форматирование дат
			return ""
		},
		"formatMoney": func(amount float64) string {
			return fmt.Sprintf("%.2f", amount)
		},
	}
}
