package main

import (
	"database/sql"
	"log"
	"sync"
	"time"
)

const (
	ApiUrl       = "https://gql.twitch.tv/gql"
	ApiClientID  = "kd1unb4b3q4t58fwlpcbzcbnm76a8fp"
	MaxRetries   = 5
	RetryWait    = 3 * time.Second
	SleepWait    = 5 * time.Minute
	MinFollowers = 500
)

func addVtuber(edge Edge) {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	selectSqlStmt := `
    SELECT EXISTS(SELECT 1 FROM users WHERE id = ?);
	`

	var exists bool
	if err := db.QueryRow(selectSqlStmt, edge.Node.Broadcaster.ID).Scan(&exists); err != nil {
		log.Printf("%q: %s\n", err, selectSqlStmt)
		return
	}

	if !exists {
		sqlStmt := `
		INSERT INTO users (id, viewersCount, login, followersCount) 
		VALUES (?, ?, ?, ?);
		`

		_, err = db.Exec(sqlStmt, edge.Node.Broadcaster.ID, edge.Node.ViewersCount, edge.Node.Broadcaster.Login, edge.Node.Broadcaster.Followers.TotalCount)
		if err != nil {
			log.Printf("%q: %s\n", err, sqlStmt)
			return
		}

		log.Printf("Added vtuber %s - %s", edge.Node.Broadcaster.ID, edge.Node.Broadcaster.Login)
	}
}

func addStreamViewers(currentTime time.Time, streamViewers ViewersPageUserData, viewer ViewersPageViewersData) {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	sqlStmt := `
	INSERT INTO viewers (login, streamerId, time) 
	VALUES (?, ?, ?);
	`

	_, err = db.Exec(sqlStmt, viewer.Login, streamViewers.ID, currentTime.Format("2006-01-02 15:04:05"))
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}

	log.Printf("Added viewer for streamer %s - %s", streamViewers.Login, viewer.Login)
}

func main() {
	log.SetPrefix("[LOG] ")

	prepareDatabse()

	for {
		log.Printf("Loadinng vtubers")

		vtuberStreams := loadVtuberStreams()
		var wg sync.WaitGroup
		var dbMu sync.Mutex
		for _, vtuberStream := range vtuberStreams {
			addVtuber(vtuberStream)

			wg.Go(func() {
				currentTime := time.Now()
				streamViewers := loadStreamViewers(vtuberStream.Node.Broadcaster.ID)
				for _, streamViewer := range streamViewers.Channel.Chatters.Viewers {
					dbMu.Lock()
					addStreamViewers(currentTime, streamViewers, streamViewer)
					dbMu.Unlock()
				}
			})
		}
		wg.Wait()
		log.Println("Sleeping")
		time.Sleep(SleepWait)
	}
}
