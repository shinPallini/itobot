package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var s *discordgo.Session

var (
	GuildID  string
	BotToken string
	Msg      *discordgo.Message
	msgerr   error
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	GuildID = os.Getenv("GUILDID")
	BotToken = os.Getenv("BOTTOKEN")
}

func init() {
	var err error
	s, err = discordgo.New("Bot " + BotToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
}

func main() {
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})

	commandHandlers := map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"command": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags:   uint64(discordgo.MessageFlagsEphemeral),
					Content: "Content Ephemeral",
				},
			})

			Msg, msgerr = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: "Followup message has been created",
			})
			if msgerr != nil {
				log.Fatal(msgerr)
			}
		},
		"edit": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags:   uint64(discordgo.MessageFlagsEphemeral),
					Content: "Content Edited",
				},
			})

			Msg, msgerr = s.FollowupMessageEdit(i.Interaction, Msg.ID, &discordgo.WebhookEdit{
				Content: "Edited followup message!",
			})
		},
	}

	err := s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}
	defer s.Close()

	_, err = s.ApplicationCommandCreate(s.State.User.ID, GuildID, &discordgo.ApplicationCommand{
		Name:        "command",
		Description: "sample command",
	})
	_, err = s.ApplicationCommandCreate(s.State.User.ID, GuildID, &discordgo.ApplicationCommand{
		Name:        "edit",
		Description: "Edit message",
	})

	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
				h(s, i)
			}
		}
	})

	if err != nil {
		log.Fatalf("Cannot create slash command: %v", err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop

	log.Println("Gracefully shutting down.")
}
