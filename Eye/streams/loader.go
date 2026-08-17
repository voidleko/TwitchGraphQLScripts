package streams

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/alebik0/TwitchGraphQLScripts/Eye/loader"
)

type OptionsData struct {
	BroadcasterLanguages []string `json:"broadcasterLanguages"`
	FreeformTags         []string `json:"freeformTags"`
	Sort                 string   `json:"sort"`
}

type VtubersPageVariables struct {
	First   uint32      `json:"first"`
	Options OptionsData `json:"options"`
	After   *string     `json:"after"`
}

type VtubersPageRequest struct {
	Query     string               `json:"query"`
	Variables VtubersPageVariables `json:"variables"`
}

type BroadcastSettings struct {
	Title string `json:"title"`
}

type Broadcaster struct {
	ID                string            `json:"id"`
	Login             string            `json:"login"`
	BroadcastSettings BroadcastSettings `json:"broadcastSettings"`
}

type StreamNode struct {
	ID              string      `json:"id"`
	ViewersCount    uint32      `json:"viewersCount"`
	PreviewImageURL string      `json:"previewImageURL"`
	Broadcaster     Broadcaster `json:"broadcaster"`
}

type StreamEdge struct {
	Cursor string     `json:"cursor"`
	Node   StreamNode `json:"node"`
}

type PageInfo struct {
	HasNextPage bool `json:"hasNextPage"`
}

type Streams struct {
	Edges    []StreamEdge `json:"edges"`
	PageInfo PageInfo     `json:"pageInfo"`
}

type Data struct {
	Streams Streams `json:"streams"`
}

type Error struct {
	Message string `json:"message"`
}

type VtubersPageResponse struct {
	Data   Data    `json:"data"`
	Errors []Error `json:"errors"`
}

type StreamData struct {
	ID               string
	ViewersCount     uint32
	PreviewImageURL  string
	BroadcasterID    string
	BroadcasterLogin string
	Title            string
}

const (
	VtubersPageQuery = `
query fetchVtubers($first: Int, $options: StreamOptions, $after: Cursor) {
  streams(first: $first, options: $options, after: $after) {
    edges {
      cursor
      node {
        id
        viewersCount
        previewImageURL(width: 640, height: 360)
        broadcaster {
          id
          login
          broadcastSettings {
            title
          }
        }
      }
    }
    pageInfo {
      hasNextPage
    }
  }
}
`
)

func prepareBody(cursor *string) (io.Reader, error) {
	variables := VtubersPageRequest{
		Query: VtubersPageQuery,
		Variables: VtubersPageVariables{
			First: 30,
			Options: OptionsData{
				BroadcasterLanguages: []string{"RU"},
				FreeformTags:         []string{"Vtuber"},
				Sort:                 "RECENT",
			},
			After: cursor,
		},
	}

	variablesAsString, err := json.Marshal(variables)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal variables: %w", err)
	}

	return strings.NewReader(string(variablesAsString)), nil
}

// Load all vtuber streams
func ForceLoad() []StreamData {
	result := make([]StreamData, 0)

	var cursor *string = nil
	var pageNumber uint = 0
	var hasNextPages bool = true

	for hasNextPages {
		pageNumber++
		log.Printf("Loadinng page #%d", pageNumber)

		body, err := prepareBody(cursor)
		if err != nil {
			log.Panicf("failed prepare load stream page body: %v", err)
		}

		respBody := loader.LoadUntilOk(body)

		var streamsData VtubersPageResponse
		if err := json.Unmarshal([]byte(respBody), &streamsData); err != nil {
			log.Panicf("failed prepare load stream page body: %v", err)
		}

		if len(streamsData.Errors) > 0 {
			log.Panic("loader.LoadUntilOk returned API errors")
		}

		hasNextPages = streamsData.Data.Streams.PageInfo.HasNextPage
		for _, edge := range streamsData.Data.Streams.Edges {
			cursor = &edge.Cursor
			streamData := StreamData{
				ID:               edge.Node.ID,
				ViewersCount:     edge.Node.ViewersCount,
				PreviewImageURL:  edge.Node.PreviewImageURL,
				BroadcasterID:    edge.Node.Broadcaster.ID,
				BroadcasterLogin: edge.Node.Broadcaster.Login,
				Title:            edge.Node.Broadcaster.BroadcastSettings.Title,
			}
			result = append(result, streamData)
		}
	}

	return result
}
