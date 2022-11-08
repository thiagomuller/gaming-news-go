package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Item interface {
	NonPolygonItem | PolygonEntry
}

type NonPolygonItem struct {
	Title           string `xml:"title"`
	Url             string `xml:"link"`
	Description     string `xml:"description"`
	PublicationDate string `xml:"pubDate"`
}

type Channel struct {
	Title       string `xml:"title"`
	Url         string `xml:"link"`
	Description string `xml:"description"`
}

type GameSpotItem struct {
	NonPolygonItem
	Creator      string `xml:"creator"`
	RelatedGames string `xml:"relatedGames"`
}

type GamespotChannel struct {
	Channel
	Items         []GameSpotItem `xml:"item"`
	LastBuildDate string         `xml:"lastBuildDate"`
}

type GamespotRss struct {
	Channel GamespotChannel `xml:"channel"`
}

type IGNChannel struct {
	Channel
	Items []NonPolygonItem `xml:"item"`
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

func fillResultingNews[T any](feedNews []T, allNews []GeneralNews) []GeneralNews {
	switch n := any(feedNews).(type) {
	case []NonPolygonItem:
		for i := 0; i < len(n); i++ {
			if !contains(allNews, n[i].Title) {
				allNews = append(allNews, GeneralNews{
					Title:       n[i].Title,
					Url:         n[i].Url,
					Description: n[i].Description,
				})
			}
		}
	case []PolygonEntry:
		for i := 0; i < len(n); i++ {
			if !contains(allNews, n[i].Title) {
				allNews = append(allNews, GeneralNews{
					Title:       n[i].Title,
					Url:         n[i].Url,
					Description: n[i].Content,
				})
			}
		}
	}
	return allNews
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
	allNews = fillResultingNews(gamespotNews, allNews)
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
	allNews = fillResultingNews(ignNews, allNews)

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

	allNews = fillResultingNews(polygonNews, allNews)

	for i := 0; i < len(allNews); i++ {
		fmt.Println(allNews[i].Title)
	}

	elapsed := time.Since(start)
	fmt.Printf("Took %s\n", elapsed)
}
