package viewers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"slices"
	"strings"

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
	Count   uint32                   `json:"count"`
	Viewers []ViewersPageViewersData `json:"viewers"`
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
    channel {
      chatters {
        count
        viewers {
          login
        }
      }
    }
  }
}
`
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
	body, err := prepareBody(login)
	if err != nil {
		log.Panicf("Failed to prepare body: %v", err)
	}

	respBody := loader.LoadUntilOk(body)

	var viewersPage ViewersPageResponse
	if err := json.Unmarshal([]byte(respBody), &viewersPage); err != nil {
		log.Panicf("Failed unmarshal body: %v", err)
	}

	if len(viewersPage.Errors) > 0 {
		log.Panic("loader.LoadUntilOk returned API errors")
	}

	return viewersPage.Data.User
}

func ForceLoad(login string) []string {
	var initialResult ViewersPageUserData

	log.Printf("Loading initial viewers page")
	initialResult = loadPage(login)

	for {
		log.Printf("Loading update viewers page")
		update := loadPage(login)
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

	result := make([]string, len(initialResult.Channel.Chatters.Viewers))
	for i, chatter := range initialResult.Channel.Chatters.Viewers {
		result[i] = chatter.Login
	}

	return result
}
