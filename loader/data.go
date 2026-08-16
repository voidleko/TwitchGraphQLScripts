package main

// Vtuber streams page query
type Followers struct {
	TotalCount uint32 `json:"totalCount"`
}

type Broadcaster struct {
	ID        string    `json:"id"`
	Login     string    `json:"login"`
	Followers Followers `json:"followers"`
}

type Node struct {
	ViewersCount uint32      `json:"viewersCount"`
	Broadcaster  Broadcaster `json:"broadcaster"`
}

type Edge struct {
	Cursor string `json:"cursor"`
	Node   Node   `json:"node"`
}

type PageInfo struct {
	HasNextPage bool `json:"hasNextPage"`
}

type Streams struct {
	Edges    []Edge   `json:"edges"`
	PageInfo PageInfo `json:"pageInfo"`
}

type Data struct {
	Streams Streams `json:"streams"`
}

type Error struct {
	Message string `json:"message"`
}

type OptionsData struct {
	BroadcasterLanguages []string `json:"broadcasterLanguages"`
	FreeformTags         []string `json:"freeformTags"`
	Sort                 string   `json:"sort"`
}

type VtubersPageVariables struct {
	First   uint32      `json:"first"`
	Options OptionsData `json:"options"`
	After   *string     `json:"after"`
}

type VtubersPageRequest struct {
	Query     string               `json:"query"`
	Variables VtubersPageVariables `json:"variables"`
}

type VtubersPageResponse struct {
	Data   Data    `json:"data"`
	Errors []Error `json:"errors"`
}

// User query
type UserVariables struct {
	ID    *string `json:"id"`
	Login *string `json:"login"`
}

type UserRequest struct {
	Query     string        `json:"query"`
	Variables UserVariables `json:"variables"`
}

type UserData struct {
	ID    string `json:"id"`
	Login string `json:"login"`
}

type UserResponseData struct {
	User UserData `json:"user"`
}

type UserResponse struct {
	Data   UserResponseData `json:"data"`
	Errors []Error          `json:"errors"`
}

// Stream viewers page query
type ViewersPageVariables struct {
	ID    *string `json:"id"`
	Login *string `json:"login"`
}

type ViewersPageRequest struct {
	Query     string               `json:"query"`
	Variables ViewersPageVariables `json:"variables"`
}

type ViewersPageStreamData struct {
	CreatedAt    string `json:"createdAt"`
	ViewersCount uint32 `json:"viewersCount"`
}

type ViewersPageBroadcastSettingsData struct {
	Title string `json:"title"`
}

type ViewersPageViewersData struct {
	Login string `json:"login"`
}

type ViewersPageChattersData struct {
	Count   uint32                   `json:"count"`
	Viewers []ViewersPageViewersData `json:"viewers"`
}

type ViewersPageChannelData struct {
	Chatters ViewersPageChattersData `json:"chatters"`
}

type ViewersPageUserData struct {
	ID                string                           `json:"id"`
	Login             string                           `json:"login"`
	Stream            ViewersPageStreamData            `json:"stream"`
	BroadcastSettings ViewersPageBroadcastSettingsData `json:"broadcastSettings"`
	Channel           ViewersPageChannelData           `json:"channel"`
}

type ViewersPageResponseData struct {
	User ViewersPageUserData `json:"user"`
}

type ViewersPageResponse struct {
	Data   ViewersPageResponseData `json:"data"`
	Errors []Error                 `json:"errors"`
}
