package main

import (
	"fmt"
)

// Guild contains all the guilds that have been signed up
type Guild struct {
	ID      string
	Owner   string
	OwnerID string
}

func (g Guild) String() string {
	return fmt.Sprintf("Guild<%s %s>", g.ID, g.Owner)
}
