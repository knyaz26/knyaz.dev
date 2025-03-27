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
	Template              string     `toml:"template"`
	Title                 string     `toml:"title"`
	Subtitle              string     `toml:"subtitle,omitempty"`
	PostDate              string     `toml:"post_date,omitempty"`
	Description           string     `toml:"description,omitempty"`
	Pictures              []string   `toml:"pictures,omitempty"`
	FuturePlans           string     `toml:"future_plans,omitempty"`
	Resources             []Resource `toml:"resources,omitempty"`
	Logs                  string     `toml:"logs,omitempty"`
	TechnologiesUsed      string     `toml:"technologies_used,omitempty"`
	IframeSrc             string     `toml:"iframe_src,omitempty"`
	IframeWidth           string     `toml:"iframe_width,omitempty"`
	IframeHeight          string     `toml:"iframe_height,omitempty"`
	IframeAllowFullscreen bool       `toml:"iframe_allowfullscreen,omitempty"`
}

var templates *template.Template
var funcMap = template.FuncMap{
	"default": func(def string, val interface{}) string {
		s := fmt.Sprintf("%v", val)
		if s == "" || s == "<nil>" {
			return def
		}
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
	templates.ExecuteTemplate(w, tmplName, data)
}

func staticPageHandler(tmplName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		renderTemplate(w, tmplName, nil)
	}
}

func contentPageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		urlPath := strings.Trim(r.URL.Path, "/")
		parts := strings.Split(urlPath, "/")

		if len(parts) == 1 && (parts[0] == "blog" || parts[0] == "projects") {
			indexTemplate := parts[0] + ".html"
			log.Printf("Rendering index path '%s' using template '%s'", r.URL.Path, indexTemplate)
			renderTemplate(w, indexTemplate, nil)
			return
		}

		cleanedPath := filepath.Clean(urlPath)
		if cleanedPath == "." || strings.Contains(cleanedPath, "..") {
			return
		}

		tomlPath := filepath.Join("content", cleanedPath+".toml")
		tomlBytes, err := os.ReadFile(tomlPath)
		if err != nil {
			if os.IsNotExist(err) {
				log.Printf("Content file not found for path '%s': %s", r.URL.Path, tomlPath)
				http.NotFound(w, r)
			} else {
				log.Printf("Non-fatal error reading content file %s: %v", tomlPath, err)
				// Production should likely return 500 here
			}
			return
		}

		var pageData PageContent
		_, err = toml.Decode(string(tomlBytes), &pageData)
		if err != nil {
			log.Printf("Non-fatal error decoding TOML file %s: %v", tomlPath, err)
			// Production should likely return 500 here
			return // Or attempt to proceed, risky
		}


		templateToRender := pageData.Template
		if templateToRender == "" {
			log.Printf("Template not specified in TOML: %s", tomlPath)
			// Production should return 500 here
			return
		}

		log.Printf("Rendering content path '%s' using template '%s' with data from '%s'", r.URL.Path, templateToRender, tomlPath)
		renderTemplate(w, templateToRender, pageData)
	}
}

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.HandleFunc("/", staticPageHandler("home.html"))
	http.HandleFunc("/contact", staticPageHandler("contact.html"))
	http.HandleFunc("/resources", staticPageHandler("resources.html"))

	http.HandleFunc("/blog/", contentPageHandler())
	http.HandleFunc("/projects/", contentPageHandler())

	port := ":10000"
	log.Printf("Server starting on http://localhost%s", port)
	http.ListenAndServe(port, nil)
}
