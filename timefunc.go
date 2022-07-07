package main

import (
	"math/rand"
	"time"
)

func Random(max int) int {
	rand.Seed(time.Now().UnixNano())
	num := rand.Intn(max) + 1
	return num

}

func GetNow() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05Z07:00")
}
