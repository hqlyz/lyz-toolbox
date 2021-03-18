package main

import (
	"fetch-website/server"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	runOnce = true
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: fetch-website some-url")
		return
	}
	webURL := os.Args[1]

	server := server.New(runOnce)
	server.Enqueue(webURL)
	server.Run()

	if runOnce {
		<-server.Ctx.Done()
	} else {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

		<-sigChan
	}

	log.Println("Server is shutting down...")
}
