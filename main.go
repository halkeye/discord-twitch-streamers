package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/spf13/viper"
)

// Guild contains all the guilds that have been signed up
type Guild struct {
	ID    int64
	Owner string
}

func (g Guild) String() string {
	return fmt.Sprintf("Guild<%d %s>", g.ID, g.Owner)
}

// Stream contains each streamer on each guild
type Stream struct {
	ID        int64
	GuildID   int64
	Guild     *Guild
	URL       string
	OwnerID   int64
	OwnerName string
}

func (s Stream) String() string {
	return fmt.Sprintf("Stream<%d %d %s %s %d %s>", s.ID, s.GuildID, s.Guild.Owner, s.URL, s.OwnerID, s.OwnerName)
}

func main() {
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

	db := pg.Connect(&pg.Options{
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
		log.Println("error creating Discord session,", err)
		return
	}
	// Cleanly close down the Discord session.
	defer dg.Close()
	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		log.Println("error opening connection,", err)
		return
	}

	guilds, err := dg.UserGuilds(100, "", "")
	if err != nil {
		log.Println("Error getting my guilds, ", err)
		return
	}

	for _, guild := range guilds {
		println("Listing users of " + guild.Name)
		members, err := dg.GuildMembers(guild.ID, "", 100)
		if err != nil {
			log.Println("Error getting guild members, ", err)
			return
		}
		for _, member := range members {
			// data, err := json.Marshal(member)
			// if err != nil {
			// 	fmt.Println("Error JSON, ", err)
			// 	return
			// }
			// println(string(data))
			if member.User.Username == "halkeye" {
				user, err := dg.User(member.User.ID)
				if err != nil {
					log.Println("Error getting user, ", err)
					return
				}
				data, err := json.Marshal(user)
				if err != nil {
					log.Println("Error JSON, ", err)
					return
				}
				println(string(data))
			}
		}

	}

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)

	log.Println("Listening...")
	go func() {
		http.ListenAndServe(":3000", nil)
	}()

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("https://discordapp.com/api/oauth2/authorize?client_id=" + viper.GetString("discord.client_id") + "&scope=bot&permissions=1")
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// {"id":"574301262057832478","channel_id":"110893872388825088","guild_id":"110893872388825088","content":"test test","timestamp":"2019-05-04T18:28:10.876000+00:00","edited_timestamp":"","mention_roles":[],"tts":false,"mention_everyone":false,"author":{"id":"105880217595211776","email":"","username":"halkeye","avatar":"26ed135d310388b8985b0b4af91bf9d5","locale":"","discriminator":"1337","token":"","verified":false,"mfa_enabled":false,"bot":false},"attachments":[],"embeds":[],"mentions":[],"reactions":null,"type":0,"webhook_id":""}

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	j, _ := json.Marshal(m)
	fmt.Println(string(j))

	// If the message is "ping" reply with "Pong!"
	if m.Content == "ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong!")
	}

	// If the message is "pong" reply with "Ping!"
	if m.Content == "pong" {
		s.ChannelMessageSend(m.ChannelID, "Ping!")
	}
}

func createSchema(db *pg.DB) error {
	for _, model := range []interface{}{(*Guild)(nil), (*Stream)(nil)} {
		err := db.CreateTable(model, &orm.CreateTableOptions{
			Temp: true,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
