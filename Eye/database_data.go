package main

import "time"

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
