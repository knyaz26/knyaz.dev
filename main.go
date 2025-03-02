package main

import (
    "html/template"
    "net/http"
)

/*
MANUAL ON HOW TO USE THIS THINGIE:
-declare a variable of a template.
-set up a function.
-parse the html in main function.
-set up a hander function in main.
*/

//declare variables here:
var(
    homeTemplate *template.Template
)

//set up functions here:
func Home(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/html")
    if err := homeTemplate.Execute(w, nil); err != nil {
        panic(err)
    }
}

func main() {
    //parse html in here:
    //make sure to pass in all the parameters includeing named templates.
    homeTemplate, _ = template.ParseFiles("templates/home.html", "templates/navbar.html", "templates/footer.html")

    //set up handler function here:
    //pass in a URL adress and function name.
    http.HandleFunc("/", Home)

    //misc:
    http.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("static"))))
    http.ListenAndServe(":10000", nil)
}

