package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

const api = "https://api.waifu.im/search"

type ImageResponse struct {
	URL string `json:"url"`
}

type APIResponse struct {
	Images []struct {
		URL string `json:"url"`
	} `json:"images"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Nem sikerült betölteni a .env fájlt:", err)
	}

	BotToken := os.Getenv("BotToken")
	if BotToken == "" {
		// Ha nincs érték, kérjük be a felhasználótól a konzolon
		fmt.Print("Add meg a BotToken értékét: ")
		fmt.Scanln(&BotToken)
	}

	dg, err := discordgo.New("Bot " + BotToken)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	// Open a websocket connection to Discord and begin listening.
	dg.Open()
	defer dg.Close()

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	// Cleanly close down the Discord session.
	dg.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Content == "!waifu" {
		sendWaifuImages(s, m.ChannelID, 1) // Alapértelmezett érték: 1 kép
	} else if strings.HasPrefix(m.Content, "!waifu ") {
		// Az üzenet tartalmaz paramétert
		numStr := strings.TrimPrefix(m.Content, "!waifu ")
		num, err := strconv.Atoi(numStr)
		if err != nil {
			fmt.Println("Hiba a számolvasás közben:", err)
			return
		}
		sendWaifuImages(s, m.ChannelID, num)
	}
}

// Új függvény: képek elküldése a megadott mennyiségben
func sendWaifuImages(s *discordgo.Session, channelID string, num int) {
	for i := 0; i < num; i++ {
		response, err := http.Get(api)
		if err != nil {
			fmt.Println("Hiba a kép lekérése közben:", err)
			continue // Folytatjuk a következő iterációt
		}
		defer response.Body.Close()

		body, err := io.ReadAll(response.Body)
		if err != nil {
			fmt.Println("Hiba az API válasz olvasása közben:", err)
			continue // Folytatjuk a következő iterációt
		}

		var apiResponse APIResponse
		err = json.Unmarshal(body, &apiResponse)
		if err != nil {
			fmt.Println("Hiba a JSON feldolgozása közben:", err)
			continue // Folytatjuk a következő iterációt
		}

		if len(apiResponse.Images) > 0 {
			fmt.Println("Kép URL-je:", apiResponse.Images[0].URL)
			s.ChannelMessageSend(channelID, apiResponse.Images[0].URL)
		} else {
			fmt.Println("Nem található kép URL.")
		}
	}
}
