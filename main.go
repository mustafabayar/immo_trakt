package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-co-op/gocron"
	"gopkg.in/gomail.v2"
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
	ID   string
	Rent float32
	Size float32
	Room float32
	Link string
}

func main() {
	mail := gomail.NewMessage()
	mail.SetHeader("From", os.Getenv("fromEmail"))
	mail.SetHeader("To", os.Getenv("toEmail"))
	mail.SetHeader("Subject", "New Flat Found!")
	d := gomail.NewDialer(os.Getenv("emailHost"), 2525, os.Getenv("emailUsername"), os.Getenv("emailPassword"))

	m := make(map[string]offer)
	firstRun := true

	s := gocron.NewScheduler(time.UTC)
	s.Every(5).Minutes().Do(func() {
		var offers = getAllListings()
		for i := 0; i < len(offers); i++ {
			if _, ok := m[offers[i].ID]; ok {
				fmt.Printf("Already exists: %s \n", offers[i].Link)
				continue
			} else {
				m[offers[i].ID] = offers[i]
				fmt.Printf("New listing found: %s \n", offers[i].Link)

				if !firstRun {
					mail.SetBody("text/html", offers[i].Link)
					if err := d.DialAndSend(mail); err != nil {
						panic(err)
					}
				}
			}
		}
		firstRun = false
	})
	s.StartBlocking()
}

func getAllListings() []offer {
	paging := requestPage(1).Paging

	offers := make([]offer, 0, 1000)
	for i := 1; i <= paging.NumberOfPages; i++ {
		var resultList resultList = requestPage(i)
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
			maxWarmRent := float32(800)

			if !wbsOffer && !tauschOffer && rent < maxWarmRent {
				offers = append(offers, offer{ID: id, Rent: rent, Size: size, Room: room, Link: fmt.Sprintf("https://www.immobilienscout24.de/expose/%s", id)})
			}
		}
	}

	sort.Slice(offers, func(i, j int) bool {
		return offers[i].Rent < offers[j].Rent
	})

	return offers
}

func requestPage(pageNumber int) resultList {
	// Let's start with a base url
	baseUrl, err := url.Parse("https://www.immobilienscout24.de")
	if err != nil {
		fmt.Println("Malformed URL: ", err.Error())
		panic(err)
	}

	// Add a Path Segment (Path segment is automatically escaped)
	baseUrl.Path += "Suche/shape/wohnung-mit-balkon-mieten"

	// Prepare Query Parameters
	params := url.Values{}
	params.Add("petsallowedtypes", "yes,negotiable")
	params.Add("numberofrooms", "1.5-")
	params.Add("price", "-800.0")
	params.Add("livingspace", "50.0-")
	params.Add("pricetype", "rentpermonth")
	params.Add("equipment", "builtinkitchen,balcony")
	params.Add("shape", "dV90X0ltZGhwQXhkRX1nQWpLeWBDZkRpfkA-ZXxAaUJfeEFxQXliQnBDbXtCbFhtfUFuZ0BndkN2SnV0QHlXZ2dDdV1lfUBpT31MY2pAZUFnU3JHdXtAaHdDb2NDeH1DZXNAcF9CZ29AalNrWn5NYWpAcUd7ZEBnUXtRcEd2SGpfQXhkQHRsQ2lrQHhyQmlcYHhBUmJ8QGpNcnNAfk1yV3pRakR5RmBAemlDZF9A")
	params.Add("pagenumber", strconv.Itoa(pageNumber))

	// Add Query Parameters to the URL
	baseUrl.RawQuery = params.Encode() // Escape Query Parameters

	fmt.Println(baseUrl.String())
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
