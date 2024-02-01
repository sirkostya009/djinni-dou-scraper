package main

import (
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
)

var sleepingTime time.Duration

func serialize(filename string, jobs []string) {
	b, err := json.Marshal(jobs)
	if err != nil {
		fmt.Println(err)
		return
	}
	_ = os.WriteFile(filename, b, 0644)
}

func deserialize(filename string) []string {
	b, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	var jobs []string
	err = json.Unmarshal(b, &jobs)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return jobs
}

func htmlUlScraper(name, url, selector string, callback func(*colly.HTMLElement) string) {
	filename := name + ".json"
	s := deserialize(filename)
	c := colly.NewCollector(colly.UserAgent(
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36",
	), colly.AllowURLRevisit())
	c.OnHTML(selector, func(e *colly.HTMLElement) {
		p := make([]string, 0, len(s))
		e.ForEach("li", func(i int, e *colly.HTMLElement) {
			p = append(p, callback(e))
		})
		for _, v := range p {
			if !slices.Contains(s, v) {
				fmt.Println(v)
			}
		}
		s = p
	})
	for {
		fmt.Println("\n\nScraping " + name + " ...")
		fmt.Println(time.Now().Format("15:04:05"))
		err := c.Visit(url)
		if err != nil {
			fmt.Printf("Failed to scrape %s %v\n", url, err)
		}
		go serialize(filename, s)
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
		"djinni",
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
		"dou",
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
