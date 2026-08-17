package main

import (
	"context"
	"log"
	"time"

	database "github.com/alebik0/TwitchGraphQLScripts/Eye/database"
	streams "github.com/alebik0/TwitchGraphQLScripts/Eye/streams"
	users "github.com/alebik0/TwitchGraphQLScripts/Eye/users"
	viewers "github.com/alebik0/TwitchGraphQLScripts/Eye/viewers"
)

const (
	ThreadPoolSize = 64
)

type ThreadPoolInput struct {
	Parse   database.StreamParse
	Chatter string
}

func main() {
	log.SetPrefix("[LOG] ")

	ctx := context.Background()
	saver, err := database.NewSaver(ctx, "admin", "admin", "localhost", "5432", "eyeofvoidleko")
	if err != nil {
		log.Panicf("Failed to create saver: %v", err)
	}
	defer saver.Close()

	// Run thread pool
	ch := make(chan ThreadPoolInput, 1024)
	for range ThreadPoolSize {
		go func(ch <-chan ThreadPoolInput) {
			for input := range ch {
				log.Printf("Got thread pool input")

				userData := users.ForceLoad(input.Chatter)
				saver.SaveViewParse(
					input.Parse.StreamParseID,
					userData.ID,
					userData.Login,
					userData.ProfileImageURL,
					userData.CreatedAt,
					userData.UpdatedAt,
					userData.DeletedAt,
					userData.Description,
					userData.Language,
				)
			}
		}(ch)
	}

	for {
		log.Printf("Loading vtubers")

		vtuberStreams := streams.ForceLoad()
		parseData := saver.SaveParseData(time.Now())
		log.Printf("INSERT INTO parse_data ID = %d", parseData.ParseID)

		for _, vtuberStream := range vtuberStreams {
			go func() {
				streamParse := saver.SaveStreamParse(
					parseData.ParseID,
					vtuberStream.ID,
					vtuberStream.Title,
					vtuberStream.PreviewImageURL,
					vtuberStream.ViewersCount,
					vtuberStream.BroadcasterID,
					vtuberStream.BroadcasterLogin,
				)
				log.Printf("INSERT INTO stream_parse ID = %d", streamParse.StreamParseID)

				viewers := viewers.ForceLoad(streamParse.BroadcasterLogin)
				log.Printf("Loaded stream viewers for %s, len = %d", streamParse.BroadcasterLogin, len(viewers))

				for _, chatter := range viewers {
					ch <- ThreadPoolInput{Parse: streamParse, Chatter: chatter}
				}
			}()
		}

		log.Printf("Parse %v completed\n", parseData)

		break
	}

	time.Sleep(30 * time.Second)
}
