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
	id
	login
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
	for i := range RepeatAmount {
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

		if viewersPage.Data.User.Channel.Chatters.Count == 0 {
			log.Printf("[%s] Detected chatters API blocking, timeout", login)
			if i == RepeatAmount-1 {
				return viewersPage.Data.User
			}
			time.Sleep(Timeout)
			continue
		}

		return viewersPage.Data.User
	}

	panic("")
}

func ForceLoad(login string) []string {
	var initialResult ViewersPageUserData

	log.Printf("[%s] Loading initial viewers page", login)
	initialResult = loadPage(login)

	for {
		log.Printf("[%s] Loading update viewers page", login)
		update := loadPage(login)
		before := len(initialResult.Channel.Chatters.Viewers)
		log.Printf("[%s] Update page: before: %d/%d", login, before, initialResult.Channel.Chatters.Count)
		joinedViewers := initialResult.Channel.Chatters.Viewers
		for _, newViewer := range update.Channel.Chatters.Viewers {
			if !slices.Contains(joinedViewers, newViewer) {
				joinedViewers = append(joinedViewers, newViewer)
			}
		}
		initialResult.Channel.Chatters.Viewers = joinedViewers
		after := len(initialResult.Channel.Chatters.Viewers)
		log.Printf("[%s] Update page: after: %d/%d", login, after, initialResult.Channel.Chatters.Count)
		if after == before {
			log.Printf("[%s] Update page break since no update", login)
			break
		}
	}

	result := make([]string, len(initialResult.Channel.Chatters.Viewers))
	for i, chatter := range initialResult.Channel.Chatters.Viewers {
		result[i] = chatter.Login
	}

	log.Printf("[%s] Completed loading viewers for %s", login, login)

	return result
}
