package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func DownloadController(w http.ResponseWriter, r *http.Request) {
	modified := r.Header["If-Modified-Since"]
	if len(modified) > 0 {
		fmt.Println("cached")
		w.Header().Set("Cache-Control", "public, max-age=,31536000 immutable")
		w.Header().Set("Last-Modified", modified[0])
		w.WriteHeader(http.StatusNotModified)
		return
	}

	fmt.Println("no cached")
	urlArr := r.URL.Query()["url"]
	if len(urlArr) < 1 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "url not specified.")
		return
	}
	targetUrl, err := url.ParseRequestURI(urlArr[0])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "url malformed.")
		return
	}

	fmt.Println(targetUrl.String())
	for _, logic := range LogicMap {
		_, id := logic.CheckUrl(*targetUrl)
		if id != "" {
			result, err := logic.Download(*targetUrl)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprintf(w, "not found.")
			}
			w.Header().Set("Content-Type", result.ContentType)
			w.Header().Set("Cache-Control", "public, max-age=,31536000 immutable")
			w.Header().Set("Last-Modified", result.LastModified)
			w.WriteHeader(http.StatusOK)

			defer result.Body.Close()

			io.Copy(w, result.Body)
			return
		}
	}

	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprintf(w, "not supported URL.")
}
