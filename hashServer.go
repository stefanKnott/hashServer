package main

import (
	"crypto/sha512"
	"encoding/base64"
	"log"
	"net/http"
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
	w.Write([]byte(hash))

}

func main() {
	// bind := flag.Int("bind", 8000, "port to serve")
	http.HandleFunc("/hash", processHashRequest)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
