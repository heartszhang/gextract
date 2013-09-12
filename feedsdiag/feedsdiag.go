package main

import (
	"encoding/json"
	"fmt"
	"github.com/heartszhang/gextract/feeds"
	"log"
	"net/http"
)

func main() {
	log.Fatal(http.ListenAndServe("localhost:1212", nil))
}

func init() {
	http.HandleFunc("/", index)
	http.HandleFunc("/diag/feeds.json", feeds_diag)
}
func write_json(w http.ResponseWriter, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.Encode(body)
}
func error_json(w http.ResponseWriter, body interface{}, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	enc.Encode(body)
}
func index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "works")
}

func feeds_diag(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	eo := feeds.NewEntryOperator()
	entries, err := eo.TopN(0, 10)
	if err != nil {
		error_json(w, new_error(err), http.StatusBadGateway)
		//		http.Error(w, err, http.StatusBadGateway)
		return
	}
	write_json(w, entries)
	//	index(w, r)
}

type http_error struct {
	Code   int    `code`
	Reason string `reason,omitempty`
}

func new_error(err error) http_error {
	e := http_error{Code: -1, Reason: err.Error()}
	return e
}
