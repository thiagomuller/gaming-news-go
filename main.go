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
	Title   string  `xml:"title"`
	Updated string  `xml:"updated"`
	Entries []Item  `xml:"entry"`
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

func fillResultingNews(rssFeed Rss, allNews []GeneralNews) []GeneralNews {
	items := rssFeed.Channel.Items
	entries := rssFeed.Entries
	if len(items) > 0 && len(entries) == 0 {
		for _, news := range items {
			if !contains(allNews, news.Title) {
				allNews = append(allNews, mountGeneralNews(news.Title, news.Url, news.Description))
			}
		}
	} else {
		for _, news := range entries {
			if !contains(allNews, news.Title) {
				allNews = append(allNews, mountGeneralNews(news.Title, news.Url, news.Content))
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
	allNews = fillResultingNews(rss, allNews)

	ignRespBody := makeHttpGetRequest("http://feeds.feedburner.com/ign/all")
	xml.Unmarshal(ignRespBody, &rss)
	allNews = fillResultingNews(rss, allNews)

	polygonRespBody := makeHttpGetRequest("https://www.polygon.com/rss/index.xml")
	xml.Unmarshal(polygonRespBody, &rss)
	allNews = fillResultingNews(rss, allNews)

	for i := 0; i < len(allNews); i++ {
		fmt.Println(allNews[i].Title)
	}

	elapsed := time.Since(start)
	fmt.Println("News quantity: ", len(allNews))
	fmt.Printf("Took %s\n", elapsed)
}
