package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

func ShowPicByAuthorController(w http.ResponseWriter, r *http.Request) {
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

	countArr := r.URL.Query()["count"]
	var count int
	if len(countArr) < 1 {
		count = 0
	} else {
		var err error
		count, err = strconv.Atoi(countArr[0])
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "count must be integer.")
			return
		}
	}

	// TODO select service and call
	profiles, err := GetPicByAuthor(site, id, count)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "invalid site")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	bodyStr, _ := json.Marshal(profiles)
	w.Write(bodyStr)
}

func GetPicByAuthor(site string, id string, count int) (ProfilesTags, error) {
	if logic, exists := LogicMap[site]; exists {
		return logic.GetPicsOfAccount(id, count)
	}
	return ProfilesTags{}, errors.New("invalid site")
}
