package main

import (
	"fmt"
	"github.com/gocolly/colly"
	"slices"
	"strings"
	"time"
)

func htmlUlScraper(page int, url, selector string, callback func(*colly.HTMLElement) string) {
	s := make([]string, page)
	c := colly.NewCollector(colly.UserAgent(
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36",
	))
	c.OnHTML(selector, func(e *colly.HTMLElement) {
		c := make([]string, page)
		copy(c, s)
		e.ForEach("li", func(i int, e *colly.HTMLElement) {
			s[i] = callback(e)
		})
		if c[0] != s[0] {
			for _, v := range s {
				if slices.Contains(c, v) {
					break
				}
				fmt.Println(v)
			}
		}
	})
	for {
		fmt.Println()
		fmt.Println()
		fmt.Println()
		fmt.Println("Scraping " + url + " ...")
		err := c.Visit(url)
		if err != nil {
			fmt.Printf("Failed to scrape %s %v\n", url, err)
		}
		time.Sleep(15 * time.Minute)
	}
}

func main() {
	go htmlUlScraper(
		15,
		"https://djinni.co/jobs/?primary_keyword=JavaScript&primary_keyword=Fullstack&primary_keyword=Java&primary_keyword=Golang&exp_level=no_exp&exp_level=1y&exp_level=2y",
		".list-unstyled",
		func(e *colly.HTMLElement) string {
			title := e.ChildText("div header div.mb-1 a")
			title = strings.ReplaceAll(title, "\n", "")
			metadata := make([]string, 0, 4)
			e.ForEach("div header div.font-weight-500 span.nobr", func(i int, e *colly.HTMLElement) {
				metadata = append(metadata, strings.ReplaceAll(e.Text, "\n ", ""))
			})
			url := "https://djinni.co" + e.ChildAttr("div header div.mb-1 div a", "href")
			return fmt.Sprintf("%s%s\n%s", title, strings.Join(metadata, ","), url)
		},
	)
	time.Sleep(2 * time.Second)
	go htmlUlScraper(
		20,
		"https://jobs.dou.ua/vacancies/?category=Java",
		".lt",
		func(e *colly.HTMLElement) string {
			title := e.ChildText("div.title a")
			locations := e.ChildText("div.title span")
			url := e.ChildAttr("div.title a", "href")
			return fmt.Sprintf("%s %s\n%s", title, locations, url)
		},
	)
	select {}
}
