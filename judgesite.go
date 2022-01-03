package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func JudgeSiteController(w http.ResponseWriter, r *http.Request) {
	urlArr := r.URL.Query()["url"]
	fmt.Println(urlArr)
	
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

	for site, logic := range LogicMap {
		kind, id := logic.CheckUrl(*targetUrl)
		if id != "" {
			bodyStr, _ := json.Marshal(SiteId{Kind: kind, Site: site, Id: id})
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(bodyStr)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "not found.")
}
