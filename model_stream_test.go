package main

import (
	"testing"
)

func TestStreamType(t *testing.T) {
	items := [][]interface{}{
		[]interface{}{StreamTwitch, "https://www.twitch.tv/"},
	}

	for _, item := range items {
		got := item[0].(StreamType).URL()
		if got != item[1].(string) {
			t.Errorf("URL() = %s; want %s", got, item[1].(string))
		}
	}
}

func TestStream(t *testing.T) {
	items := [][]interface{}{
		[]interface{}{Stream{Type: StreamTwitch, StreamUsername: "halkeye"}, "https://www.twitch.tv/halkeye"},
	}

	for _, item := range items {
		got := item[0].(Stream).URL()
		if got != item[1].(string) {
			t.Errorf("URL() = %s; want %s", got, item[1].(string))
		}
	}
}
