package main

import "time"

// CREATE TABLE parse_data (
//
//	id    INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
//	time  TIMESTAMPTZ NOT NULL
//
// );
type ParseData struct {
	ID   uint32
	Time time.Time
}

// CREATE TABLE stream_parse (

// 	    parse_id          INTEGER NOT NULL,
// 	    id                BIGINT NOT NULL,
// 	    title             TEXT NOT NULL,
// 	    preview_image_url TEXT,
// 	    viewers_count     INTEGER NOT NULL,
// 	    broadcaster_id    BIGINT NOT NULL,
// 	    broadcaster_login TEXT NOT NULL,
// 		CONSTRAINT fk_stream_parse_parse
// 		    FOREIGN KEY (parse_id)
// 		    REFERENCES parse_data (id)

// );
type StreamParse struct {
	ParseID          uint32
	ID               uint64
	Title            string
	PreviewImageURL  string
	ViewersCount     uint32
	BroadcasterID    uint64
	BroadcasterLogin string
}
