package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

type Content struct {
	imageSource        string
	textContent        string
	continueReadingUrl string
}

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

type feedResponse struct {
	feedUrl  string
	response []byte
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

func mountContentForPolygonNews(content Content) string {
	return content.textContent + "\n" + "Image source: " + content.imageSource + "\n" + "Read more: " + content.continueReadingUrl
}

func setImgSrc(tagName []byte, hasAttr bool, tokenizer *html.Tokenizer, content *Content) {
	if string(tagName) == "img" && hasAttr {
		for {
			attrKey, attrValue, moreAttr := tokenizer.TagAttr()
			if string(attrKey) == "src" && attrValue != nil {
				content.imageSource = string(attrValue)
			}
			if !moreAttr {
				break
			}
		}
	}
}

func setContinueReading(tokenizer *html.Tokenizer, content *Content, previousTokenValue *string) {
	for {
		attrKey, attrValue, moreAttr := tokenizer.TagAttr()
		if string(attrKey) == "href" && attrValue != nil {
			content.continueReadingUrl = string(attrValue)
			*previousTokenValue = ""
			break
		}
		if !moreAttr {
			break
		}
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
				var content Content
				previousTokenValue := ""
				reader := strings.NewReader(news.Content)
				tokenizer := html.NewTokenizer(reader)

				for {
					tt := tokenizer.Next()

					if tt == html.ErrorToken {
						break
					}

					tagName, hasAttr := tokenizer.TagName()
					setImgSrc(tagName, hasAttr, tokenizer, &content)

					if string(tagName) == "p" && previousTokenValue != "p" {
						tokenizer.Next()
						content.textContent += tokenizer.Token().Data
						previousTokenValue = "p"
						continue
					}

					if string(tagName) == "p" && previousTokenValue == "p" {
						for {
							tokenizer.Next()
							tagName, hasAttr = tokenizer.TagName()
							if string(tagName) == "a" && hasAttr {
								setContinueReading(tokenizer, &content, &previousTokenValue)
							}
							if content.continueReadingUrl != "" {
								break
							}
						}
					}

				}

				allNews = append(allNews, mountGeneralNews(news.Title, news.Url, mountContentForPolygonNews(content)))
			}
		}
	}
	return allNews
}

func fetchBytesForUnmarshalling(rssFeed string, respChannel chan<- []byte, wg *sync.WaitGroup) {
	response, err := http.Get(rssFeed)
	if err != nil {
		fmt.Printf("%s", err.Error())
	}
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("%s", err.Error())
	}
	respChannel <- responseBody
	wg.Done()
}

func main() {
	start := time.Now()
	var allNews []GeneralNews
	var rss Rss
	feedsToFetch := [3]string{"https://www.gamespot.com/feeds/game-news", "http://feeds.feedburner.com/ign/all", "https://www.polygon.com/rss/index.xml"}
	respChannel := make(chan []byte, 3)

	wg := sync.WaitGroup{}

	for feed := range feedsToFetch {
		wg.Add(1)
		go fetchBytesForUnmarshalling(feedsToFetch[feed], respChannel, &wg)
	}

	go func() {
		wg.Wait()
		close(respChannel)
	}()

	for resp := range respChannel {
		xml.Unmarshal(resp, &rss)
		allNews = fillResultingNews(rss, allNews)
	}

	elapsed := time.Since(start)
	fmt.Println("News quantity: ", len(allNews))
	fmt.Printf("Took %s\n", elapsed)
}
