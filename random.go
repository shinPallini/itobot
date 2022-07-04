package main

import (
	"math/rand"
	"time"
)

func Random() int {
	rand.Seed(time.Now().UnixNano())
	num := rand.Intn(100) + 1
	return num

}
