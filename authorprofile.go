package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

func ShowAuthorProfileController(w http.ResponseWriter, r *http.Request) {
	siteArr := r.URL.Query()["site"]
	if len(siteArr) < 1 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "site not specified.")
		return
	}
	site := siteArr[0]

	idArr := r.URL.Query()["id"]
	if len(idArr) < 1 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "id not specified.")
		return
	}
	id := idArr[0]

	// TODO select service and call
	profile, err := GetAuthorProfile(site, id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "invalid site")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	bodyStr, _ := json.Marshal(profile)
	w.Write(bodyStr)
}

func GetAuthorProfile(site string, id string) (AccountProfile, error) {
	if logic, exists := LogicMap[site]; exists {
		return logic.GetAccountProfile(id)
	}
	return AccountProfile{}, errors.New("invalid site")
}
