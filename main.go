package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var s *discordgo.Session

const (
	NumberMax int = 100
)

func Contains(s []int, e int) (int, bool) {
	for i, a := range s {
		if a == e {
			return i, true
		}
	}
	return -1, false
}

func removeAll(baseSlice []int, deleteSlice []int) []int {
	l := make([]int, len(baseSlice))
	copy(l, baseSlice)
	log.Println(deleteSlice)
	remove := func(s []int, i int) []int {
		return append(s[:i], s[i+1:]...)
	}

	for _, v := range deleteSlice {
		if i, ok := Contains(l, v); ok {
			l = remove(l, i)
		}
	}

	return l
}

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

	numChecker := func() []int {
		l := make([]int, 0, NumberMax)
		for i := 1; i < NumberMax+1; i++ {
			l = append(l, i)
		}
		return l
	}()

	if len(u.userNumber) == 0 {
		u.userNumber[username] = i
		log.Println("First: ", u.userNumber[username])
		return
	}

	vals := make([]int, 0)

	for k, v := range u.userNumber {
		if k != username {
			vals = append(vals, v)
		}
	}

	// log.Println(vals)

	if _, ok := Contains(vals, i); ok {
		numChecker = removeAll(numChecker, vals)
		// log.Println("numChecker", numChecker)
		idx := Random(len(numChecker) - 1)
		u.userNumber[username] = numChecker[idx]
		log.Println("equal: ", u.userNumber[username])
	} else {
		u.userNumber[username] = i
		log.Println("not equal: ", u.userNumber[username])
	}
}

var (
	GuildID  string
	BotToken string
	Msg      *discordgo.Message
	msgerr   error

	channelUserMap = make(map[string]*UsersInfo)
	NumberEmojis   = map[int]string{
		1: ":one:",
		2: ":two:",
		3: ":three:",
		4: ":four:",
		5: ":five:",
		6: ":six:",
		7: ":seven:",
		8: ":eight:",
	}
	footer = &discordgo.MessageEmbedFooter{
		Text:    "made by shin pallini.",
		IconURL: "https://pbs.twimg.com/profile_images/1319857864899395584/HvHVUJh3_400x400.jpg",
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		// "command": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// 	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		// 		Type: discordgo.InteractionResponseChannelMessageWithSource,
		// 		Data: &discordgo.InteractionResponseData{
		// 			Flags:   uint64(discordgo.MessageFlagsEphemeral),
		// 			Content: "Content Ephemeral",
		// 		},
		// 	})
		// 	msgSend := discordgo.MessageSend{
		// 		Content: "Message Send Compolex",
		// 		Embeds: []*discordgo.MessageEmbed{
		// 			{
		// 				Title:       "Embed title1",
		// 				Description: "Description1",
		// 			},
		// 		},
		// 	}
		// 	Msg, msgerr = s.ChannelMessageSendComplex(i.ChannelID, &msgSend)
		// 	if msgerr != nil {
		// 		log.Fatal(msgerr)
		// 	}
		// },
		// "edit": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// 	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		// 		Type: discordgo.InteractionResponseChannelMessageWithSource,
		// 		Data: &discordgo.InteractionResponseData{
		// 			Flags:   uint64(discordgo.MessageFlagsEphemeral),
		// 			Content: "Content Edited",
		// 		},
		// 	})

		// 	msgEdit := discordgo.NewMessageEdit(i.ChannelID, Msg.ID)
		// 	msgEdit.SetContent("Edited content!!!!!!!")
		// 	msgEdit.SetEmbed(&discordgo.MessageEmbed{
		// 		Title:       "Edited Title1",
		// 		Description: "Edited Description1",
		// 		Fields: []*discordgo.MessageEmbedField{
		// 			{
		// 				Name:  "Field1",
		// 				Value: "Value1",
		// 			},
		// 		},
		// 	})

		// 	Msg, msgerr = s.ChannelMessageEditComplex(msgEdit)

		// },
		"random": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			num := Random(NumberMax)
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
						discordgo.Button{
							Label: "結果発表！",
							Style: discordgo.SuccessButton,
							Emoji: discordgo.ComponentEmoji{
								Name: "🎯",
							},
							CustomID: "answer_button",
						},
					},
				})
				embeds := []*discordgo.MessageEmbed{
					{
						Title:       "Ito",
						Description: "ボードゲームのItoを遊べるBotです。\nボタンをクリックしてランダムな数字をGetしよう！",
						Color:       0xF7F7F7,
						Timestamp:   GetNow(),
						Thumbnail: &discordgo.MessageEmbedThumbnail{
							URL: "https://m.media-amazon.com/images/I/71lTZzCnvRL._AC_SY355_.jpg",
						},
						Footer: footer,
						Author: &discordgo.MessageEmbedAuthor{
							Name:    s.State.User.Username,
							URL:     "https://twitter.com/shin_0205",
							IconURL: "https://better-default-discord.netlify.app/Icons/Gradient-Green.png",
						},
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
			num := Random(NumberMax)
			member := i.Member.User.Username
			channelUserMap[i.ChannelID].SetUnique(member, num)

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags:   uint64(discordgo.MessageFlagsEphemeral),
					Content: fmt.Sprintf("あなたの数字は「**%d**」です！", channelUserMap[i.ChannelID].userNumber[member]),
				},
			})
		},
		"answer_button": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			embeds := []*discordgo.MessageEmbed{
				{
					Title:     "結果発表！",
					Timestamp: GetNow(),
					Color:     0x29df3c,
					Fields: func() []*discordgo.MessageEmbedField {
						l := make([]*discordgo.MessageEmbedField, 0)
						keys := make([]string, 0, len(channelUserMap[i.ChannelID].userNumber))
						for k, _ := range channelUserMap[i.ChannelID].userNumber {
							keys = append(keys, k)
						}

						sort.SliceStable(keys, func(m, n int) bool {
							return channelUserMap[i.ChannelID].userNumber[keys[m]] > channelUserMap[i.ChannelID].userNumber[keys[n]]
						})

						for count, k := range keys {
							l = append(l, &discordgo.MessageEmbedField{
								Name:   fmt.Sprintf("%s位: %sさん", NumberEmojis[count+1], k),
								Value:  strconv.Itoa(channelUserMap[i.ChannelID].userNumber[k]),
								Inline: false,
							})
						}

						return l
					}(),
					Footer: footer,
					Author: &discordgo.MessageEmbedAuthor{
						Name:    s.State.User.Username,
						URL:     "https://twitter.com/shin_0205",
						IconURL: "https://better-default-discord.netlify.app/Icons/Gradient-Green.png",
					},
				},
			}
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: embeds,
				},
			})
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

	// _, err = s.ApplicationCommandCreate(s.State.User.ID, GuildID, &discordgo.ApplicationCommand{
	// 	Name:        "command",
	// 	Description: "sample command",
	// })
	// _, err = s.ApplicationCommandCreate(s.State.User.ID, GuildID, &discordgo.ApplicationCommand{
	// 	Name:        "edit",
	// 	Description: "Edit message",
	// })

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
