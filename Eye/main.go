package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

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

func loadVtuberStreams() []StreamEdge {
	result := []StreamEdge{}

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
			result = append(result, edge)
		}
	}

	return result
}

func main() {
	log.SetPrefix("[LOG] ")

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, "postgres://admin:admin@localhost:5432/eyeofvoidleko?sslmode=disable")
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer conn.Close(ctx)

	for {
		log.Printf("Loading vtubers")

		vtuberStreams := loadVtuberStreams()
		parseData := ParseData{}

		ctx := context.Background()
		log.Printf("INSERT INTO parse_data")
		err := conn.
			QueryRow(
				ctx,
				"INSERT INTO parse_data (time) VALUES ($1) RETURNING id, time;",
				time.Now(),
			).
			Scan(
				&parseData.ID,
				&parseData.Time,
			)
		if err != nil {
			log.Printf("INSERT INTO parse_data error: %v", err)
			break
		}

		for _, vtuberStream := range vtuberStreams {
			streamId, _ := strconv.ParseUint(vtuberStream.Node.ID, 10, 32)
			broadcasterId, _ := strconv.ParseUint(vtuberStream.Node.Broadcaster.ID, 10, 32)
			streamParse := StreamParse{}

			log.Printf("INSERT INTO stream_parse")
			err := conn.
				QueryRow(
					ctx,
					`INSERT INTO 
					stream_parse (parse_id, id, title, preview_image_url, viewers_count, broadcaster_id, broadcaster_login) 
					VALUES ($1, $2, $3, $4, $5, $6, $7)
					RETURNING parse_id, id, title, preview_image_url, viewers_count, broadcaster_id, broadcaster_login;`,
					parseData.ID,
					streamId,
					vtuberStream.Node.Broadcaster.BroadcastSettings.Title,
					vtuberStream.Node.PreviewImageURL,
					vtuberStream.Node.ViewersCount,
					broadcasterId,
					vtuberStream.Node.Broadcaster.Login,
				).
				Scan(
					&streamParse.ParseID,
					&streamParse.ID,
					&streamParse.Title,
					&streamParse.PreviewImageURL,
					&streamParse.ViewersCount,
					&streamParse.BroadcasterID,
					&streamParse.BroadcasterLogin,
				)
			if err != nil {
				log.Printf("INSERT INTO stream_parse error: %v", err)
				break
			}

		}

		log.Printf("Parse %v completed\n", parseData)

		break
		// log.Println("Sleeping")
		// time.Sleep(SleepWait)
	}
}
