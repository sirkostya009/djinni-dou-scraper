package main

import (
	"fmt"
	"github.com/gocolly/colly"
)

func hrefScraper(url, selector, baseUrl string) (scraped []string) {
	c := colly.NewCollector()
	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:122.0) Gecko/20100101 Firefox/122.0"
	c.OnHTML(selector, func(e *colly.HTMLElement) {
		scraped = append(scraped, baseUrl+e.Attr("href"))
	})
	err := c.Visit(url)
	if err != nil {
		fmt.Printf("Failed to scrape %s %v\n", err, url)
	}
	return
}
