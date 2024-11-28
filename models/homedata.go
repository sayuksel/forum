package models
type HomePageData struct {
	User            User
	IsAuthenticated bool
	Posts           []Posts
}