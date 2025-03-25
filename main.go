package main

import (
    "fmt"
    "html/template"
    "net/http"
    "path/filepath"
)

func renderTemplate(w http.ResponseWriter, tmpl string) {
    layoutPath := filepath.Join("templates", "layout", "navbar.html")
    footerPath := filepath.Join("templates", "layout", "footer.html")
    headerPath := filepath.Join("templates", "layout", "header.html")
    tmplPath := filepath.Join("templates", tmpl)
    t, err := template.ParseFiles(tmplPath, layoutPath, footerPath, headerPath)
    if err != nil {
        http.Error(w, fmt.Sprintf("Error parsing template: %v", err), http.StatusInternalServerError)
        return
    }
    if err := t.Execute(w, nil); err != nil {
        http.Error(w, fmt.Sprintf("Error executing template: %v", err), http.StatusInternalServerError)
    }
}

func handler(tmpl string) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        renderTemplate(w, tmpl)
    }
}

func main() {
    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

    // Main pages
    http.HandleFunc("/", handler("home.html"))
    http.HandleFunc("/contact", handler("contact.html"))
    http.HandleFunc("/blog", handler("blog.html"))
    http.HandleFunc("/projects", handler("projects.html"))
    http.HandleFunc("/resources", handler("resources.html"))

    // Secondary pages
    http.HandleFunc("/blog/website-deployed", handler("website-deployed.html"))
    http.HandleFunc("/blog/aa-defence", handler("aa-defence.html"))
    http.HandleFunc("/blog/jetpacker", handler("jetpacker.html"))
    http.HandleFunc("/blog/roblox-minesweeper", handler("roblox-minesweeper.html"))

    http.ListenAndServe(":10000", nil)
}
