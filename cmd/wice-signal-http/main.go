package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"

	log "github.com/sirupsen/logrus"
	"riasc.eu/wice/pkg/backend"
	"riasc.eu/wice/pkg/crypto"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

var offers map[crypto.Key]map[crypto.Key]backend.Offer

func candidateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ourPKStr := vars["ours"]
	theirPKStr := vars["theirs"]

	ourPKStrUnescaped, _ := url.PathUnescape(ourPKStr)
	theirPKStrUnescaped, _ := url.PathUnescape(theirPKStr)

	var err error
	var ourPK, theirPK crypto.Key

	ourPK, err = crypto.ParseKey(ourPKStrUnescaped)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse key: %s", err), http.StatusBadRequest)
	}

	theirPK, err = crypto.ParseKey(theirPKStrUnescaped)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse key: %s", err), http.StatusBadRequest)
	}

	// Create missing maps and slice
	if _, ok := offers[ourPK]; !ok {
		offers[ourPK] = make(map[crypto.Key]backend.Offer)
	}

	if r.Method == "POST" {
		dec := json.NewDecoder(r.Body)
		var offer backend.Offer
		err := dec.Decode(&offer)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		offers[ourPK][theirPK] = offer
	} else if r.Method == "GET" {
		offer, ok := offers[ourPK][theirPK]
		if !ok {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		enc := json.NewEncoder(w)
		enc.Encode(offer)

	} else if r.Method == "DELETE" {
		_, ok := offers[ourPK][theirPK]
		if !ok {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		delete(offers[ourPK], theirPK)
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/offers/{ours}/{theirs}", candidateHandler)

	lr := handlers.LoggingHandler(os.Stdout, r)

	http.Handle("/", lr)

	offers = make(map[crypto.Key]map[crypto.Key]backend.Offer)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
