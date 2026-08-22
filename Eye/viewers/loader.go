package viewers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/alebik0/TwitchGraphQLScripts/Eye/loader"
)

type Error struct {
	Message string `json:"message"`
}

type ViewersPageVariables struct {
	Login *string `json:"login"`
}

type ViewersPageRequest struct {
	Query     string               `json:"query"`
	Variables ViewersPageVariables `json:"variables"`
}

type ViewersPageStreamData struct {
	CreatedAt    string `json:"createdAt"`
	ViewersCount uint32 `json:"viewersCount"`
}

type ViewersPageBroadcastSettingsData struct {
	Title string `json:"title"`
}

type ViewersPageViewersData struct {
	Login string `json:"login"`
}

type ViewersPageChattersData struct {
	Count      uint32                   `json:"count"`
	Moderators []ViewersPageViewersData `json:"moderators"`
	Viewers    []ViewersPageViewersData `json:"viewers"`
	Vips       []ViewersPageViewersData `json:"vips"`
}

type ViewersPageChannelData struct {
	Chatters ViewersPageChattersData `json:"chatters"`
}

type ViewersPageUserData struct {
	Channel ViewersPageChannelData `json:"channel"`
}

type ViewersPageResponseData struct {
	User ViewersPageUserData `json:"user"`
}

type ViewersPageResponse struct {
	Data   ViewersPageResponseData `json:"data"`
	Errors []Error                 `json:"errors"`
}

const (
	ViewersPageQuery = `
query fetchViewers($login: String) {
  user(login: $login, lookupType: ALL) {
	id
	login
    channel {
      chatters {
        count
        moderators {
          login
        }
        viewers {
          login
        }
        vips {
          login
        }
      }
    }
  }
}
`
	RepeatAmount = 15
	Timeout      = 2 * time.Second
)

func prepareBody(login string) (io.Reader, error) {
	variables := ViewersPageRequest{
		Query: ViewersPageQuery,
		Variables: ViewersPageVariables{
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

func loadPage(login string) ViewersPageUserData {
	for range RepeatAmount {
		body, err := prepareBody(login)
		if err != nil {
			log.Panicf("Failed to prepare body: %v", err)
		}

		respBody, err := loader.LoadWithRetries(body)
		if err != nil {
			log.Printf("[%s] Failed to load stream data for user with login, timeout: %v", login, err)
			time.Sleep(Timeout)
			continue
		}

		var viewersPage ViewersPageResponse
		if err := json.Unmarshal([]byte(respBody), &viewersPage); err != nil {
			log.Panicf("Failed unmarshal body: %v", err)
		}

		if len(viewersPage.Errors) > 0 {
			log.Panic("loader.LoadUntilOk returned API errors")
		}

		if viewersPage.Data.User.Channel.Chatters.Count == 0 {
			log.Printf("[%s] Detected chatters API blocking, timeout", login)
			time.Sleep(Timeout)
			continue
		}

		return viewersPage.Data.User
	}

	log.Printf("[%s] Failed to load stream page, return empty", login)

	return ViewersPageUserData{
		Channel: ViewersPageChannelData{
			Chatters: ViewersPageChattersData{
				Count:      0,
				Moderators: make([]ViewersPageViewersData, 0),
				Viewers:    make([]ViewersPageViewersData, 0),
				Vips:       make([]ViewersPageViewersData, 0),
			},
		},
	}
}

func ForceLoad(login string) []string {
	result := make([]string, 0)

	for {
		pageResult := loadPage(login)
		beforeSize := len(result)

		for _, chatter := range pageResult.Channel.Chatters.Moderators {
			if !slices.Contains(result, chatter.Login) {
				result = append(result, chatter.Login)
			}
		}
		for _, chatter := range pageResult.Channel.Chatters.Viewers {
			if !slices.Contains(result, chatter.Login) {
				result = append(result, chatter.Login)
			}
		}
		for _, chatter := range pageResult.Channel.Chatters.Vips {
			if !slices.Contains(result, chatter.Login) {
				result = append(result, chatter.Login)
			}
		}

		afterSize := len(result)
		log.Printf("[%s] Update page: %d -> %d / %d", login, beforeSize, afterSize, pageResult.Channel.Chatters.Count)
		if afterSize == beforeSize {
			log.Printf("[%s] Update page break since no update", login)
			break
		}
	}

	log.Printf("[%s] Completed loading viewers", login)

	return result
}
