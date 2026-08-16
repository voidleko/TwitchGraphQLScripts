package main

import "time"

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
