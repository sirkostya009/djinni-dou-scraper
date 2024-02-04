package main

import (
	"fmt"
	"github.com/gocolly/colly"
	"strings"
)

type crawler func(e *colly.HTMLElement) string

func djinniCrawler(e *colly.HTMLElement) string {
	title := e.ChildText("div header div.mb-1 a")
	title = strings.ReplaceAll(title, "\n", "")
	metadata := make([]string, 0, 4)
	e.ForEach("div header div.font-weight-500 span.nobr", func(i int, e *colly.HTMLElement) {
		metadata = append(metadata, strings.ReplaceAll(e.Text, "\n ", ""))
	})
	url := "https://djinni.co" + e.ChildAttr("div header div.mb-1 div a", "href")
	return fmt.Sprintf("%s%s\n%s", title, strings.Join(metadata, ","), url)
}

func douCrawler(e *colly.HTMLElement) string {
	title := e.ChildText("div.title a")
	locations := e.ChildText("div.title span")
	url := e.ChildAttr("div.title a", "href")
	return fmt.Sprintf("%s %s\n%s", title, locations, url)
}
