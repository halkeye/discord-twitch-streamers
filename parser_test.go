package main

import (
	"testing"
)

func TestRealUrl(t *testing.T) {
	items := [][]interface{}{
		[]interface{}{"http://www.twitch.tv/kaitlyn", StreamTwitch, "kaitlyn"},
		[]interface{}{"https://www.twitch.tv/allyqtea", StreamTwitch, "allyqtea"},
		[]interface{}{"https://twitch.tv/threeternity", StreamTwitch, "threeternity"},
	}

	for _, item := range items {
		gotType, gotUsername, err := streamFromText(item[0].(string))
		if err != nil {
			t.Errorf("streamFromText(\"%s\") got an error: %s", item[0].(string), err)
		}
		if gotType != item[1] {
			t.Errorf("streamFromText(\"%s\") = %s; want %s", item[0].(string), gotType, item[1].(StreamType))
		}
		if gotUsername != item[2] {
			t.Errorf("streamFromText(\"%s\") = %s; want %s", item[0].(string), gotUsername, item[2].(string))
		}
	}
}
