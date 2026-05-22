package pdfgen

import "github.com/nativebpm/gotenberg-client"

type Config struct {
	Endpoint      string
	TemplatesPath string // путь к папке с шаблонами (например, "./templates")
}

type Service struct {
	client      *gotenberg.Client
	config      Config
	templateDir string
}

type GeneratePDFOptions struct {
	PaperSize    string
	MarginTop    float64
	MarginBottom float64
	MarginLeft   float64
	MarginRight  float64
	Landscape    bool
}

type GenerateFromTemplateRequest struct {
	TemplatePath string // путь относительно папки templates, например "deals/invoice.html"
	Data         interface{}
	Options      *GeneratePDFOptions
}
