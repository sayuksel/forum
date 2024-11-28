package errors
import (
	"html/template"
	"net/http"
	"fmt"
)
var templates = template.Must(template.ParseGlob("templates/*.html"))
// executes error400 html template
func Error400(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	err := templates.ExecuteTemplate(w, "400.html", nil)
	if err != nil {
		return
	}
}
// executes error401 html template
func Error401(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusUnauthorized)
	err := templates.ExecuteTemplate(w, "401.html", nil)
	if err != nil {
		return
	}
}
// executes error404 html template
func Error404(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	err := templates.ExecuteTemplate(w, "404.html", nil)
	if err != nil {
		return
	}
}
// executes error405 html template
func Error405(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
	err := templates.ExecuteTemplate(w, "405.html", nil)
	if err != nil {
		return
	}
}
// executes error409 html template
func Error409(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusConflict)
	err := templates.ExecuteTemplate(w, "409.html", nil)
	if err != nil {
		return
	}
}
// executes error500 html template
func Error500(w http.ResponseWriter, r *http.Request) {
    if w.Header().Get("Content-Type") == "" { // Prevent duplicate writes
        w.WriteHeader(http.StatusInternalServerError)
    }
    fmt.Fprintln(w, "500 Internal Server Error")
}