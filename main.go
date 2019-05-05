package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/getsentry/raven-go"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/spf13/viper"
)

var log = GetLogger()

// Guild contains all the guilds that have been signed up
type Guild struct {
	ID      string
	Owner   string
	OwnerID string
}

func (g Guild) String() string {
	return fmt.Sprintf("Guild<%s %s>", g.ID, g.Owner)
}

// Stream contains each streamer on each guild
type Stream struct {
	ID                 string
	GuildID            string
	Guild              *Guild `sql:"composite:guilds"`
	URL                string
	OwnerID            string
	OwnerName          string
	OwnerDiscriminator string
}

func (s Stream) String() string {
	return fmt.Sprintf("Stream<%s %s %s %s %s %s>", s.ID, s.GuildID, s.Guild.Owner, s.URL, s.OwnerID, s.OwnerName)
}

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

var db *pg.DB

func main() {
	var err error

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

	dg, err := discordgo.New("Bot " + viper.GetString("discord.token"))
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

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)

	log.Info("Listening...")
	go func() {
		http.ListenAndServe(":3000", nil)
	}()

	// Wait here until CTRL-C or other term signal is received.
	log.Notice("https://discordapp.com/api/oauth2/authorize?client_id=" + viper.GetString("discord.client_id") + "&scope=bot&permissions=1")
	log.Notice("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

}

func guildCreate(s *discordgo.Session, m *discordgo.GuildCreate) {
	// j, _ := json.Marshal(m)
	// fmt.Println("guildCreate", string(j))
	for _, member := range m.Members {
		if member.User.ID == m.OwnerID {
			_, err := db.Exec(`INSERT INTO guilds (id, owner, owner_id) values(?, ?, ?)
					ON CONFLICT(id)
					DO UPDATE SET owner_id=?, owner=?`, m.ID, m.OwnerID, member.User.Username, m.OwnerID, member.User.Username)
			if err != nil {
				raven.CaptureErrorAndWait(err, nil)
				log.Error("Error saving guild", err)
			}
			break
		}
	}
}

func guildUpdate(s *discordgo.Session, m *discordgo.GuildUpdate) {
	j, _ := json.Marshal(m)
	fmt.Println("guildUpdate", string(j))
}

func guildDelete(s *discordgo.Session, m *discordgo.GuildDelete) {
	j, _ := json.Marshal(m)
	fmt.Println("guildDelete", string(j))
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

	// {"id":"574301262057832479","channel_id":"110893872388825088","guild_id":"110893872388825088","content":"test test","timestamp":"2019-05-04T18:28:10.876000+00:00","edited_timestamp":"","mention_roles":[],"tts":false,"mention_everyone":false,"author":{"id":"105880217595211776","email":"","username":"halkeye","avatar":"26ed135d310388b8985b0b4af91bf9d5","locale":"","discriminator":"1337","token":"","verified":false,"mfa_enabled":false,"bot":false},"attachments":[],"embeds":[],"mentions":[],"reactions":null,"type":0,"webhook_id":""}

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	j, _ := json.Marshal(m)
	fmt.Println("messageCreate", string(j))

	// If the message is "ping" reply with "Pong!"
	if m.Content == "ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong!")
	}

	// If the message is "pong" reply with "Ping!"
	if m.Content == "pong" {
		s.ChannelMessageSend(m.ChannelID, "Ping!")
	}
}

/*
func saveStream() (err error) {
	err = db.Insert(&Stream{
		ID:                 1000,
		GuildID:            1000,
		URL:                "blah",
		OwnerID:            1000,
		OwnerName:          "halkeye",
		OwnerDiscriminator: "1337",
	})
	return
}
*/

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
