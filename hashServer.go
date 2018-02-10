package main

import (
	"crypto/sha512"
	"encoding/base64"
	"log"
	"net/http"
	"os"
	"time"
)

type env struct {
	requestStatus chan string //used for tracking open requests
	shutdown      bool
	remaining     int
}

const (
	POST     string = "POST"
	PROCESS  string = "PROCESSING"
	FINISHED string = "FINSIHED"
	SHUTDOWN string = "SHUTDOWN"
)

//solution for handle rest of requests on shutdown could include a buffered channel

func getHash(password string) string {
	hasher := sha512.New()
	hasher.Write([]byte(password))
	sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))

	return sha
}

func (env *env) processHashRequest(w http.ResponseWriter, r *http.Request) {
	env.requestStatus <- PROCESS

	//filter out any request that is not a POST
	if r.Method != POST {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	//get string from request
	r.ParseForm()
	hash := getHash(r.Form.Get("password"))

	//per requirements..stall for 5 seconds
	time.Sleep(5 * time.Second)
	w.Write([]byte(hash))
	env.requestStatus <- FINISHED
}

func (env *env) processShutdownRequest(w http.ResponseWriter, r *http.Request) {
	//TODO: refactor this out
	if r.Method != POST {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	//write 200 back to requester
	w.WriteHeader(http.StatusOK)
	//send shutdown request
	env.requestStatus <- SHUTDOWN
}

func (env *env) graceful() {
	for req := range env.requestStatus {
		if req == PROCESS && env.shutdown == false {
			env.remaining++
		} else if req == FINISHED {
			env.remaining--
		} else if req == SHUTDOWN {
			env.shutdown = true
		}

		if env.shutdown && env.remaining == 0 {
			os.Exit(0)
		}
	}
}

func main() {
	env := new(env)
	//buffer 100 requests, should be plenty for simple hash
	env.requestStatus = make(chan string, 100)

	http.HandleFunc("/hash", env.processHashRequest)
	http.HandleFunc("/shutdown", env.processShutdownRequest)
	go env.graceful()
	log.Println("Server listeniing for requests on 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
