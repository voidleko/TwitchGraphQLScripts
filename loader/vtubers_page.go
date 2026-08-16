package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
)

func loadVtubersCount() int {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var count int
	query := `SELECT COUNT(*) FROM users`

	if err := db.QueryRow(query).Scan(&count); err != nil {
		return -1
	}
	return count
}

func prepareVtubersStreamsBody(cursor *string) (io.Reader, error) {
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
	body := strings.NewReader(string(variablesAsString))
	return body, nil
}

func loadVtuberStreams() []Edge {
	result := []Edge{}

	vtubersCount := loadVtubersCount()
	log.Printf("Loaded vtubers: %d", vtubersCount)

	var cursor *string = nil
	var pageNumber uint = 0
	var hasNextPages bool = true
	for hasNextPages {
		pageNumber++
		log.Printf("Loadinng page #%d", pageNumber)

		body, err := prepareVtubersStreamsBody(cursor)
		if err != nil {
			log.Printf("Failed to prepare body: %v", err)
			continue
		}

		respBody, err := loadWithRetries(body)
		if err != nil {
			log.Printf("Failed to load: %v", err)
			hasNextPages = false
			break
		}

		var streamsData VtubersPageResponse
		if err := json.Unmarshal([]byte(respBody), &streamsData); err != nil {
			log.Printf("Failed unmarshal body: err = %v, body = %s", err, respBody)
			continue
		}

		if len(streamsData.Errors) > 0 {
			log.Printf("Got some errors")
			for index, err := range streamsData.Errors {
				log.Printf("%d) %s", index, err.Message)
			}
			continue
		}

		hasNextPages = streamsData.Data.Streams.PageInfo.HasNextPage

		for _, edge := range streamsData.Data.Streams.Edges {
			cursor = &edge.Cursor
			if edge.Node.Broadcaster.Followers.TotalCount > MinFollowers {
				result = append(result, edge)
			}
		}
	}

	return result
}
