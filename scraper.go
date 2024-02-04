package main

import (
	"fmt"
	"github.com/gocolly/colly"
	"github.com/mymmrac/telego"
	"regexp"
	"slices"
	"time"
)

var isScraping bool

func htmlUlScraper(url, selector string, callback crawler) []string {
	var scraped []string
	c := colly.NewCollector(colly.UserAgent(
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36",
	), colly.AllowURLRevisit())
	c.OnHTML(selector, func(e *colly.HTMLElement) {
		e.ForEach("li", func(i int, e *colly.HTMLElement) {
			scraped = append(scraped, callback(e))
		})
	})
	err := c.Visit(url)
	if err != nil {
		fmt.Printf("Failed to scrape %s\n%v\n", url, err)
	}
	return scraped
}

func scrapingJob() {
	djinniRegex := regexp.MustCompile(`https://djinni\.co/jobs/.*`)

	for {
		var finish time.Time
		start := time.Now().UTC()
		isScraping = start.Hour() < 9 || start.Hour() > 18
		if isScraping {
			fmt.Println("Beginning scraping at", start.Format("15:04:05"))

			iter := iterateSubscriptions()
			for iter.Next() {
				sub, _ := iter.Get()
				var cr crawler
				var selector string
				if djinniRegex.MatchString(sub.Url) {
					cr = djinniCrawler
					selector = ".list-unstyled"
				} else {
					cr = douCrawler
					selector = ".lt"
				}

				scraped := htmlUlScraper(sub.Url, selector, cr)

				for _, s := range scraped {
					if !slices.Contains(sub.Data, s) {
						for _, id := range sub.Subscribers {
							_, _ = bot.SendMessage(&telego.SendMessageParams{
								ChatID: telego.ChatID{ID: id},
								Text:   s,
							})
						}
					}
				}

				sub.Data = scraped
				_, err := updateSubscription(sub)
				if err != nil {
					fmt.Printf("Failed to update subscription %s\n%v\n", sub.Url, err)
				}
			}
			_ = iter.Close()

			isScraping = false
			finish = time.Now().UTC()
			fmt.Println("Scraping finished at", finish.Format("15:04:05"))
		} else {
			fmt.Println("Scraping is disabled at", start.Format("15:04:05"))
		}

		waitTime := 3600 - finish.Minute()*60 - finish.Second()
		fmt.Println("Waiting for", waitTime)
		time.Sleep(time.Duration(waitTime) * time.Second)
	}
}
