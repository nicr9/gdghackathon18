package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

// -----

type Beacon struct {
	UUID       string `json:"uuid"`
	MAC        string `json:"mac"`
	Name       string `json:"name"`
	SessionURL string `json:"session_url,omitempty"`
}

func (b *Beacon) Register() {
	sessionPath := fmt.Sprintf("/session/%s", b.UUID)
	sessionUrl, err := url.Parse(sessionPath)
	if err != nil {
		b.SessionURL = ""
	} else {
		b.SessionURL = sessionUrl.Path

		wsPath := fmt.Sprintf("/wb/%s", b.UUID)
		wsUrl, _ := url.Parse(wsPath)
		session := newSession()
		http.Handle(wsUrl, session)
		go session.run()
	}
}

// -----

type ErrorResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
}

func Homepage(w http.ResponseWriter, r *http.Request) {
	file, _ := os.Open("./index.html")

	if _, err := io.Copy(w, file); err != nil {
		log.Fatal(err)
	}
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

type client struct {
	socket  *websocket.Conn
	send    chan []byte
	session *session
}

func (c *client) read() {
	for {
		if _, body, err := c.socket.ReadMessage(); err != nil {
			c.session.forward <- body
		} else {
			break
		}
	}

	c.socket.Close()
}

func (c *client) write() {
	for msg := range c.send {
		if err := c.socket.WriteMessage(websocket.TextMessage, msg); err != nil {
			break
		}
	}

	c.socket.Close()
}

type session struct {
	forward chan []byte
	join    chan *client
	leave   chan *client
	clients map[*client]bool
}

func newSession() *session {
	return &session{
		forward: make(chan []byte),
		join:    make(chan *client),
		leave:   make(chan *client),
		clients: make(map[*client]bool),
	}
}

func (s *session) run() {
	for {
		select {
		case client := <-s.join:
			s.clients[client] = true
		case client := <-s.leave:
			delete(s.clients, client)
			close(client.send)
		case msg := <-s.forward:
			for client := range s.clients {
				select {
				case client.send <- msg:
					//
				default:
					delete(s.clients, client)
					close(client.send)
				}
			}
			//
		}
	}
}

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{ReadBufferSize: socketBufferSize, WriteBufferSize: socketBufferSize}

func (s *session) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	socket, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal("Failed to establish a websocket:", err)
		return
	}

	client := &client{
		socket:  socket,
		send:    make(chan []byte, messageBufferSize),
		session: s,
	}

	s.join <- client
	defer func() { s.leave <- client }()
	go client.write()
	client.read()
}

// -----

func main() {
	http.HandleFunc("/", Homepage)
	http.HandleFunc("/find/", FindBeacon)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
