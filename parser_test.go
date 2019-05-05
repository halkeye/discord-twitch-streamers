package main

import (
	"testing"
)

func TestRealUrl(t *testing.T) {
	items := [][]string{
		[]string{"http://www.twitch.tv/kaitlyn", "twitch:kaitlyn"},
		[]string{"https://www.twitch.tv/allyqtea", "twitch:allyqtea"},
		[]string{"https://twitch.tv/threeternity", "twitch:threeternity"},
	}

	for _, item := range items {
		got, err := streamFromText(item[0])
		if err != nil {
			t.Errorf("streamFromText(\"%s\") got an error: %s", item[0], err)
		}
		if got != item[1] {
			t.Errorf("streamFromText(\"%s\") = %s; want %s", item[0], got, item[1])
		}
	}
}
