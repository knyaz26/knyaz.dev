package main

import (
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
	Logs                  []Resource `toml:"logs,omitempty"`
	TechnologiesUsed      string     `toml:"technologies_used,omitempty"`
	IframeSrc             string     `toml:"iframe_src,omitempty"`
	IframeWidth           string     `toml:"iframe_width,omitempty"`
	IframeHeight          string     `toml:"iframe_height,omitempty"`
	IframeAllowFullscreen bool       `toml:"iframe_allowfullscreen,omitempty"`
}

var templates *template.Template

func defaultFunc(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func init() {
	templates = template.New("").Funcs(template.FuncMap{
		"default": defaultFunc,
	})
	templates = template.Must(templates.ParseGlob("templates/*.html"))
	templates = template.Must(templates.ParseGlob("templates/layout/*.html"))
}

func renderTemplate(w http.ResponseWriter, tmplName string, data interface{}) {
	err := templates.ExecuteTemplate(w, tmplName, data)
	if err != nil {
		log.Printf("!!! TEMPLATE EXECUTION ERROR for '%s': %v", tmplName, err)
	}
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
				http.NotFound(w, r)
			}
			return
		}

		var pageData PageContent
		_, err = toml.Decode(string(tomlBytes), &pageData)
		if err != nil {
			return
		}

		if pageData.Template == "" {
			return
		}

		renderTemplate(w, pageData.Template, pageData)
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
	http.ListenAndServe(port, nil)
}

