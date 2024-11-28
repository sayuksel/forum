package main
import (
    "log"
    "net/http"
    "forum/errors"
    "forum/handlers"
)
func handler(w http.ResponseWriter, r *http.Request) {
    switch r.URL.Path {
    case "/":
        handlers.Home(w, r)
    case "/login":
        handlers.Login(w, r)
    case "/logout":
        handlers.Logout(w, r)
    case "/register":
        handlers.RegisterUser(w, r)
    case "/post":
        handlers.Post(w, r)
    case "/postpage":
        handlers.PostPage(w, r)
    case "/submit_post":
        handlers.SubmitPost(handlers.GetDB(), w, r)
    case "/submit_comment":
        handlers.SubmitComment(handlers.GetDB(), w, r)
    case "/like_post":
        handlers.LikePost(handlers.GetDB(), w, r)
    case "/like_comment":
        handlers.LikeComment(handlers.GetDB(), w, r)
    case "/dislike_post":
        handlers.DislikePost(handlers.GetDB(), w, r)
    case "/dislike_comment":
        handlers.DislikeComment(handlers.GetDB(), w, r)
    case "/delete_post":
        handlers.DeletePost(handlers.GetDB(), w, r)
    case "/delete_comment":
        handlers.DeleteComment(handlers.GetDB(), w, r)
    default:
        errors.Error404(w, r)
    }
}
func main() {
    // Initialize the database
    handlers.InitDB()
    // Serve static files
    fs := http.FileServer(http.Dir("static"))
    http.Handle("/static/", http.StripPrefix("/static/", fs))
    // Use the handler function for routing
    http.HandleFunc("/", handler)
    // Start the server
    log.Println("Starting server on localhost:8080")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatalf("Server failed to start: %v", err)
    }
}