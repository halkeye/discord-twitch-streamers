package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	twitch "github.com/Onestay/go-new-twitch"
	"github.com/bwmarrin/discordgo"
	"github.com/getsentry/raven-go"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

var log = GetLogger()

func init() {
	var err error
	viper.AutomaticEnv()                            // Any time viper.Get is called, check env
	viper.SetEnvPrefix("DISCORD_STREAMERS")         // prefix any env variables with this
	viper.SetConfigType("yaml")                     // configfile is yaml
	viper.SetConfigName(".discord-streamers")       // name of config file (without extension)
	viper.AddConfigPath("/etc/discord-streamers/")  // path to look for the config file in
	viper.AddConfigPath("$HOME/.discord-streamers") // call multiple times to add many search paths
	viper.AddConfigPath(".")                        // optionally look for config in the working directory
	err = viper.ReadInConfig()                      // Find and read the config file
	if err != nil {                                 // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %s", err))
	}

	raven.SetDSN(viper.GetString("sentry.dsn"))
	// raven.SetEnvironment("staging")
	// raven.SetRelease("h3ll0w0rld")
}

var (
	db       *pg.DB
	oauthCfg *oauth2.Config
	store    *sessions.CookieStore
)

const (
	sessionStoreKey = "sess"
)

func main() {
	var err error

	store = sessions.NewCookieStore([]byte(viper.GetString("cookies.secret")))
	oauthCfg = &oauth2.Config{
		ClientID:     viper.GetString("discord.client_id"),
		ClientSecret: viper.GetString("discord.secret_id"),
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://discordapp.com/api/oauth2/authorize",
			TokenURL: "https://discordapp.com/api/oauth2/token",
		},
		RedirectURL: viper.GetString("self_url") + "auth-callback",
		Scopes:      []string{"guilds", "identify"},
	}

	r := mux.NewRouter()
	r.HandleFunc("/", homePageHandler)
	r.HandleFunc("/start", startHandler)
	r.HandleFunc("/auth-callback", authCallbackHandler)
	r.HandleFunc("/destroy-session", sessionDestroyHandler)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))

	http.Handle("/", r)

	log.Info("Listening...")
	go func() {
		err := http.ListenAndServe(":3000", nil)
		if err != nil {
			panic(err)
		}
	}()

	db = pg.Connect(&pg.Options{
		Addr:     viper.GetString("database.addr"),
		User:     viper.GetString("database.user"),
		Password: viper.GetString("database.password"),
		Database: viper.GetString("database.database"),
	})
	defer db.Close()

	err = createSchema(db)
	if err != nil {
		panic(err)
	}

	dg, err := discordgo.New("Bot " + viper.GetString("discord.bot.token"))
	if err != nil {
		log.Info("error creating Discord session,", err)
		return
	}
	// Cleanly close down the Discord session.
	defer dg.Close()
	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)
	dg.AddHandler(guildCreate)
	dg.AddHandler(guildUpdate)
	dg.AddHandler(guildDelete)
	dg.AddHandler(guildMemberAdd)
	dg.AddHandler(guildMemberRemove)
	dg.AddHandler(guildMemberUpdate)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		log.Info("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	log.Notice("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	log.Notice("All done, quitting")

}

func saveGuild(guild *discordgo.Guild) {
	for _, member := range guild.Members {
		if member.User.ID == guild.OwnerID {
			_, err := db.Exec(`INSERT INTO guilds (id, owner, owner_id) values(?, ?, ?)
					ON CONFLICT(id)
					DO UPDATE SET owner_id=?, owner=?`, guild.ID, guild.OwnerID, member.User.Username, guild.OwnerID, member.User.Username)
			if err != nil {
				raven.CaptureErrorAndWait(err, nil)
				log.Error("Error saving guild", err)
			}
			break
		}
	}
}

func guildCreate(s *discordgo.Session, m *discordgo.GuildCreate) {
	saveGuild(m.Guild)
}

func guildUpdate(s *discordgo.Session, m *discordgo.GuildUpdate) {
	saveGuild(m.Guild)
}

func guildDelete(s *discordgo.Session, m *discordgo.GuildDelete) {
	_, err := db.Exec(`DELETE FROM guilds WHERE id=?=`, m.ID)
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		log.Error("Error saving guild", err)
	}
}

