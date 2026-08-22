package main

import (
	"context"
	"log"
	"os"
	"time"

	database "github.com/alebik0/TwitchGraphQLScripts/Eye/database"
	streams "github.com/alebik0/TwitchGraphQLScripts/Eye/streams"
	users "github.com/alebik0/TwitchGraphQLScripts/Eye/users"
	viewers "github.com/alebik0/TwitchGraphQLScripts/Eye/viewers"
)

const (
	UsersParseThreadPoolSize   = 16
	StreamsParseThreadPoolSize = 8
)

type UsersParseThreadPoolInput struct {
	Parse   database.StreamParse
	Chatter string
}
type StreamsParseThreadPoolInput struct {
	Parse         database.StreamParse
	StreamerLogin string
}

func UserThreadPoolFunc(
	saver *database.Saver,
	usersCh <-chan UsersParseThreadPoolInput,
) {
	for input := range usersCh {
		cachedUserData := saver.ReadUserData(input.Chatter)
		if cachedUserData == nil {
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
		} else {
			saver.SaveViewParse(
				input.Parse.StreamParseID,
				cachedUserData.UserID,
				cachedUserData.Login,
				cachedUserData.ProfileImageURL,
				cachedUserData.CreatedAt,
				cachedUserData.UpdatedAt,
				cachedUserData.DeletedAt,
				cachedUserData.Description,
				cachedUserData.Language,
			)
		}
	}
}

func StreamsThreadPoolFunc(streamsCh <-chan StreamsParseThreadPoolInput, usersCh chan<- UsersParseThreadPoolInput) {
	for input := range streamsCh {
		vs := viewers.ForceLoad(input.StreamerLogin)

		for _, chatter := range vs {
			usersCh <- UsersParseThreadPoolInput{Parse: input.Parse, Chatter: chatter}
		}
	}
}

func main() {
	log.SetPrefix("[LOG] ")

	ctx := context.Background()
	user := os.Getenv("EYE_OF_VOIDLEKO_DATABASE_USER")
	password := os.Getenv("EYE_OF_VOIDLEKO_DATABASE_PASSWORD")
	host := os.Getenv("EYE_OF_VOIDLEKO_DATABASE_HOST")
	port := os.Getenv("EYE_OF_VOIDLEKO_DATABASE_PORT")
	dname := os.Getenv("EYE_OF_VOIDLEKO_DATABASE_DNAME")
	redisHost := os.Getenv("EYE_OF_VOIDLEKO_REDIS_HOST")
	redisPort := os.Getenv("EYE_OF_VOIDLEKO_REDIS_PORT")
	redisPassword := os.Getenv("EYE_OF_VOIDLEKO_REDIS_PASSWORD")
	saver, err := database.NewSaver(ctx, user, password, host, port, dname, redisHost, redisPort, redisPassword)
	if err != nil {
		log.Panicf("Failed to create saver: %v", err)
	}
	defer saver.Close()

	// Run thread pools
	usersCh := make(chan UsersParseThreadPoolInput, UsersParseThreadPoolSize)
	streamsCh := make(chan StreamsParseThreadPoolInput, StreamsParseThreadPoolSize)
	for range UsersParseThreadPoolSize {
		go UserThreadPoolFunc(saver, usersCh)
	}
	for range StreamsParseThreadPoolSize {
		go StreamsThreadPoolFunc(streamsCh, usersCh)
	}

	for {
		vtuberStreams := streams.ForceLoad()
		parseData := saver.SaveParseData(time.Now())

		for _, vtuberStream := range vtuberStreams {
			streamParse := saver.SaveStreamParse(
				parseData.ParseID,
				vtuberStream.ID,
				vtuberStream.Title,
				vtuberStream.PreviewImageURL,
				vtuberStream.ViewersCount,
				vtuberStream.BroadcasterID,
				vtuberStream.BroadcasterLogin,
			)

			streamsCh <- StreamsParseThreadPoolInput{
				Parse:         streamParse,
				StreamerLogin: streamParse.BroadcasterLogin,
			}
		}

		log.Printf("Parse %v completed\n", parseData)

		time.Sleep(10 * time.Minute)
	}
}
