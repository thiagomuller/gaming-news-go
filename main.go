package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Item struct {
	Title           string `xml:"title"`
	Url             string `xml:"link"`
	Description     string `xml:"description"`
	PublicationDate string `xml:"pubDate"`
	Creator         string `xml:"creator"`
	RelatedGames    string `xml:"relatedGames"`
	Content         string `xml:"content"`
	Published       string `xml:"published"`
	Updated         string `xml:"updated"`
}

type Channel struct {
	Title         string `xml:"title"`
	Url           string `xml:"link"`
	Description   string `xml:"description"`
	LastBuildDate string `xml:"lastBuildDate"`
	Items         []Item `xml:"item"`
}

type Rss struct {
	Channel Channel `xml:"channel"`
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

func mountGeneralNews(title string, url string, description string) GeneralNews {
	return GeneralNews{
		Title:       title,
		Url:         url,
		Description: description,
	}
}

func fillResultingNews[T any](feedNews []T, allNews []GeneralNews) []GeneralNews {
	switch n := any(feedNews).(type) {
	case []PolygonEntry:
		for _, news := range n {
			if !contains(allNews, news.Title) {
				allNews = append(allNews, mountGeneralNews(news.Title, news.Url, news.Content))
			}
		}
	case []Item:
		for _, news := range n {
			if !contains(allNews, news.Title) {
				allNews = append(allNews, mountGeneralNews(news.Title, news.Url, news.Description))
			}
		}
	}

	return allNews
}

func makeHttpGetRequest(rssFeed string) []byte {
	response, err := http.Get(rssFeed)
	if err != nil {
		fmt.Printf("%s", err.Error())
	}
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("%s", err.Error())
	}
	return responseBody
}

func main() {
	start := time.Now()
	var allNews []GeneralNews

	gamespotRespBody := makeHttpGetRequest("https://www.gamespot.com/feeds/game-news")
	var rss Rss
	xml.Unmarshal(gamespotRespBody, &rss)
	gamespotNews := rss.Channel.Items
	allNews = fillResultingNews(gamespotNews, allNews)

	ignRespBody := makeHttpGetRequest("http://feeds.feedburner.com/ign/all")
	xml.Unmarshal(ignRespBody, &rss)
	ignNews := rss.Channel.Items
	allNews = fillResultingNews(ignNews, allNews)

	polygonRespBody := makeHttpGetRequest("https://www.polygon.com/rss/index.xml")
	var polygonFeed PolygonFeed
	xml.Unmarshal(polygonRespBody, &polygonFeed)
	polygonNews := polygonFeed.Entries
	allNews = fillResultingNews(polygonNews, allNews)

	for i := 0; i < len(allNews); i++ {
		fmt.Println(allNews[i].Title)
	}

	elapsed := time.Since(start)
	fmt.Println("News quantity: ", len(allNews))
	fmt.Printf("Took %s\n", elapsed)
}
