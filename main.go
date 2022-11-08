package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"
)

type GameSpotItem struct {
	Title           string `xml:"title"`
	Url             string `xml:"link"`
	Description     string `xml:"description"`
	PublicationDate string `xml:"pubDate"`
	Creator         string `xml:"creator"`
	RelatedGames    string `xml:"relatedGames"`
}

type GamespotChannel struct {
	Title         string         `xml:"title"`
	Url           string         `xml:"link"`
	Description   string         `xml:"description"`
	LastBuildDate string         `xml:"lastBuildDate"`
	Items         []GameSpotItem `xml:"item"`
}

type GamespotRss struct {
	Channel GamespotChannel `xml:"channel"`
}

type IGNItem struct {
	Title           string `xml:"title"`
	Url             string `xml:"link"`
	Description     string `xml:"description"`
	PublicationDate string `xml:"pubDate"`
}

type IGNChannel struct {
	Title       string    `xml:"title"`
	Url         string    `xml:"link"`
	Description string    `xml:"description"`
	Items       []IGNItem `xml:"item"`
}

type IGNRss struct {
	Channel IGNChannel `xml:"channel"`
}

type PolygonEntry struct {
	Published string `xml:"published"`
	Updated   string `xml:"updated"`
	Title     string `xml:"title"`
	Content   string `xml:"content"`
	Url       string `xml:"link"`
}

type PolygonFeed struct {
	Title   string         `xml:"title"`
	Updated string         `xml:"updated"`
	Entries []PolygonEntry `xml:"entry"`
}

type GeneralNews struct {
	Title       string
	Url         string
	Description string
}

func contains(allNews []GeneralNews, value string) bool {
	for _, news := range allNews {
		if news.Title == value {
			return true
		}
	}
	return false
}

func main() {
	start := time.Now()
	var allNews []GeneralNews
	gamespotResp, err := http.Get("https://www.gamespot.com/feeds/game-news")
	if err != nil {
		fmt.Printf("%s", err.Error())
	}
	gamespotBody, err := io.ReadAll(gamespotResp.Body)
	if err != nil {
		fmt.Printf("%s", err.Error())
	}
	var gamespotRss GamespotRss
	xml.Unmarshal(gamespotBody, &gamespotRss)
	gamespotNews := gamespotRss.Channel.Items
	for i := 0; i < len(gamespotNews); i++ {
		if !contains(allNews, gamespotNews[i].Title) {
			allNews = append(allNews, GeneralNews{
				Title:       gamespotNews[i].Title,
				Url:         gamespotNews[i].Url,
				Description: gamespotNews[i].Description,
			})
		}
	}
	ignResp, err := http.Get("http://feeds.feedburner.com/ign/all")
	if err != nil {
		fmt.Printf("%s", err.Error())
	}
	ignBody, err := io.ReadAll(ignResp.Body)
	if err != nil {
		fmt.Printf("%s", err.Error())
	}
	var ignRss IGNRss
	xml.Unmarshal(ignBody, &ignRss)
	ignNews := ignRss.Channel.Items
	for i := 0; i < len(ignNews); i++ {
		if !contains(allNews, ignNews[i].Title) {
			allNews = append(allNews, GeneralNews{
				Title:       ignNews[i].Title,
				Url:         ignNews[i].Url,
				Description: ignNews[i].Description,
			})
		}
	}
	for i := 0; i < len(ignNews); i++ {
		if !contains(allNews, ignNews[i].Title) {
			allNews = append(allNews, GeneralNews{
				Title:       ignNews[i].Title,
				Url:         ignNews[i].Url,
				Description: ignNews[i].Description,
			})
		}
	}

	polygonResp, err := http.Get("https://www.polygon.com/rss/index.xml")
	if err != nil {
		fmt.Printf("%s", err.Error())
	}
	polygonBody, err := io.ReadAll(polygonResp.Body)
	if err != nil {
		fmt.Printf("%s", err.Error())
	}
	var polygonFeed PolygonFeed
	xml.Unmarshal(polygonBody, &polygonFeed)
	polygonNews := polygonFeed.Entries

	for i := 0; i < len(polygonNews); i++ {
		if !contains(allNews, polygonNews[i].Title) {
			allNews = append(allNews, GeneralNews{
				Title:       polygonNews[i].Title,
				Url:         polygonNews[i].Url,
				Description: polygonNews[i].Content,
			})
		}
	}

	for i := 0; i < len(allNews); i++ {
		fmt.Println(allNews[i].Title)
	}

	elapsed := time.Since(start)
	fmt.Printf("Took %s\n", elapsed)
}
