package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type Resource struct {
	Text string
	URL  string
}

type PageContent struct {
	Title            string     `toml:"title"`
	Subtitle         string     `toml:"subtitle,omitempty"`
	PostDate         string     `toml:"post_date,omitempty"`
	Description      string     `toml:"description,omitempty"`
	Pictures         []string   `toml:"pictures,omitempty"`
	TechnologiesUsed string     `toml:"technologies_used,omitempty"`
	FuturePlans      string     `toml:"future_plans,omitempty"`
	Resources        []Resource `toml:"resources,omitempty"`
	Logs             string     `toml:"logs,omitempty"`
}


var templates *template.Template
var funcMap = template.FuncMap{
	"default": func(def string, val interface{}) string {
		s := fmt.Sprintf("%v", val)
		if s == "" || s == "<nil>" { return def }
		return s
	},
}

func init() {
	templates = template.New("").Funcs(funcMap)
	templates = template.Must(templates.ParseGlob("templates/*.html"))
	templates = template.Must(templates.ParseGlob("templates/layout/*.html"))
	log.Println("Templates parsed successfully")
}

func renderTemplate(w http.ResponseWriter, tmplName string, data interface{}) {
	if tmpl := templates.Lookup(tmplName); tmpl == nil {
		log.Printf("Error: Template '%s' not found", tmplName)
		http.Error(w, fmt.Sprintf("Internal error: Template '%s' not found", tmplName), http.StatusInternalServerError)
		return
	}
	err := templates.ExecuteTemplate(w, tmplName, data)
	if err != nil {
		log.Printf("Error executing template %s: %v", tmplName, err)
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
	}
}

func staticPageHandler(tmplName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		renderTemplate(w, tmplName, nil)
	}
}

func contentPageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		urlPath := strings.TrimPrefix(r.URL.Path, "/")
		cleanedPath := filepath.Clean(urlPath)
		if cleanedPath == "." || strings.Contains(cleanedPath, "..") {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}

		tomlPath := filepath.Join("content", cleanedPath+".toml")
		tomlBytes, err := os.ReadFile(tomlPath)
		if err != nil {
			if os.IsNotExist(err) {
				log.Printf("Content file not found: %s", tomlPath)
				http.NotFound(w, r)
			} else {
				log.Printf("Error reading content file %s: %v", tomlPath, err)
				http.Error(w, "Error reading content file", http.StatusInternalServerError)
			}
			return
		}

		var pageData PageContent
		_, err = toml.Decode(string(tomlBytes), &pageData)
		if err != nil {
			log.Printf("Error decoding TOML file %s: %v", tomlPath, err)
			http.Error(w, "Error processing content file", http.StatusInternalServerError)
			return
		}

		templateToRender := "blog-post.html"

		log.Printf("Rendering path '%s' using template '%s' with data from '%s'", r.URL.Path, templateToRender, tomlPath)
		renderTemplate(w, templateToRender, pageData)
	}
}

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.HandleFunc("/", staticPageHandler("home.html"))
	// Add other static routes like /contact, /resources if they exist

	http.HandleFunc("/blog/", contentPageHandler())
	http.HandleFunc("/projects/", contentPageHandler())
	// Add any other base paths that should use TOML + post.html

	port := ":10000"
	log.Printf("Server starting on http://localhost%s", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
