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

type response struct {
	SearchResponseModel searchResponseModel `json:"searchResponseModel"`
}

type searchResponseModel struct {
	ResultList resultList `json:"resultlist.resultlist"`
}

type resultList struct {
	Paging            paging              `json:"paging"`
	ResultlistEntries []resultlistEntries `json:"resultlistEntries"`
}

type paging struct {
	PageNumber       int `json:"pageNumber"`
	PageSize         int `json:"pageSize"`
	NumberOfPages    int `json:"numberOfPages"`
	NumberOfHits     int `json:"numberOfHits"`
	NumberOfListings int `json:"numberOfListings"`
}

type resultlistEntries struct {
	ResultlistEntry []resultlistEntry `json:"resultlistEntry"`
}

type resultlistEntry struct {
	ID          string     `json:"@id"`
	PublishDate string     `json:"@publishDate"`
	RealEstate  realEstate `json:"resultlist.realEstate"`
}

type realEstate struct {
	ID            string   `json:"@id"`
	Title         string   `json:"title"`
	ColdRent      coldRent `json:"price"`
	WarmRent      warmRent `json:"calculatedTotalRent"`
	LivingSpace   float32  `json:"livingSpace"`
	NumberOfRooms float32  `json:"numberOfRooms"`
}

type coldRent struct {
	Value    float32 `json:"value"`
	Currency string  `json:"currency"`
}

type warmRent struct {
	Rent coldRent `json:"totalRent"`
}

type offer struct {
	ID    string
	Title string
	Rent  float32
	Size  float32
	Room  float32
	Link  string
}

type Config struct {
	ImmoTrakt struct {
		Frequency             string `yaml:"frequency"`
		IncludeExistingOffers bool   `yaml:"include_existing_offers"`
	} `yaml:"immo_trakt"`
	Telegram struct {
		Token  string `yaml:"token"`
		ChatId int64  `yaml:"chat_id"`
	} `yaml:"telegram"`
	ImmobilienScout struct {
		Search        string `yaml:"search"`
		ExcludeWBS    bool   `yaml:"exclude_wbs"`
		ExcludeTausch bool   `yaml:"exclude_tausch"`
	} `yaml:"immobilien_scout"`
}

func main() {
	var cfg Config
	readFile(&cfg)

	bot, err := tgbotapi.NewBotAPI(cfg.Telegram.Token)
	if err != nil {
		log.Panic(err)
	}
	log.Printf("Telegram Bot authorized on account %s", bot.Self.UserName)

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
				msg := tgbotapi.NewMessage(cfg.Telegram.ChatId, message)
				bot.Send(msg)
			}
		}
		firstRun = false
	})
	s.StartBlocking()
}

func getAllListings(config *Config) []offer {
	numberOfPages := 1
	offers := make([]offer, 0, 1000)
	for i := 1; i <= numberOfPages; i++ {
		var resultList resultList = requestPage(config, i)
		numberOfPages = resultList.Paging.NumberOfPages
		results := resultList.ResultlistEntries[0].ResultlistEntry
		for i := 0; i < len(results); i++ {
			entry := results[i]
			id := entry.ID
			rent := entry.RealEstate.WarmRent.Rent.Value
			size := entry.RealEstate.LivingSpace
			room := entry.RealEstate.NumberOfRooms
			title := entry.RealEstate.Title

			wbsOffer := strings.Contains(strings.ToLower(title), "wbs")
			tauschOffer := strings.Contains(strings.ToLower(title), "tausch")
			maxWarmRent := float32(1000)

			if (!wbsOffer || !config.ImmobilienScout.ExcludeWBS) && (!tauschOffer || !config.ImmobilienScout.ExcludeTausch) && rent < maxWarmRent {
				offers = append(offers, offer{ID: id, Title: title, Rent: rent, Size: size, Room: room, Link: fmt.Sprintf("https://www.immobilienscout24.de/expose/%s", id)})
			}
		}
	}

	sort.Slice(offers, func(i, j int) bool {
		return offers[i].Rent < offers[j].Rent
	})

	return offers
}

func requestPage(config *Config, pageNumber int) resultList {
	// Let's start with a base url
	baseUrl, err := url.Parse(config.ImmobilienScout.Search)
	if err != nil {
		fmt.Println("Malformed URL: ", err.Error())
		panic(err)
	}

	// Handle pagination
	query_params, _ := url.ParseQuery(baseUrl.RawQuery)
	query_params.Set("pagenumber", strconv.Itoa(pageNumber))
	baseUrl.RawQuery = query_params.Encode()

	log.Printf("Making request to %s", baseUrl.String())

	resp, err := http.Post(baseUrl.String(), "application/json", nil)
	if err != nil {
		panic(err)
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)

	response := response{}
	err = json.Unmarshal(bodyBytes, &response)
	if err != nil {
		panic(err)
	}
	return response.SearchResponseModel.ResultList
}

func readFile(config *Config) {
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
