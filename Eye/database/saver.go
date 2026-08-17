package database

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
)

type Saver struct {
	mu         sync.Mutex
	ctx        context.Context
	connection *pgx.Conn
}

const (
	CreateQuery = `
CREATE TABLE parse_data (
	parse_id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	time     TIMESTAMPTZ NOT NULL
);
CREATE TABLE stream_parse (
	stream_parse_id   INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	parse_id          INTEGER NOT NULL,
	stream_id         BIGINT NOT NULL,
	title             TEXT NOT NULL,
	preview_image_url TEXT,
	viewers_count     INTEGER NOT NULL,
	broadcaster_id    BIGINT NOT NULL,
	broadcaster_login TEXT NOT NULL,

	CONSTRAINT fk_stream_parse_parse_data
		FOREIGN KEY (parse_id)
		REFERENCES parse_data (parse_id)
);
CREATE TABLE viewer_parse (
	viewer_parse_id   INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	stream_parse_id   INTEGER NOT NULL,
	user_id           BIGINT NOT NULL,
	login             TEXT NOT NULL,
	profile_image_url TEXT,
	created_at        TIMESTAMPTZ NOT NULL,
	updated_at        TIMESTAMPTZ,
	deleted_at        TIMESTAMPTZ,
	description       TEXT NOT NULL,
	language          TEXT NOT NULL,

	CONSTRAINT fk_viewer_parse_stream_parse
		FOREIGN KEY (stream_parse_id)
		REFERENCES stream_parse (stream_parse_id)
);
`
	DeleteAllQuery = `
DROP TABLE viewer_parse;
DROP TABLE stream_parse;
DROP TABLE parse_data;
	`
)

type ParseData struct {
	ParseID uint32
	Time    time.Time
}

type StreamParse struct {
	StreamParseID    uint32
	ParseID          uint32
	StreamID         uint64
	Title            string
	PreviewImageURL  string
	ViewersCount     uint32
	BroadcasterID    uint64
	BroadcasterLogin string
}

type ViewerParse struct {
	ViewerParseID   uint32
	StreamParseID   uint32
	UserID          uint64
	Login           string
	ProfileImageURL string
	CreatedAt       time.Time
	UpdatedAt       *time.Time
	DeletedAt       *time.Time
	Description     string
	Language        string
}

func NewSaver(ctx context.Context, user, password, host, port, database string) (*Saver, error) {
	conn, err := pgx.Connect(
		ctx,
		fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=disable",
			user,
			password,
			host,
			port,
			database,
		),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	return &Saver{
		ctx:        ctx,
		connection: conn,
	}, nil
}

func (saver *Saver) Close() {
	saver.connection.Close(saver.ctx)
}

func (saver *Saver) SaveParseData(t time.Time) ParseData {
	saver.mu.Lock()
	defer saver.mu.Unlock()

	parseData := ParseData{}
	err := saver.
		connection.
		QueryRow(
			saver.ctx,
			"INSERT INTO parse_data (time) VALUES ($1) RETURNING parse_id, time;",
			t,
		).
		Scan(
			&parseData.ParseID,
			&parseData.Time,
		)
	if err != nil {
		log.Panicf("INSERT INTO parse_data error: %v", err)
	}
	return parseData
}

func (saver *Saver) SaveStreamParse(
	parseID uint32,
	streamId string,
	title string,
	previewImageURL string,
	viewersCount uint32,
	broadcasterID string,
	broadcasterLogin string,
) StreamParse {
	saver.mu.Lock()
	defer saver.mu.Unlock()

	streamParse := StreamParse{}
	err := saver.
		connection.
		QueryRow(
			saver.ctx,
			`INSERT INTO 
					stream_parse (parse_id, stream_id, title, preview_image_url, viewers_count, broadcaster_id, broadcaster_login) 
					VALUES ($1, $2, $3, $4, $5, $6, $7)
					RETURNING stream_parse_id, parse_id, stream_id, title, preview_image_url, viewers_count, broadcaster_id, broadcaster_login;`,
			parseID,
			streamId,
			title,
			previewImageURL,
			viewersCount,
			broadcasterID,
			broadcasterLogin,
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
	return streamParse
}

func (saver *Saver) SaveViewParse(
	streamParseID uint32,
	userID uint64,
	login string,
	profileImageURL string,
	createdAt time.Time,
	updatedAt *time.Time,
	deletedAt *time.Time,
	description string,
	language string,
) ViewerParse {
	saver.mu.Lock()
	defer saver.mu.Unlock()

	viewerParse := ViewerParse{}
	err := saver.connection.
		QueryRow(
			saver.ctx,
			`INSERT 
				INTO viewer_parse (stream_parse_id, user_id, login, profile_image_url, created_at, updated_at, deleted_at, description, language) 
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
				RETURNING viewer_parse_id, stream_parse_id, user_id, login, profile_image_url, created_at, updated_at, deleted_at, description, language;`,
			streamParseID,
			userID,
			login,
			profileImageURL,
			createdAt,
			updatedAt,
			deletedAt,
			description,
			language,
		).
		Scan(
			&viewerParse.ViewerParseID,
			&viewerParse.StreamParseID,
			&viewerParse.UserID,
			&viewerParse.Login,
			&viewerParse.ProfileImageURL,
			&viewerParse.CreatedAt,
			&viewerParse.UpdatedAt,
			&viewerParse.DeletedAt,
			&viewerParse.Description,
			&viewerParse.Language,
		)
	if err != nil {
		log.Panicf("INSERT INTO viewer_parse error: %v", err)
	}
	return viewerParse
}
