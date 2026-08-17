package streams

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/alebik0/TwitchGraphQLScripts/Eye/loader"
)

type Error struct {
	Message string `json:"message"`
}

type UserSettingsResponse struct {
	PreferredLanguageTag string `json:"preferredLanguageTag"`
}

type UserDataResponse struct {
	ID              string               `json:"id"`
	Login           string               `json:"login"`
	ProfileImageURL string               `json:"profileImageURL"`
	CreatedAt       time.Time            `json:"createdAt"`
	UpdatedAt       *time.Time           `json:"updatedAt"`
	DeletedAt       *time.Time           `json:"deletedAt"`
	Description     string               `json:"description"`
	Settings        UserSettingsResponse `json:"settings"`
}

type UserDataResponseData struct {
	User UserDataResponse `json:"user"`
}

type UserDataResponseesponse struct {
	Data   UserDataResponseData `json:"data"`
	Errors []Error              `json:"errors"`
}

type UserDataVariables struct {
	ID    *string `json:"id"`
	Login *string `json:"login"`
}

type UserDataRequest struct {
	Query     string            `json:"query"`
	Variables UserDataVariables `json:"variables"`
}

type UserData struct {
	ID              uint64
	Login           string
	ProfileImageURL string
	CreatedAt       time.Time
	UpdatedAt       *time.Time
	DeletedAt       *time.Time
	Description     string
	Language        string
}

const (
	UserQuery = `
query fetchUser($id: ID, $login: String) {
  user(id: $id, login: $login, lookupType: ALL) {
    id
    login
    profileImageURL(width: 150)
    createdAt
    updatedAt
    deletedAt
    description
    settings {
      preferredLanguageTag
    }
  }
}
`
)

func prepareBody(login string) (io.Reader, error) {
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

func ForceLoad(login string) UserData {
	body, err := prepareBody(login)
	if err != nil {
		log.Panicf("failed prepare load stream page body: %v", err)
	}

	respBody := loader.LoadUntilOk(body)

	var userData UserDataResponseesponse
	if err := json.Unmarshal([]byte(respBody), &userData); err != nil {
		log.Panicf("failed unmarshal user data body: %v", err)
	}

	if len(userData.Errors) > 0 {
		log.Panic("loader.LoadUntilOk returned API errors")
	}

	userId, err := strconv.ParseUint(userData.Data.User.ID, 10, 64)
	if err != nil {
		log.Panicf("Failed parse user ID: %v", err)
	}

	return UserData{
		ID:              userId,
		Login:           userData.Data.User.Login,
		ProfileImageURL: userData.Data.User.ProfileImageURL,
		CreatedAt:       userData.Data.User.CreatedAt,
		UpdatedAt:       userData.Data.User.UpdatedAt,
		DeletedAt:       userData.Data.User.DeletedAt,
		Description:     userData.Data.User.Description,
		Language:        userData.Data.User.Settings.PreferredLanguageTag,
	}
}
