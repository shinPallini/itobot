package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var s *discordgo.Session

type UsersInfo struct {
	mu         sync.Mutex
	userNumber map[string]int
}

func NewUsersInfo() *UsersInfo {
	return &UsersInfo{
		userNumber: make(map[string]int),
	}
}

func (u *UsersInfo) SetUnique(username string, i int) {
	u.mu.Lock()
	defer u.mu.Unlock()

	if len(u.userNumber) == 0 {
		u.userNumber[username] = i
		return
	}

	for _, v := range u.userNumber {
		if i == v {
			n := Random()
			u.SetUnique(username, n)
		} else {
			u.userNumber[username] = i
		}
	}
}

var (
	GuildID  string
	BotToken string
	Msg      *discordgo.Message
	msgerr   error

	channelUserMap = make(map[string]*UsersInfo)

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"command": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags:   uint64(discordgo.MessageFlagsEphemeral),
					Content: "Content Ephemeral",
				},
			})
			msgSend := discordgo.MessageSend{
				Content: "Message Send Compolex",
				Embeds: []*discordgo.MessageEmbed{
					{
						Title:       "Embed title1",
						Description: "Description1",
					},
				},
			}
			Msg, msgerr = s.ChannelMessageSendComplex(i.ChannelID, &msgSend)
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

			msgEdit := discordgo.NewMessageEdit(i.ChannelID, Msg.ID)
			msgEdit.SetContent("Edited content!!!!!!!")
			msgEdit.SetEmbed(&discordgo.MessageEmbed{
				Title:       "Edited Title1",
				Description: "Edited Description1",
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:  "Field1",
						Value: "Value1",
					},
				},
			})

			Msg, msgerr = s.ChannelMessageEditComplex(msgEdit)

		},
		"random": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			num := Random()
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags:   uint64(discordgo.MessageFlagsEphemeral),
					Content: fmt.Sprintf("Random Number: %d", num),
				},
			})
			channelUserMap[i.ChannelID].SetUnique(s.State.User.Username, num)
		},
		"get": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags:   uint64(discordgo.MessageFlagsEphemeral),
					Content: fmt.Sprintf("Get numberMap: %v", channelUserMap[i.ChannelID].userNumber),
				},
			})
		},
		"ito": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			options := i.ApplicationCommandData().Options
			var (
				content    string
				components []discordgo.MessageComponent
			)
			switch options[0].Name {
			case "start":
				channelUserMap[i.ChannelID] = NewUsersInfo()
				components = append(components, discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Label: "ランダムな数字を抽選!",
							Style: discordgo.PrimaryButton,
							Emoji: discordgo.ComponentEmoji{
								Name: "🎲",
							},
							CustomID: "random_button",
						},
					},
				})
				embeds := []*discordgo.MessageEmbed{
					{
						Title:       "Ito",
						Description: "ボードゲームのItoを遊べるBotです。\nボタンをクリックしてランダムな数字をGetしよう！",
						Color:       0xF7F7F7,
						Timestamp:   "2017-10-31T12:00:00.000Z",
					},
				}
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Embeds:     embeds,
						Components: components,
					},
				})
			case "help":
				content = "Ito help!"
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Flags:   uint64(discordgo.MessageFlagsEphemeral),
						Content: content,
					},
				})
			}
		},
	}

	componentHandler = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"random_button": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			num := Random()
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags:   uint64(discordgo.MessageFlagsEphemeral),
					Content: fmt.Sprintf("Random Number: %d", num),
				},
			})
			member := i.Member.User.Username
			channelUserMap[i.ChannelID].SetUnique(member, num)
		},
	}
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}
	GuildID = os.Getenv("GUILDID")
	BotToken = os.Getenv("BOTTOKEN")
	if GuildID == "" {
		log.Fatal("Cannot connect discord bot. Set environment variable [GuildID].")
	}
	if BotToken == "" {
		log.Fatal("Cannot connect discord bot. Set environment variable [BotToken].")
	}
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

	_, err = s.ApplicationCommandCreate(s.State.User.ID, GuildID, &discordgo.ApplicationCommand{
		Name:        "random",
		Description: "Random number message",
	})
	_, err = s.ApplicationCommandCreate(s.State.User.ID, GuildID, &discordgo.ApplicationCommand{
		Name:        "get",
		Description: "Get users info",
	})

	_, err = s.ApplicationCommandCreate(s.State.User.ID, GuildID, &discordgo.ApplicationCommand{
		Name:        "ito",
		Description: "Itoのゲーム開始やヘルプに関連するコマンドです",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "start",
				Description: "Itoのゲームを開始するコマンドです",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
			},
			{
				Name:        "help",
				Description: "Ito Botの操作方法を確認するコマンドです",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
			},
		},
	})

	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
				h(s, i)
			}
		case discordgo.InteractionMessageComponent:
			if h, ok := componentHandler[i.MessageComponentData().CustomID]; ok {
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
