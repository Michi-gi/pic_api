package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi"
)

func main() {
	router := chi.NewRouter()

	router.Get("/picprofile", ShowPicProfileController)
	router.Get("/authorprofile", ShowAuthorProfileController)
	router.Get("/picbyauthor", ShowPicByAuthorController)
	router.Get("/judgesite", JudgeSiteController)
	router.Get("/download", DownloadController)

	port := os.Getenv("PORT")
	fmt.Print("port: " + port)
	http.ListenAndServe(port, router)
}
