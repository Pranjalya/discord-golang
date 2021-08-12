package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/thecodingmachine/gotenberg-go-client/v7"
	"mvdan.cc/xurls/v2"
)

type Config struct {
	DiscordToken   string   `json:"discord_token"`
	ChannelIDs     []string `json:"channel_ids"`
	Administrators []string `json:"administrators"`
	BackendURL     string   `json:"backend_url"`
}

var config Config
var rxRelaxed = xurls.Relaxed()
var c *gotenberg.Client

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func main() {

	direrr := os.Mkdir("output", 0755)
	if direrr != nil {
		log.Println(direrr)
	}

	jsonFile, err := os.Open("config.json")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("config.json loaded")
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal([]byte(byteValue), &config)

	// Get Bot token and start the Bot instance
	token := config.DiscordToken
	if token == "" {
		log.Fatalf("No token available")
		return
	}

	c = &gotenberg.Client{Hostname: config.BackendURL}

	bot, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatal(err)
		return
	}

	bot.AddHandler(messageCreate)

	if err := bot.Open(); err != nil {
		log.Fatal(err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	log.Println("Discord bot is now running. Press CTRL-C to exit.")
	log.Println("Type !activate to add the bot in the channel and !deactivate to remove it")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	bot.Close()
}

func messageCreate(bot *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	if m.Author.ID == bot.State.User.ID {
		return
	}

	// If the message is "ping" reply with "Pong!"
	if m.Content == "!activate" {
		userID := m.Author.Username + "#" + m.Author.Discriminator
		if _, ok := existsIn(userID, config.Administrators); ok {
			if _, found := existsIn(m.ChannelID, config.ChannelIDs); !found {
				config.ChannelIDs = append(config.ChannelIDs, m.ChannelID)
				bot.ChannelMessageSend(m.ChannelID, "Reader bot is now activated in this channel!")
				saveConfig()
			} else {
				bot.ChannelMessageSend(m.ChannelID, "Reader bot is already activated for this channel!")
			}
		} else {
			bot.ChannelMessageSend(m.ChannelID, "You are not an administrator for the bot!")
		}
	}

	if m.Content == "!deactivate" {
		userID := m.Author.Username + "#" + m.Author.Discriminator
		if _, ok := existsIn(userID, config.Administrators); ok {
			if i, found := existsIn(m.ChannelID, config.ChannelIDs); found {
				config.ChannelIDs = removeItem(config.ChannelIDs, i)
				bot.ChannelMessageSend(m.ChannelID, "Reader bot is now deactivated in this channel!")
				saveConfig()
			} else {
				bot.ChannelMessageSend(m.ChannelID, "Reader bot has not been activated for this channel!")
			}
		} else {
			bot.ChannelMessageSend(m.ChannelID, "You are not an administrator for the bot!")
		}
	}

	if _, ok := existsIn(m.ChannelID, config.ChannelIDs); ok {
		link := rxRelaxed.FindString(m.Content)
		if link != "" {
			u, _ := url.Parse(link)
			domain := strings.TrimLeft(u.Host, "www.")
			if u.Scheme == "https" {
				if LinkAvailable(link) {
					if domain == "github.com" {
						log.Println("Accessed GitHub site :", u)
					} else {
						pdfPath := getPDF(link)
						pdf, err := os.Open("output/" + pdfPath)
						if err != nil {
							log.Println(err)
						}
						ref := &discordgo.MessageReference{MessageID: m.ID, ChannelID: m.ChannelID}
						file := &discordgo.File{Name: link + ".pdf", ContentType: "pdf", Reader: pdf}
						data := &discordgo.MessageSend{Content: link, File: file, Reference: ref}
						_, err = bot.ChannelMessageSendComplex(m.ChannelID, data)
						if err != nil {
							log.Println(err)
						}
						os.Remove("output/" + pdfPath)
					}
				} else {
					log.Println("Link", link, "not reachable.")
				}
			}
		}
	}
}

func saveConfig() {
	configJson, _ := json.Marshal(config)
	err := ioutil.WriteFile("config.json", configJson, 0755)
	if err != nil {
		log.Println(err)
	}
}

func existsIn(el string, list []string) (int, bool) {
	for i, val := range list {
		if val == el {
			return i, true
		}
	}
	return -1, false
}

func removeItem(list []string, index int) []string {
	list[len(list)-1], list[index] = "", list[len(list)-1]
	return list[:len(list)-1]
}

func getPDF(link string) string {
	req := gotenberg.NewURLRequest(link)
	req.Margins(gotenberg.NoMargins)
	req.WaitTimeout(30)
	u, _ := url.Parse(link)
	dest := u.Host + "_" + RandStringBytes(5) + ".pdf"
	err := c.Store(req, "output/"+dest)
	if err != nil {
		log.Println(err)
	}
	return dest
}

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func LinkAvailable(s string) bool {
	r, e := http.Head(s)
	return e == nil && r.StatusCode == 200
}
