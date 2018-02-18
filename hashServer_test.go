package main

import (
	"testing"
	"time"
)

const (
	password         = "angryMonkey"
	expectedHash     = "ZEHhWB65gUlzdVwtDQArEyx-KVLzp_aTaRaPlBzYRIFj6vjFdqEb0Q5B8zVKCZ0vKbZPZklJz0Fd7su2A-gf7Q=="
	expectedHashCode = 3968902113
)

func TestSHA512Hash(t *testing.T) {
	hashString := getHash(password)

	if hashString != expectedHash {
		t.Fatalf("Did not receive expected hash, got : %v\n", hashString)
	}
}

func TestCalculatAvgDuration(t *testing.T) {
	env := new(env)
	env.Stats.Total = 5
	env.Stats.totalDuration = time.Since(time.Now().Add(-5 * time.Minute))

	ms := calculateaverageDurationMs(env.Stats.Total, env.Stats.totalDuration)

	if ms != 60000 {
		t.Fatal("Did not receive expected value, got: ", ms)
	}
}

func TestGetHashCode(t *testing.T) {
	code := getHashCode(expectedHash)
	if code != expectedHashCode {
		t.Fatal("Did not receive expected value, got: ", code)
	}
}
