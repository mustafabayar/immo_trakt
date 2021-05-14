package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-co-op/gocron"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"gopkg.in/yaml.v3"
)

type immoOffer struct {
	SearchresponseModel struct {
		ResultlistResultlist struct {
			Paging struct {
				Pagenumber       int `json:"pageNumber"`
				Pagesize         int `json:"pageSize"`
				NumberOfPages    int `json:"numberOfPages"`
				NumberOfHits     int `json:"numberOfHits"`
				NumberOfListings int `json:"numberOfListings"`
			} `json:"paging"`
			ResultlistEntries []struct {
				ResultlistEntry []struct {
					ID                   string `json:"@id"`
					Publishdate          string `json:"@publishDate"`
					ResultlistRealEstate struct {
						ID    string `json:"@id"`
						Title string `json:"title"`
						Price struct {
							Value    float32 `json:"value"`
							Currency string  `json:"currency"`
						} `json:"price"`
						LivingSpace         float32 `json:"livingSpace"`
						NumberOfRooms       float32 `json:"numberOfRooms"`
						CalculatedTotalRent struct {
							Totalrent struct {
								Value    float32 `json:"value"`
								Currency string  `json:"currency"`
							} `json:"totalRent"`
						} `json:"calculatedTotalRent"`
					} `json:"resultlist.realEstate"`
				} `json:"resultlistEntry"`
			} `json:"resultlistEntries"`
		} `json:"resultlist.resultlist"`
	} `json:"searchResponseModel"`
}

type offer struct {
	ID    string
	Title string
	Rent  float32
	Size  float32
	Room  float32
	Link  string
}

type config struct {
	ImmoTrakt struct {
		Frequency             string `yaml:"frequency"`
		IncludeExistingOffers bool   `yaml:"include_existing_offers"`
	} `yaml:"immo_trakt"`
	Telegram struct {
		Token string `yaml:"token"`
	} `yaml:"telegram"`
	ImmobilienScout struct {
		Search        string `yaml:"search"`
		ExcludeWBS    bool   `yaml:"exclude_wbs"`
		ExcludeTausch bool   `yaml:"exclude_tausch"`
	} `yaml:"immobilien_scout"`
}

func main() {
	var cfg config
	readFile(&cfg)

	bot, err := tgbotapi.NewBotAPI(cfg.Telegram.Token)
	if err != nil {
		log.Panic(err)
	}
	log.Printf("Telegram Bot authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	updates, err := bot.GetUpdates(u)

	if err != nil {
		log.Panic(err)
	}

	if len(updates) == 0 {
		log.Fatalf("Telegram chat not found, please send a message to the bot first and try to run the ImmoTrakt again!")
	}

	chatID := updates[0].Message.Chat.ID
	log.Printf("Telegram Chat ID found as %v", chatID)

	m := make(map[string]offer)
	firstRun := true

	log.Printf("Program scheduled to run with following frequency: %s", cfg.ImmoTrakt.Frequency)
	s := gocron.NewScheduler(time.UTC)
	s.Every(cfg.ImmoTrakt.Frequency).Do(func() {
		var offers = getAllListings(&cfg)
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

func getAllListings(config *config) []offer {
	numberOfPages := 1
	offers := make([]offer, 0, 1000)
	for i := 1; i <= numberOfPages; i++ {
		immoResponse := requestPage(config, i)
		numberOfPages = immoResponse.SearchresponseModel.ResultlistResultlist.Paging.NumberOfPages
		results := immoResponse.SearchresponseModel.ResultlistResultlist.ResultlistEntries[0].ResultlistEntry
		for i := 0; i < len(results); i++ {
			entry := results[i]
			id := entry.ID
			rent := entry.ResultlistRealEstate.CalculatedTotalRent.Totalrent.Value
			size := entry.ResultlistRealEstate.LivingSpace
			room := entry.ResultlistRealEstate.NumberOfRooms
			title := entry.ResultlistRealEstate.Title

			wbsOffer := strings.Contains(strings.ToLower(title), "wbs")
			tauschOffer := strings.Contains(strings.ToLower(title), "tausch")

			if (!wbsOffer || !config.ImmobilienScout.ExcludeWBS) && (!tauschOffer || !config.ImmobilienScout.ExcludeTausch) {
				offers = append(offers, offer{ID: id, Title: title, Rent: rent, Size: size, Room: room, Link: fmt.Sprintf("https://www.immobilienscout24.de/expose/%s", id)})
			}
		}
	}

	sort.Slice(offers, func(i, j int) bool {
		return offers[i].Rent < offers[j].Rent
	})

	return offers
}

func requestPage(config *config, pageNumber int) immoOffer {
	// Let's start with a base url
	baseURL, err := url.Parse(config.ImmobilienScout.Search)
	if err != nil {
		fmt.Println("Malformed URL: ", err.Error())
		panic(err)
	}

	// Handle pagination
	queryParams, _ := url.ParseQuery(baseURL.RawQuery)
	queryParams.Set("pagenumber", strconv.Itoa(pageNumber))
	baseURL.RawQuery = queryParams.Encode()

	log.Printf("Making request to %s", baseURL.String())

	resp, err := http.Post(baseURL.String(), "application/json", nil)
	if err != nil {
		panic(err)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	response := immoOffer{}
	err = json.Unmarshal(bodyBytes, &response)
	if err != nil {
		log.Fatalf(err.Error())
	}

	return response
}

func readFile(config *config) {
	f, err := os.Open("config.yml")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(config)
	if err != nil {
		panic(err)
	}
}
