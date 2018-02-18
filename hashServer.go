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
	"sync"
	"time"
)

type hashObj struct {
	code uint32
	hash string
}

type statistics struct {
	Total             int `json:"total"`
	totalDuration     time.Duration
	AverageDurationMs float32 `json:"average"`
}

type env struct {
	requestStatus chan string //used for tracking open requests
	hashChan      chan hashObj
	idPwHashTable map[uint32]string //id:password hash table
	shutdown      bool
	remaining     int
	Stats         statistics
	lock          *sync.Mutex
}

const (
	GET      string = "GET"
	POST     string = "POST"
	PROCESS  string = "PROCESSING"
	FINISHED string = "FINSIHED"
	SHUTDOWN string = "SHUTDOWN"
)

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

//set hashtable information after a 5s delay
func (env *env) processHashSets() {
	for hashObj := range env.hashChan {
		time.Sleep(5 * time.Second)
		env.lock.Lock()
		env.idPwHashTable[hashObj.code] = hashObj.hash
		env.lock.Unlock()
	}
}

//process requests to our /hash/ API
func (env *env) processHashRequest(w http.ResponseWriter, r *http.Request) {
	env.requestStatus <- PROCESS
	var hash string
	var hashCode uint32

	start := time.Now()
	switch r.Method {
	case GET:
		id, err := strconv.Atoi(r.URL.Query().Get("id"))
		if err != nil {
			//write invalid to requester
			w.WriteHeader(http.StatusNotImplemented)
		}
		//get the hash password & write back to client
		env.lock.Lock()
		w.Write([]byte(env.idPwHashTable[uint32(id)]))
		env.lock.Unlock()
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

	//check to ensure we are setting hash
	if hash != "" {
		env.hashChan <- hashObj{code: hashCode, hash: hash}
	}

	env.lock.Lock()
	env.Stats.Total++
	env.Stats.totalDuration += time.Since(start)
	env.Stats.AverageDurationMs = calculateaverageDurationMs(env.Stats.Total, env.Stats.totalDuration)
	env.lock.Unlock()
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
	env.lock = new(sync.Mutex)
	//buffer 100 requests, should be plenty for simple hash
	env.requestStatus = make(chan string, 100)
	env.hashChan = make(chan hashObj, 10)
	env.idPwHashTable = make(map[uint32]string)
	http.HandleFunc("/hash/", env.processHashRequest)
	http.HandleFunc("/shutdown/", env.processShutdownRequest)
	http.HandleFunc("/stats/", env.processStatsRequest)
	go env.processHashSets()
	go env.graceful()
	log.Println("Server listeniing for requests on 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
