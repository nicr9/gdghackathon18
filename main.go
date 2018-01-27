package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
)

// -----

type Beacon struct {
	UUID       string `json:"uuid"`
	MAC        string `json:"mac"`
	Name       string `json:"name"`
	SessionURL string `json:"session-url,omitempty"`
}

func (b *Beacon) Register() {
	uri := fmt.Sprintf("/session/%s", b.UUID)
	session, err := url.Parse(uri)
	if err != nil {
		b.SessionURL = ""
	} else {
		b.SessionURL = session.Path
	}
}

// -----

type ErrorResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
}

func Homepage(w http.ResponseWriter, r *http.Request) {
	resp := ErrorResponse{Error: true, Message: "Endpoint not supported"}
	json.NewEncoder(w).Encode(resp)
}

// -----

type FindRequest struct {
	Beacon Beacon `json:"beacon"`
}

type FindResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Beacon  Beacon `json:"beacon"`
}

func FindBeacon(w http.ResponseWriter, r *http.Request) {
	// TODO: Ensure this endpoint only accepts POSTs
	var req FindRequest

	// Check request for body
	if r.Body == nil {
		resp := ErrorResponse{Error: true, Message: "Beacon details required"}
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Marshal body into a request obj
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		resp := ErrorResponse{Error: true, Message: err.Error()}
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Do something with the request obj
	req.Beacon.Register()
	log.Println(req)

	// Return a response
	resp := FindResponse{Error: false, Message: "Session created!", Beacon: req.Beacon}
	json.NewEncoder(w).Encode(resp)
}

// -----

func main() {
	http.HandleFunc("/", Homepage)
	http.HandleFunc("/find/", FindBeacon)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
