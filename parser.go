package main

import (
	"errors"
	"net/url"
	"strings"
)

func streamFromText(input string) (streamType StreamType, streamUsername string, err error) {
	input = strings.TrimSpace(input)

	u, err := url.Parse(input)
	if err != nil {
		return
	}
	if strings.HasSuffix(u.Host, "twitch.tv") {
		if !strings.HasPrefix(u.Path, "/") {
			err = errors.New("Url's path doesn't start with a slash")
			return
		}
		streamType = StreamTwitch
		streamUsername = strings.Split(u.Path, "/")[1]
		return
	}
	err = errors.New("Unable to handle url type: " + input)
	return
}
