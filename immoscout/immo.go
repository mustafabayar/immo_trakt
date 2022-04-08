package immoscout

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/mustafabayar/immo-trakt/config"
)

type immoscoutResponse struct {
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

// Listing represents struct for an immobilien scout offer
type Listing struct {
	ID    string
	Title string
	Rent  float32
	Size  float32
	Room  float32
	Link  string
}

// GetAllListings returns all listings for the configured url
func GetAllListings(config *config.Config) []Listing {
	numberOfPages := 1
	offers := make([]Listing, 0, 1000)
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

			if strings.Contains(strings.ToLower(title), "wbs") && config.ImmobilienScout.ExcludeWBS {
				continue
			}

			if strings.Contains(strings.ToLower(title), "tausch") && config.ImmobilienScout.ExcludeTausch {
				continue
			}

			if strings.Contains(strings.ToLower(title), "senior") && config.ImmobilienScout.ExcludeSenior {
				continue
			}

			offers = append(offers, Listing{ID: id, Title: title, Rent: rent, Size: size, Room: room, Link: fmt.Sprintf("https://www.immobilienscout24.de/expose/%s", id)})
		}
	}

	sort.Slice(offers, func(i, j int) bool {
		return offers[i].Rent < offers[j].Rent
	})

	return offers
}

func requestPage(config *config.Config, pageNumber int) immoscoutResponse {
	// Let's start with a base url
	baseURL, err := url.Parse(config.ImmobilienScout.Search)
	if err != nil {
		log.Fatalln("Malformed URL: ", err)
	}

	// Handle pagination
	queryParams, _ := url.ParseQuery(baseURL.RawQuery)
	queryParams.Set("pagenumber", strconv.Itoa(pageNumber))
	baseURL.RawQuery = queryParams.Encode()

	log.Printf("Making request to %s", baseURL.String())

	resp, err := http.Post(baseURL.String(), "application/json", nil)
	if err != nil {
		log.Fatalln("Request to ImmoScout failed:  ", err)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	response := immoscoutResponse{}
	err = json.Unmarshal(bodyBytes, &response)
	if err != nil {
		log.Println(string(bodyBytes))
		log.Fatalln("Unable to parse Immoscout response: ", err)
	}

	return response
}
