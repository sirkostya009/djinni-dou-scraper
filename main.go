package main

import (
	"fmt"
	"github.com/gocolly/colly"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
)

var sleepingTime time.Duration

func htmlUlScraper(elementsPerPage int, url, selector string, callback func(*colly.HTMLElement) string) {
	s := make([]string, elementsPerPage)
	c := colly.NewCollector(colly.UserAgent(
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36",
	), colly.AllowURLRevisit())
	c.OnHTML(selector, func(e *colly.HTMLElement) {
		c := make([]string, elementsPerPage)
		copy(c, s)
		e.ForEach("li", func(i int, e *colly.HTMLElement) {
			s[i] = callback(e)
		})
		for _, v := range s {
			if !slices.Contains(c, v) {
				fmt.Println(v)
			}
		}
	})
	for {
		fmt.Println("\n\nScraping " + url + " ...")
		fmt.Println(time.Now().Format("15:04:05"))
		err := c.Visit(url)
		if err != nil {
			fmt.Printf("Failed to scrape %s %v\n", url, err)
		}
		time.Sleep(sleepingTime * time.Minute)
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please specify the sleeping time in minutes")
		return
	}
	if timing, err := strconv.Atoi(os.Args[1]); err != nil {
		fmt.Println("Please specify as an integer, example: 10")
		return
	} else {
		sleepingTime = time.Duration(timing)
	}
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
	time.Sleep(2 * time.Second) // a little delay so that the output is not mixed
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
