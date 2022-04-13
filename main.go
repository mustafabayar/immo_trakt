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
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
)

type Response struct {
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
				ResultlistEntry ResultlistEntries `json:"resultlistEntry"`
			} `json:"resultlistEntries"`
		} `json:"resultlist.resultlist"`
	} `json:"searchResponseModel"`
}

type ResultlistEntry struct {
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
}

// Immoscout response sometimes contains this field as array and sometimes as single object. Creating this custom type here to be able write custom JSON marshaller on it
type ResultlistEntries []ResultlistEntry

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
		Frequency             string `default:"1m" yaml:"frequency" envconfig:"IMMOTRAKT_FREQUENCY"`
		IncludeExistingOffers bool   `default:"false" yaml:"include_existing_offers" envconfig:"IMMOTRAKT_INCLUDE_EXISTING"`
	} `yaml:"immo_trakt"`
	Telegram struct {
		Token  string `yaml:"token" envconfig:"IMMOTRAKT_TELEGRAM_TOKEN"`
		ChatID string `yaml:"chat_id" envconfig:"IMMOTRAKT_TELEGRAM_CHAT_ID"`
	} `yaml:"telegram"`
	ImmobilienScout struct {
		Search        string `yaml:"search" envconfig:"IMMOTRAKT_SEARCH"`
		ExcludeWBS    bool   `default:"false" yaml:"exclude_wbs" envconfig:"IMMOTRAKT_EXCLUDE_WBS"`
		ExcludeTausch bool   `default:"false" yaml:"exclude_tausch" envconfig:"IMMOTRAKT_EXCLUDE_TAUSCH"`
		ExcludeSenior bool   `default:"false" yaml:"exclude_senior" envconfig:"IMMOTRAKT_EXCLUDE_SENIOR"`
	} `yaml:"immobilien_scout"`
}

func main() {
	var cfg config
	readFile(&cfg)
	readEnv(&cfg)

	if len(cfg.Telegram.Token) == 0 || len(cfg.ImmobilienScout.Search) == 0 {
		log.Fatalf("Both config.yml and environment variables are not provided. Please provide a config file and try again.")
	}

	log.Printf("ImmoTrakt is going to run with following configuration: \nFrequency: %s\nInclude Existing Offers: %v\nSearch: %s\nExclude WBS: %v\nExclude Tausch: %v",
		cfg.ImmoTrakt.Frequency,
		cfg.ImmoTrakt.IncludeExistingOffers,
		cfg.ImmobilienScout.Search,
		cfg.ImmobilienScout.ExcludeWBS,
		cfg.ImmobilienScout.ExcludeTausch)

	bot, err := tgbotapi.NewBotAPI(cfg.Telegram.Token)
	if err != nil {
		log.Panic(err)
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

	m := make(map[string]offer)
	firstRun := true

	log.Printf("ImmoTrakt scheduled to run with following frequency: %s", cfg.ImmoTrakt.Frequency)
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
			seniorenOffer := strings.Contains(strings.ToLower(title), "senior")

			if (!wbsOffer || !config.ImmobilienScout.ExcludeWBS) && (!tauschOffer || !config.ImmobilienScout.ExcludeTausch) && (!seniorenOffer || !config.ImmobilienScout.ExcludeSenior) {
				offers = append(offers, offer{ID: id, Title: title, Rent: rent, Size: size, Room: room, Link: fmt.Sprintf("https://www.immobilienscout24.de/expose/%s", id)})
			}
		}
	}

	sort.Slice(offers, func(i, j int) bool {
		return offers[i].Rent < offers[j].Rent
	})

	return offers
}

func requestPage(config *config, pageNumber int) Response {
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
	response := Response{}
	err = json.Unmarshal(bodyBytes, &response)
	if err != nil {
		log.Fatalf(err.Error())
	}

	return response
}

func readFile(config *config) {
	f, err := os.Open("config.yml")
	if err != nil {
		log.Println("config.yml is not found, as a backup we will try to load the values from environment variables.")
		return
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(config)
	if err != nil {
		panic(err)
	}
}

func readEnv(config *config) {
	err := envconfig.Process("", config)
	if err != nil {
		panic(err)
	}
}

// Immoscout response sometimes contains resultlistEntry field as array and sometimes as single object. Creating this custom marshaller to parse it correctly depending on the type.
func (r *ResultlistEntries) UnmarshalJSON(b []byte) (err error) {
	single, multi := ResultlistEntry{}, []ResultlistEntry{}
	if err = json.Unmarshal(b, &single); err == nil {
		*r = make([]ResultlistEntry, 1)
		[]ResultlistEntry(*r)[0] = ResultlistEntry(single)
		return
	}
	if err = json.Unmarshal(b, &multi); err == nil {
		*r = ResultlistEntries(multi)
		return
	}
	return
}
