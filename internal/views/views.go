package views

import (
	"html/template"
	"log"
)

// Templates holds all parsed HTML templates for rendering pages throughout the app.
var Templates *template.Template

// InitTemplates parses all HTML templates in the web/templates directory.
// Should be called once at startup to initialize the Templates variable.
// If template parsing fails, the app will log a fatal error and exit.
func InitTemplates() {
	var err error
	// Parse all .html files in the web/templates directory
	Templates, err = template.ParseGlob("web/templates/*.html")
	if err != nil {
		// If parsing fails, log the error and stop the application
		log.Fatal("failed to parse templates:", err)
	}
}
