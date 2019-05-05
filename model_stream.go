package main

import (
	"fmt"
	"strings"
)

// Stream contains each streamer on each guild
type Stream struct {
	// Guild              *Guild `sql:"composite:guilds"`
	ID                 int64
	GuildID            string `sql:"unique:guild_user"`
	URL                string
	OwnerID            string `sql:"unique:guild_user"`
	OwnerName          string
	OwnerDiscriminator string
}

// String returns a stringified version of the object
func (s Stream) String() string {
	return fmt.Sprintf("Stream<%d %s %s %s %s>", s.ID, s.GuildID, s.URL, s.OwnerID, s.OwnerName)
}

// Channel returns the channel part of the url
func (s Stream) Channel() string {
	return strings.Split(s.URL, ":")[1]
}
