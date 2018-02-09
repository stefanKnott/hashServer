package main

import (
	"crypto/sha512"
	"encoding/base64"
	"log"
	"net/http"
	"time"
)

type hashReq struct {
	password string
}

func getHash(password string) string {
	hasher := sha512.New()
	hasher.Write([]byte(password))
	sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))

	return sha
}

func processHashRequest(w http.ResponseWriter, r *http.Request) {
	//get string from request
	r.ParseForm()
	hash := getHash(r.Form.Get("password"))
	//per requirements..stall for 5 seconds
	time.Sleep(5 * time.Second)
	w.Write([]byte(hash))

}

func main() {
	http.HandleFunc("/hash", processHashRequest)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
