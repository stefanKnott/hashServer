package main

import (
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"hash/fnv"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type statistics struct {
	Total             int `json:"total"`
	totalDuration     time.Duration
	AverageDurationMs float32 `json:"average"`
}

type env struct {
	requestStatus chan string       //used for tracking open requests
	idPwHashTable map[uint32]string //id:password hash table
	shutdown      bool
	remaining     int
	Stats         statistics
}

const (
	GET      string = "GET"
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

//return the average duration of hash requests in milliseconds
func calculateaverageDurationMs(count int, totalDuration time.Duration) float32 {
	return float32(totalDuration.Nanoseconds()) / float32(1000000) / float32(count)
}

func (env *env) processHashRequest(w http.ResponseWriter, r *http.Request) {
	env.requestStatus <- PROCESS
	var hash string
	var hashCode uint32

	env.Stats.Total++
	start := time.Now()
	switch r.Method {
	case GET:
		id, err := strconv.Atoi(r.URL.Query().Get("id"))
		if err != nil {
			//write invalid to requester
			w.WriteHeader(http.StatusNotImplemented)
		}
		//get the hash password
		w.Write([]byte(env.idPwHashTable[uint32(id)]))
		return
	case POST:
		//get string from request
		r.ParseForm()
		hash = getHash(r.Form.Get("password"))
		hashCode = getHashCode(hash)

		str := strconv.Itoa(int(hashCode))
		w.Write([]byte(str))
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	env.requestStatus <- FINISHED
	env.Stats.totalDuration += time.Since(start)
	env.Stats.AverageDurationMs = calculateaverageDurationMs(env.Stats.Total, env.Stats.totalDuration)

	//per requirements..stall for 5 seconds
	time.Sleep(5 * time.Second)
	env.idPwHashTable[hashCode] = hash
}

func getHashCode(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
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

func (env *env) processStatsRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case GET:
		b, err := json.Marshal(env.Stats)
		if err != nil {
			panic(err)
		}
		w.Write(b)
		return
	}
}

func main() {
	env := new(env)
	//buffer 100 requests, should be plenty for simple hash
	env.requestStatus = make(chan string, 100)
	env.idPwHashTable = make(map[uint32]string)
	http.HandleFunc("/hash/", env.processHashRequest)
	http.HandleFunc("/shutdown/", env.processShutdownRequest)
	http.HandleFunc("/stats/", env.processStatsRequest)
	// http.HandleFunc("/hashID", env.processHashRequestByID)
	go env.graceful()
	log.Println("Server listeniing for requests on 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
