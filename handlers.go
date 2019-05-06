package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"html/template"
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/getsentry/raven-go"
	"golang.org/x/oauth2"
)

func homePageHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var streams []Stream

	// FIXME - move out of parsing every time
	t := template.Must(template.ParseFiles("./templates/index.tpl"))

	session, err := store.Get(r, sessionStoreKey)
	if err != nil {
		log.Info("error getting session,", err)
		http.Redirect(w, r, "/start", 302)
		return
	}

	if _, ok := session.Values["accessToken"]; !ok {
		log.Info("No session token,")
		http.Redirect(w, r, "/start", 302)
		return
	}

	if _, ok := session.Values["accessToken"].(string); !ok {
		log.Info("No session token,")
		http.Redirect(w, r, "/start", 302)
		return
	}

	clientDG, err := discordgo.New("Bearer " + session.Values["accessToken"].(string))
	if err != nil {
		log.Error("error creating Discord session,", err)
		http.Redirect(w, r, "/start", 302)
		return
	}
	// Cleanly close down the Discord session.
	defer clientDG.Close()

	guilds, err := clientDG.UserGuilds(100, "", "")
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		fmt.Fprintf(w, "Unable to get guilds")
		log.Error("getting guilds", err)
		return
	}

	err = db.Model(&streams).Select()
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		fmt.Fprintf(w, "Unable to get streams")
		log.Error("getting streams", err)
		return
	}

	data := map[string]interface{}{
		"Streams": streams,
		"Guilds":  guilds,
		"Title":   "there",
	}
	// j, _ := json.Marshal(data)
	// fmt.Println("data", string(j))
	err = t.Execute(w, data)
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		log.Error("error rendering template", err)
		return
	}
}

func startHandler(w http.ResponseWriter, r *http.Request) {
	b := make([]byte, 16)
	rand.Read(b)

	state := base64.URLEncoding.EncodeToString(b)

	session, _ := store.Get(r, sessionStoreKey)
	session.Values["state"] = state
	session.Save(r, w)

	url := oauthCfg.AuthCodeURL(state)
	http.Redirect(w, r, url, 302)
}

func authCallbackHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, sessionStoreKey)
	if err != nil {
		fmt.Fprintln(w, "aborted")
		return
	}

	if r.URL.Query().Get("state") != session.Values["state"] {
		fmt.Fprintln(w, "no state match; possible csrf OR cookies not enabled")
		return
	}

	token, err := oauthCfg.Exchange(oauth2.NoContext, r.URL.Query().Get("code"))
	if err != nil {
		fmt.Fprintln(w, "there was an issue getting your token")
		return
	}

	if !token.Valid() {
		fmt.Fprintln(w, "retreived invalid token")
		return
	}

	clientDG, err := discordgo.New("Bearer " + token.AccessToken)
	if err != nil {
		log.Info("error creating Discord session,", err)
		return
	}
	// Cleanly close down the Discord session.
	defer clientDG.Close()

	user, err := clientDG.User("@me")
	if err != nil {
		log.Error("error getting name", err)
		fmt.Println(w, "error getting name")
		return
	}

	session.Values["userName"] = user.Username
	session.Values["accessToken"] = token.AccessToken
	session.Save(r, w)

	http.Redirect(w, r, "/", 302)
}

func sessionDestroyHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, sessionStoreKey)
	if err != nil {
		fmt.Fprintln(w, "aborted")
		return
	}

	session.Options.MaxAge = -1

	session.Save(r, w)
	http.Redirect(w, r, "/", 302)

}