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
			log.Panicf("failed prepare load stream page body: %v", err)
		}

		respBody, err := loadWithRetries(body)
		if err != nil {
			log.Printf("failed load stream page: %v", err)
			log.Printf("Retry page loading")
			continue
		}

		var streamsData VtubersPageResponse
		if err := json.Unmarshal([]byte(respBody), &streamsData); err != nil {
			log.Panicf("failed prepare load stream page body: %v", err)
		}

		if len(streamsData.Errors) > 0 {
			log.Printf("Got API errors")
			for index, err := range streamsData.Errors {
				log.Printf("%d) %s", index, err.Message)
			}
			log.Printf("Retry page loading")
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
		err := conn.
			QueryRow(
				ctx,
				"INSERT INTO parse_data (time) VALUES ($1) RETURNING parse_id, time;",
				time.Now(),
			).
			Scan(
				&parseData.ParseID,
				&parseData.Time,
			)
		if err != nil {
			log.Panicf("INSERT INTO parse_data error: %v", err)
		}
		log.Printf("INSERT INTO parse_data ID = %d", parseData.ParseID)

		for _, vtuberStream := range vtuberStreams {
			streamId, _ := strconv.ParseUint(vtuberStream.Node.ID, 10, 32)
			broadcasterId, _ := strconv.ParseUint(vtuberStream.Node.Broadcaster.ID, 10, 32)
			streamParse := StreamParse{}

			err := conn.
				QueryRow(
					ctx,
					`INSERT INTO 
					stream_parse (parse_id, stream_id, title, preview_image_url, viewers_count, broadcaster_id, broadcaster_login) 
					VALUES ($1, $2, $3, $4, $5, $6, $7)
					RETURNING stream_parse_id, parse_id, stream_id, title, preview_image_url, viewers_count, broadcaster_id, broadcaster_login;`,
					parseData.ParseID,
					streamId,
					vtuberStream.Node.Broadcaster.BroadcastSettings.Title,
					vtuberStream.Node.PreviewImageURL,
					vtuberStream.Node.ViewersCount,
					broadcasterId,
					vtuberStream.Node.Broadcaster.Login,
				).
				Scan(
					&streamParse.StreamParseID,
					&streamParse.ParseID,
					&streamParse.StreamID,
					&streamParse.Title,
					&streamParse.PreviewImageURL,
					&streamParse.ViewersCount,
					&streamParse.BroadcasterID,
					&streamParse.BroadcasterLogin,
				)
			if err != nil {
				log.Panicf("INSERT INTO stream_parse error: %v", err)
			}
			log.Printf("INSERT INTO stream_parse ID = %d", streamParse.StreamParseID)

			go func() {
				pageViewers := loadStreamViewers(streamParse.BroadcasterID)
				log.Printf("Loaded stream viewers for ID = %d", streamParse.StreamID)

				for _, chatter := range pageViewers.Channel.Chatters.Viewers {
					userData := loadUserData(chatter.Login)
					log.Printf("Loaded user data: %v", userData)
					viewerParse := ViewerParse{}

					err := conn.
						QueryRow(
							ctx,
							`INSERT INTO viewer_parse
							(stream_parse_id, user_id, login, profile_image_url, created_at, updated_at, deleted_at, description, language) 
							VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
							RETURNING viewer_parse_id, stream_parse_id, user_id, login, profile_image_url, created_at, updated_at, deleted_at, description, language;`,
							streamParse.StreamParseID,
							userData.ID,
							userData.Login,
							userData.ProfileImageURL,
							userData.CreatedAt,
							userData.UpdatedAt,
							userData.DeletedAt,
							userData.Description,
							userData.Settings.PreferredLanguageTag,
						).
						Scan(
							viewerParse.ViewerParseID,
							viewerParse.StreamParseID,
							viewerParse.UserID,
							viewerParse.Login,
							viewerParse.ProfileImageURL,
							viewerParse.CreatedAt,
							viewerParse.UpdatedAt,
							viewerParse.DeletedAt,
							viewerParse.Description,
							viewerParse.Language,
						)
					if err != nil {
						log.Panicf("INSERT INTO viewer_parse error: %v", err)
					}
					log.Printf("INSERT INTO viewer_parse ID = %d", viewerParse.ViewerParseID)
				}
			}()
		}

		log.Printf("Parse %v completed\n", parseData)

		break
		// log.Println("Sleeping")
		// time.Sleep(SleepWait)
	}
}
