package main

import (
	"log"
	"time"
)

func main() {
	for {
		log.Println("Pinging")
		time.Sleep(time.Second)
	}
}
