package infrastructure

import (
	"bytes"
	"fmt"
	"html/template"
	"path/filepath"
)

type TemplateEngine struct {
	templates map[string]*template.Template
}

func NewTemplateEngine(templatesDir string) (*TemplateEngine, error) {
	templates := make(map[string]*template.Template)
	
	// Load all templates from directory
	files, err := filepath.Glob(filepath.Join(templatesDir, "*.html"))
	if err != nil {
		return nil, fmt.Errorf("failed to read templates: %w", err)
	}
	
	for _, file := range files {
		name := filepath.Base(file)
		tmpl, err := template.ParseFiles(file)
		if err != nil {
			return nil, fmt.Errorf("failed to parse template %s: %w", name, err)
		}
		templates[name] = tmpl
	}
	
	return &TemplateEngine{
		templates: templates,
	}, nil
}

func (e *TemplateEngine) Render(templateName string, data interface{}) (string, error) {
	tmpl, ok := e.templates[templateName]
	if !ok {
		return "", fmt.Errorf("template %s not found", templateName)
	}
	
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}
	
	return buf.String(), nil
}

func (e *TemplateEngine) AddTemplate(name, content string) error {
	tmpl, err := template.New(name).Parse(content)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}
	
	e.templates[name] = tmpl
	return nil
}