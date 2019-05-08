package main

import (
	"encoding/json"
	"fmt"
	"strings"

	twitch "github.com/Onestay/go-new-twitch"
	"github.com/bwmarrin/discordgo"
	"github.com/getsentry/raven-go"
	"github.com/spf13/viper"
)

func saveGuild(guild *discordgo.Guild) {
	for _, member := range guild.Members {
		if member.User.ID == guild.OwnerID {
			guild := &Guild{
				ID:      guild.ID,
				Owner:   member.User.Username,
				OwnerID: guild.OwnerID,
			}
			allGuilds[guild.ID] = guild
			_, err := db.Model(guild).OnConflict("(id) DO UPDATE").Set("owner=EXCLUDED.owner, owner_id=EXCLUDED.owner_id").Insert()
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
	delete(allGuilds, m.Guild.ID)
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
