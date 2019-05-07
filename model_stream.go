package main

import (
	"fmt"
)

// StreamType for twitch/etc
type StreamType int

const (
	// StreamTwitch is enum
	StreamTwitch StreamType = 0
)

func (s StreamType) String() string {
	names := [...]string{
		"Twitch",
	}
	return names[s]
}

// URL will return the url prefix for this stream type
func (s StreamType) URL() string {
	if s == StreamTwitch {
		return "https://www.twitch.tv/"
	}
	panic(fmt.Errorf("Not handling: %d", s))
}

// Stream contains each streamer on each guild
type Stream struct {
	// Guild              *Guild `sql:"composite:guilds"`
	ID                 int64
	GuildID            string `sql:"unique:guild_user"`
	OwnerID            string `sql:"unique:guild_user"`
	OwnerName          string
	OwnerDiscriminator string
	Type               StreamType
	StreamUsername     string
	StreamUserID       string
}

// String returns a stringified version of the object
func (s Stream) String() string {
	return fmt.Sprintf("Stream<%d %s %s %s %s %s>", s.ID, s.GuildID, s.Type, s.StreamUsername, s.OwnerID, s.OwnerName)
}

// Channel returns the channel part of the url
func (s Stream) Channel() string {
	return s.StreamUsername
}

// URL returns the full pretty url
func (s Stream) URL() string {
	return fmt.Sprintf("%s%s", s.Type.URL(), s.StreamUsername)
}
