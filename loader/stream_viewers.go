package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"slices"
	"strings"
	"time"
)

func prepareStreamViewersBody(id string) (io.Reader, error) {
	variables := ViewersPageRequest{
		Query: ViewersPageQuery,
		Variables: ViewersPageVariables{
			ID:    &id,
			Login: nil,
		},
	}
	variablesAsString, err := json.Marshal(variables)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal variables: %w", err)
	}
	body := strings.NewReader(string(variablesAsString))
	return body, nil
}

func loadStreamViewersPage(id string) (ViewersPageUserData, error) {
	log.Printf("Loadinng viewers page")

	body, err := prepareStreamViewersBody(id)
	if err != nil {
		log.Printf("Failed to prepare body: %v", err)
		return ViewersPageUserData{}, errors.New("failed to load")
	}

	respBody, err := loadWithRetries(body)
	if err != nil {
		log.Printf("Failed to load: %v", err)
		return ViewersPageUserData{}, errors.New("failed to load")
	}

	var viewersPage ViewersPageResponse
	if err := json.Unmarshal([]byte(respBody), &viewersPage); err != nil {
		log.Printf("Failed unmarshal body: err = %v, body = %s", err, respBody)
		return ViewersPageUserData{}, errors.New("failed to load")
	}

	if len(viewersPage.Errors) > 0 {
		log.Printf("Got some errors")
		for index, err := range viewersPage.Errors {
			log.Printf("%d) %s", index, err.Message)
		}
		return ViewersPageUserData{}, errors.New("failed to load")
	}

	return viewersPage.Data.User, nil
}

func loadStreamViewers(id string) ViewersPageUserData {
	var initialResult ViewersPageUserData
	var err error

	log.Printf("Loading initial viewers page")
	for {
		initialResult, err = loadStreamViewersPage(id)
		if err != nil {
			log.Printf("Failed to load initial page")
			time.Sleep(RetryWait)
			continue
		}
		break
	}

	for {
		log.Printf("Loading update viewers page")
		var update ViewersPageUserData
		for {
			update, err = loadStreamViewersPage(id)
			if err != nil {
				log.Printf("Failed to load initial page")
				time.Sleep(RetryWait)
				continue
			}
			break
		}
		before := len(initialResult.Channel.Chatters.Viewers)
		log.Printf("Update page: before: %d/%d", before, initialResult.Channel.Chatters.Count)
		joinedViewers := initialResult.Channel.Chatters.Viewers
		for _, newViewer := range update.Channel.Chatters.Viewers {
			if !slices.Contains(joinedViewers, newViewer) {
				joinedViewers = append(joinedViewers, newViewer)
			}
		}
		initialResult.Channel.Chatters.Viewers = joinedViewers
		after := len(initialResult.Channel.Chatters.Viewers)
		log.Printf("Update page: after: %d/%d", after, initialResult.Channel.Chatters.Count)
		if after == before {
			break
		}
	}
	return initialResult
}
