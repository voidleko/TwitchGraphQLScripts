package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
)

func prepareUserDataBody(login string) (io.Reader, error) {
	variables := UserDataRequest{
		Query: UserQuery,
		Variables: UserDataVariables{
			ID:    nil,
			Login: &login,
		},
	}
	variablesAsString, err := json.Marshal(variables)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal variables: %w", err)
	}
	body := strings.NewReader(string(variablesAsString))
	return body, nil
}

func loadUserData(login string) UserDataResponse {
	body, err := prepareUserDataBody(login)
	if err != nil {
		log.Panicf("failed prepare load stream page body: %v", err)
	}

	respBody, err := loadWithRetries(body)
	if err != nil {
		log.Panicf("failed to load user data: %v", err)
	}

	var userData UserDataResponseesponse
	if err := json.Unmarshal([]byte(respBody), &userData); err != nil {
		log.Panicf("failed unmarshal user data body: %v", err)
	}

	if len(userData.Errors) > 0 {
		log.Printf("Got API errors")
		for index, err := range userData.Errors {
			log.Printf("%d) %s", index, err.Message)
		}
		log.Panicf("Failed to load user data")
	}

	return userData.Data.User
}
