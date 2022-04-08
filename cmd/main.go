package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/go-co-op/gocron"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/mustafabayar/immo-trakt/config"
	"github.com/mustafabayar/immo-trakt/immoscout"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatalln("Unable to load configurations: ", err)
	}

	log.Printf("ImmoTrakt is going to run with following configuration: \nFrequency: %s\nInclude Existing Offers: %v\nSearch: %s\nExclude WBS: %v\nExclude Tausch: %v\nExclude Senior: %v",
		cfg.ImmoTrakt.Frequency,
		cfg.ImmoTrakt.IncludeExistingOffers,
		cfg.ImmobilienScout.Search,
		cfg.ImmobilienScout.ExcludeWBS,
		cfg.ImmobilienScout.ExcludeTausch,
		cfg.ImmobilienScout.ExcludeSenior,
	)

	bot, err := tgbotapi.NewBotAPI(cfg.Telegram.Token)
	if err != nil {
		log.Fatalln("Unable to access telegram bot: ", err)
	}
	log.Printf("Telegram Bot authorized on account %s", bot.Self.UserName)

	var chatID int64
	if len(cfg.Telegram.ChatID) > 0 {
		chatID, _ = strconv.ParseInt(cfg.Telegram.ChatID, 10, 64)
		log.Printf("Telegram chat ID is configured as %v", chatID)
	} else {
		log.Println("Telegram chat ID is not provided via configuration, I will try to retrieve it from Telegram")
		u := tgbotapi.NewUpdate(0)
		updates, err := bot.GetUpdates(u)

		if err != nil {
			log.Panic(err)
		}

		if len(updates) == 0 {
			log.Fatalf("Telegram chat not found, please first send a message to the bot on Telegram and then try to run the ImmoTrakt again!")
		}

		chatID = updates[0].Message.Chat.ID
		log.Printf("Telegram chat ID found as %v", chatID)
	}

	m := make(map[string]immoscout.Listing)
	firstRun := true

	log.Printf("ImmoTrakt scheduled to run with following frequency: %s", cfg.ImmoTrakt.Frequency)
	s := gocron.NewScheduler(time.UTC)
	s.Every(cfg.ImmoTrakt.Frequency).Do(func() {
		var offers = immoscout.GetAllListings(cfg)
		for i := 0; i < len(offers); i++ {
			_, found := m[offers[i].ID]
			if found {
				continue
			}

			listing := offers[i]
			m[offers[i].ID] = listing

			if !firstRun || cfg.ImmoTrakt.IncludeExistingOffers {
				log.Printf("Found new offer %s", listing.Link)
				message := fmt.Sprintf("%s\n%v m²  -  %v rooms  -  %v € warm\n%s", listing.Title, listing.Size, listing.Room, listing.Rent, listing.Link)
				msg := tgbotapi.NewMessage(chatID, message)
				bot.Send(msg)
			}
		}
		firstRun = false
	})
	s.StartBlocking()
}
