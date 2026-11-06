package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spitgranger/personal-network-monitor/internal/checks"
)

const (
	HTML = iota
	JSON
	XML
	OTHER
)

func isValidUrl(urlToTest string) bool {
	_, err := url.ParseRequestURI(urlToTest)
	if err != nil {
		return false
	}
	u, err := url.Parse(urlToTest)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("Agent: ")
	urlPtr := flag.String("url", "http://google.ca", "The Url to request")
	timeoutPtr := flag.Int64("timeout", 1000, "Timeout for http request in milliseconds")
	flag.Parse()
	if !isValidUrl(*urlPtr) {
		log.Fatal("Provided Url is not valid")
	}
	if *timeoutPtr <= 0 {
		log.Fatal("Invalid timeout length")
	}
	ctx, stopSignal := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopSignal()
	timeoutCtx, stop := context.WithTimeout(ctx, time.Duration(*timeoutPtr)*time.Millisecond)
	defer stop()
	client := &http.Client{}
	result := checks.CheckHttp(timeoutCtx, client, *urlPtr)
	if err := json.NewEncoder(os.Stdout).Encode(result); err != nil {
		fmt.Println("Failed to encode output", err)
	}
}