func guildMemberAdd(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	j, _ := json.Marshal(m)
	fmt.Println("guildMemberAdd", string(j))
}

func guildMemberRemove(s *discordgo.Session, m *discordgo.GuildMemberRemove) {
	j, _ := json.Marshal(m)
	fmt.Println("guildMemberRemove", string(j))
}

func guildMemberUpdate(s *discordgo.Session, m *discordgo.GuildMemberUpdate) {
	j, _ := json.Marshal(m)
	fmt.Println("guildMemberUpdate", string(j))
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	var err error
	var streamUserID string
	var streamUsername string
	var streamType StreamType

	// {"id":"574301262057832479","channel_id":"110893872388825088","guild_id":"110893872388825088","content":"test test","timestamp":"2019-05-04T18:28:10.876000+00:00","edited_timestamp":"","mention_roles":[],"tts":false,"mention_everyone":false,"author":{"id":"105880217595211776","email":"","username":"halkeye","avatar":"26ed135d310388b8985b0b4af91bf9d5","locale":"","discriminator":"1337","token":"","verified":false,"mfa_enabled":false,"bot":false},"attachments":[],"embeds":[],"mentions":[],"reactions":null,"type":0,"webhook_id":""}

	// messageCreate {"id":"574427767161225216","channel_id":"574047051608883214","content":"this is my private message","timestamp":"2019-05-05T02:50:52.043000+00:00","edited_timestamp":"","mention_roles":[],"tts":false,"mention_everyone":false,"author":{"id":"105880217595211776","email":"","username":"halkeye","avatar":"26ed135d310388b8985b0b4af91bf9d5","locale":"","discriminator":"1337","token":"","verified":false,"mfa_enabled":false,"bot":false},"attachments":[],"embeds":[],"mentions":[],"reactions":null,"type":0,"webhook_id":""}

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	if strings.HasPrefix(strings.ToLower(m.Content), "!addtwitch ") {
		if m.GuildID == "" {
			s.ChannelMessageSend(m.ChannelID, "Private messages are not currently supported")
			return
		}
		streamType, streamUsername, err = streamFromText(strings.TrimSpace(m.Content[len("!addTwitch "):len(m.Content)]))
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error processing text")
			raven.CaptureErrorAndWait(err, nil)
			log.Error("Error processing message: "+m.Content, err)
			return
		}

		if streamType == StreamTwitch {
			twitchClient := twitch.NewClient(viper.GetString("twitch.client_id"))
			twitchUsers, err := twitchClient.GetUsersByLogin(streamUsername)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("User does not exist, or twitch is having errors: %s", err))
				raven.CaptureErrorAndWait(err, nil)
				log.Error("Looking up username: "+m.Content, err)
				return
			}
			streamUserID = twitchUsers[0].ID
		}

		stream := &Stream{
			GuildID:            m.GuildID,
			OwnerID:            m.Author.ID,
			OwnerName:          m.Author.Username,
			OwnerDiscriminator: m.Author.Discriminator,
			Type:               streamType,
			StreamUsername:     streamUsername,
			StreamUserID:       streamUserID,
		}
		j, _ := json.Marshal(stream)
		fmt.Println("stream", string(j))

		_, err = db.Model(stream).OnConflict("(guild_id, owner_id) DO UPDATE").Set("owner_name=EXCLUDED.owner_name, owner_discriminator=EXCLUDED.owner_discriminator, type=EXCLUDED.type, stream_username=EXCLUDED.stream_username, stream_user_id=EXCLUDED.stream_user_id").Insert()
		if err != nil {
			raven.CaptureErrorAndWait(err, nil)
			log.Error("Error saving guild", err)
			return
		}
		log.Notice(m.Author.Username, "Added new twitch", stream.URL())
		s.ChannelMessageSend(m.ChannelID, "Added the URL: "+stream.URL())
		return
	}

	j, _ := json.Marshal(m)
	fmt.Println("messageCreate", string(j))
}

func createSchema(db *pg.DB) error {
	for _, model := range []interface{}{(*Stream)(nil), (*Guild)(nil)} {
		err := db.CreateTable(model, &orm.CreateTableOptions{
			IfNotExists: true,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
