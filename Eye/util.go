package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	ApiUrl      = "https://gql.twitch.tv/gql"
	ApiClientID = "kd1unb4b3q4t58fwlpcbzcbnm76a8fp"
	MaxRetries  = 5
	RetryWait   = 3 * time.Second
)

func loadWithRetries(body io.Reader) ([]byte, error) {
	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	for attempt := 0; attempt < MaxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("Retry %d/%d after %v", attempt, MaxRetries, RetryWait)
			time.Sleep(RetryWait)
		}

		req, err := http.NewRequest("POST", ApiUrl, body)
		if err != nil {
			log.Printf("Failed to load page: %v", err)
			continue
		}

		req.Header.Set("Content-Type", "text/plain;charset=UTF-8")
		req.Header.Set("Client-ID", ApiClientID)

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Failed to load page: %v", err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			log.Printf("Failed to load page: status = %s", resp.Status)
			continue
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			log.Printf("Failed to read all: $w", err)
			continue
		}

		return respBody, nil
	}

	return []byte{}, fmt.Errorf("failed to load with retries")
}
