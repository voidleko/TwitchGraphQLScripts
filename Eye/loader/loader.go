package loader

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type Error struct {
	Message string `json:"message"`
}

type ApiResponse struct {
	Errors []Error `json:"errors"`
}

const (
	ApiUrl      = "https://gql.twitch.tv/gql"
	ApiClientID = "kd1unb4b3q4t58fwlpcbzcbnm76a8fp"
	MaxRetries  = 5
	RetryWait   = 3 * time.Second
)

func LoadOnce(body io.Reader) ([]byte, error) {
	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	req, err := http.NewRequest("POST", ApiUrl, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "text/plain;charset=UTF-8")
	req.Header.Set("Client-ID", ApiClientID)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid status code = %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed read body: %w", err)
	}

	var response ApiResponse
	if err := json.Unmarshal([]byte(respBody), &response); err != nil {
		return nil, fmt.Errorf("failed unmarshal body: %w", err)
	}

	if len(response.Errors) > 0 {
		log.Printf("Got API errors")
		for index, err := range response.Errors {
			log.Printf("%d) %s", index, err.Message)
		}
		return nil, fmt.Errorf("got api errors")
	}

	return respBody, nil
}

func LoadWithRetries(body io.Reader) ([]byte, error) {
	for attempt := 0; attempt < MaxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("Retry %d/%d after %v", attempt, MaxRetries, RetryWait)
			time.Sleep(RetryWait)
		}

		response, err := LoadOnce(body)
		if err != nil {
			fmt.Printf("[ERROR] Failed send api request: %v", err)
			continue
		}

		return response, nil
	}

	return nil, fmt.Errorf("failed to load with retries")
}

func LoadUntilOk(body io.Reader) []byte {
	for attempt := 0; ; attempt++ {
		if attempt > 0 {
			log.Printf("Retry %d/%d after %v", attempt, MaxRetries, RetryWait)
			time.Sleep(RetryWait)
		}

		response, err := LoadOnce(body)
		if err != nil {
			fmt.Printf("[ERROR] Failed send api request: %v", err)
			continue
		}

		return response
	}
}
