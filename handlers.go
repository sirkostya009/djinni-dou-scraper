package main

import (
	"fmt"
	"github.com/gocolly/colly"
	tg "github.com/mymmrac/telego"
	"regexp"
	"slices"
	"strings"
)

type Handlers struct {
	adding   []int64
	removing []int64
}

func (h *Handlers) IsAdding(update tg.Update) bool {
	return slices.Contains(h.adding, update.Message.Chat.ID)
}

func (h *Handlers) IsRemoving(update tg.Update) bool {
	return slices.Contains(h.removing, update.Message.Chat.ID)
}

func (h *Handlers) CancelHandler(bot *tg.Bot, message tg.Message) {
	if i := slices.Index(h.adding, message.Chat.ID); i != -1 {
		h.adding = append(h.adding[:i], h.adding[i+1:]...)
		go bot.SendMessage(&tg.SendMessageParams{
			ChatID: message.Chat.ChatID(),
			Text:   "Addition cancelled",
		})
	}

	if i := slices.Index(h.removing, message.Chat.ID); i != -1 {
		h.removing = append(h.removing[:i], h.removing[i+1:]...)
		go bot.SendMessage(&tg.SendMessageParams{
			ChatID: message.Chat.ChatID(),
			Text:   "Removal cancelled",
		})
	}
}

var (
	djinniRegex      = regexp.MustCompile(`^https://(www\.)?djinni\.co/jobs/\?([a-zA-Z_-]+=[%_.0-9a-zA-Z]+&?)+$`)
	douRegex         = regexp.MustCompile(`^https://(www\.)?jobs\.dou\.ua/vacancies/\?([a-zA-Z_-]+=[%_.0-9a-zA-Z]+&?)+$`)
	nofluffjobsRegex = regexp.MustCompile(`^https://(www\.)?nofluffjobs.com/\w{2}/(\w+)\??([a-zA-Z_-]+=[%_.0-9a-zA-Z]+&?)*$`)
	// TODO: make scraper scrape javascript-rendered pages
	indeedJobsRegex = regexp.MustCompile(`^https://(www\.)?(\w{2}\.)?indeed\.com/jobs\?([a-zA-Z_-]+=[%_.0-9a-zA-Z]+&?)+$`)
)

func (h *Handlers) AddMessage(bot *tg.Bot, message tg.Message) {
	var response string
	chatId := message.Chat.ChatID()
	defer func() {
		bot.SendMessage(&tg.SendMessageParams{
			ChatID: chatId,
			Text:   response,
		})
	}()

	var selector, baseUrl string
	switch {
	case djinniRegex.MatchString(message.Text):
		selector, baseUrl = "a[class*=\" job-list\"]", "https://djinni.co"
	case douRegex.MatchString(message.Text):
		selector = "a.vt"
	case nofluffjobsRegex.MatchString(message.Text):
		selector, baseUrl = "nfj-postings-list[listname=\"search\"] a", "https://nofluffjobs.com"
	default:
		response = "Invalid link, but go ahead, try, try again"
		return
	}

	sub, err := findByUrl(message.Text)
	if slices.Contains(sub.Subscribers, chatId.ID) {
		response = "Already subscribed"
		return
	}
	defer func() {
		i := slices.Index(h.adding, message.Chat.ID)
		h.adding = append(h.adding[:i], h.adding[i+1:]...)
	}()

	sub.Subscribers = append(sub.Subscribers, chatId.ID)
	if err != nil {
		sub.Url = message.Text
		c := colly.NewCollector()
		c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:122.0) Gecko/20100101 Firefox/122.0"
		c.OnHTML(selector, func(e *colly.HTMLElement) {
			sub.Data = append(sub.Data, baseUrl+e.Attr("href"))
		})
		err = c.Visit(sub.Url)
		if err != nil {
			bot.Logger().Errorf("Failed to scrape %s %v\n", err, sub.Url)
		}
		_, err = addSubscription(sub)
	} else {
		_, err = updateSubscription(sub)
	}
	if err != nil {
		bot.Logger().Errorf("%v", err)
		response = "Failed to add subscription"
		return
	}
	response = "Subscription added"
}

func (h *Handlers) RemoveMessage(bot *tg.Bot, message tg.Message) {
	var response string
	chatId := message.Chat.ChatID()
	defer func() {
		bot.SendMessage(&tg.SendMessageParams{
			ChatID: chatId,
			Text:   response,
		})
	}()

	switch {
	case djinniRegex.MatchString(message.Text):
	case douRegex.MatchString(message.Text):
	case nofluffjobsRegex.MatchString(message.Text):
	default:
		response = "Invalid link, but go ahead, try, try again"
		return
	}

	defer func() {
		i := slices.Index(h.removing, message.Chat.ID)
		h.removing = append(h.removing[:i], h.removing[i+1:]...)
	}()

	sub, err := findByUrl(message.Text)
	if err != nil {
		bot.Logger().Errorf("%v", err)
		response = "Subscription not found"
		return
	}

	index := slices.Index(sub.Subscribers, chatId.ID)
	sub.Subscribers = append(sub.Subscribers[:index], sub.Subscribers[index+1:]...)

	if len(sub.Subscribers) == 0 {
		_, err = deleteSubscription(sub.Url)
	} else {
		_, err = updateSubscription(sub)
	}
	if err != nil {
		bot.Logger().Errorf("%v", err)
		response = "Failed to remove subscription"
		return
	}
	response = "Subscription removed"
}

const maxSubscriptions = 3

func (h *Handlers) AddHandler(bot *tg.Bot, message tg.Message) {
	var response string
	if count, err := countSubscriptions(message.Chat.ID); err != nil {
		bot.Logger().Errorf("%v", err)
		response = "Cannot add subscriptions at this time"
	} else if count >= maxSubscriptions {
		response = "Subscription limit reached"
	} else {
		response = "Sure, let's add another subscription, just drop the link here!"
		h.adding = append(h.adding, message.Chat.ID)
	}
	bot.SendMessage(&tg.SendMessageParams{
		ChatID: message.Chat.ChatID(),
		Text:   response,
	})
}

func (h *Handlers) ListIntoString(id int64, str string) (string, bool) {
	if subs := listSubscriptions(id); subs != nil {
		urls := make([]string, len(subs))
		for i, sub := range subs {
			urls[i] = sub.Url
		}
		return fmt.Sprintf("%s\n\n%s", str, strings.Join(urls, "\n\n")), true
	} else {
		return "No active subscriptions", false
	}
}

func (h *Handlers) RemoveHandler(bot *tg.Bot, message tg.Message) {
	response, notEmpty := h.ListIntoString(message.Chat.ID, "Sure, what subscriptions would you like to remove?")
	if notEmpty {
		h.removing = append(h.removing, message.Chat.ID)
	}
	bot.SendMessage(&tg.SendMessageParams{
		ChatID: message.Chat.ChatID(),
		Text:   response,
	})
}

func (h *Handlers) ListHandler(bot *tg.Bot, message tg.Message) {
	response, _ := h.ListIntoString(message.Chat.ID, "List of active subscriptions:")
	bot.SendMessage(&tg.SendMessageParams{
		ChatID: message.Chat.ChatID(),
		Text:   response,
	})
}

func (h *Handlers) StopHandler(bot *tg.Bot, update tg.ChatMemberUpdated) {
	err := deleteSubscriptionsByChatId(update.Chat.ID)
	if err != nil {
		bot.Logger().Errorf("Failed to delete subscriptions: %v", err)
	}
}
