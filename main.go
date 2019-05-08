package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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
var (
	db        *pg.DB
	oauthCfg  *oauth2.Config
	store     *sessions.CookieStore
	allGuilds map[string]*Guild
)

const (
	sessionStoreKey = "sess"
)

func init() {
	var err error
	allGuilds = map[string]*Guild{}
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

func main() {
	var err error
	log.Notice("Version: " + Version + ", GitCommit: " + GitCommit + ", GitState: " + GitState + ", BuildDate: " + BuildDate)

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
	r.HandleFunc("/", raven.RecoveryHandler(homePageHandler))
	r.HandleFunc("/start", raven.RecoveryHandler(startHandler))
	r.HandleFunc("/auth-callback", raven.RecoveryHandler(authCallbackHandler))
	r.HandleFunc("/destroy-session", raven.RecoveryHandler(sessionDestroyHandler))
	r.Handle("/healthcheck", healthcheckHandler())
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))

	http.Handle("/", r)

	log.Info("Listening...")
	go func() {
		err := http.ListenAndServe(":3000", nil)
		if err != nil {
			raven.CaptureErrorAndWait(err, nil)
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
		raven.CaptureErrorAndWait(err, nil)
		panic(err)
	}

	dg, err := discordgo.New("Bot " + viper.GetString("discord.bot.token"))
	if err != nil {
		log.Info("error creating Discord session,", err)
		raven.CaptureErrorAndWait(err, nil)
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
		raven.CaptureErrorAndWait(err, nil)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	log.Notice("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	log.Notice("All done, quitting")

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
