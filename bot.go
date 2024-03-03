package main

import (
	"fmt"
	"github.com/mymmrac/telego"
	"os"
	"regexp"
	"slices"
	"strings"
)

var (
	addCommand    []int64
	removeCommand []int64
)

func isAdding(update telego.Update) bool {
	return slices.Contains(addCommand, update.Message.Chat.ID)
}

func isRemoving(update telego.Update) bool {
	return slices.Contains(removeCommand, update.Message.Chat.ID)
}

func listIntoString(id int64, str string) (string, bool) {
	if subs := listSubscriptions(id); subs == nil {
		return "No active subscriptions", false
	} else {
		urls := make([]string, len(subs))
		for i, sub := range subs {
			urls[i] = sub.Url
		}
		return fmt.Sprintf("%s\n%s", str, strings.Join(urls, "\n\n")), true
	}
}

func createBot(opts ...telego.BotOption) *telego.Bot {
	bot, err := telego.NewBot(os.Getenv("TELEGRAM_BOT_TOKEN"), opts...)
	if err != nil {
		panic(err)
	}

	return bot
}

func cancelHandler(bot *telego.Bot, message telego.Message) {
	var responses []string
	addI := slices.Index(addCommand, message.Chat.ID)
	if addI != -1 {
		addCommand = append(addCommand[:addI], addCommand[addI+1:]...)
		responses = append(responses, "Addition cancelled")
	}
	removeI := slices.Index(removeCommand, message.Chat.ID)
	if removeI != -1 {
		removeCommand = append(removeCommand[:removeI], removeCommand[removeI+1:]...)
		responses = append(responses, "Removal cancelled")
	}
	for _, response := range responses {
		bot.SendMessage(&telego.SendMessageParams{
			ChatID: message.Chat.ChatID(),
			Text:   response,
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

func addMessage(bot *telego.Bot, message telego.Message) {
	var response string
	chatId := message.Chat.ChatID()
	defer func() {
		bot.SendMessage(&telego.SendMessageParams{
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

	defer func() {
		i := slices.Index(addCommand, message.Chat.ID)
		addCommand = append(addCommand[:i], addCommand[i+1:]...)
	}()

	sub, _ := findByUrl(message.Text)
	sub.Subscribers = append(sub.Subscribers, chatId.ID)
	var err error
	if sub.Url == "" {
		sub.Url = message.Text
		sub.Data = hrefScraper(sub.Url, selector, baseUrl)
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

func removeMessage(bot *telego.Bot, message telego.Message) {
	var response string
	chatId := message.Chat.ChatID()
	defer func() {
		bot.SendMessage(&telego.SendMessageParams{
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
		i := slices.Index(removeCommand, message.Chat.ID)
		removeCommand = append(removeCommand[:i], removeCommand[i+1:]...)
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

func addHandler(bot *telego.Bot, message telego.Message) {
	var response string
	if countSubscriptions(message.Chat.ID) >= 3 {
		response = "Subscription limit reached"
	} else {
		response = "Sure, let's add another subscription, just drop the link here!"
		addCommand = append(addCommand, message.Chat.ID)
	}
	bot.SendMessage(&telego.SendMessageParams{
		ChatID: message.Chat.ChatID(),
		Text:   response,
	})
}

func removeHandler(bot *telego.Bot, message telego.Message) {
	response, yuh := listIntoString(message.Chat.ID, "Sure, what subscriptions would you like to remove?")
	if yuh {
		removeCommand = append(removeCommand, message.Chat.ID)
	}
	bot.SendMessage(&telego.SendMessageParams{
		ChatID: message.Chat.ChatID(),
		Text:   response,
	})
}

func listHandler(bot *telego.Bot, message telego.Message) {
	response, _ := listIntoString(message.Chat.ID, "List of active subscriptions:")
	bot.SendMessage(&telego.SendMessageParams{
		ChatID: message.Chat.ChatID(),
		Text:   response,
	})
}

func stopHandler(bot *telego.Bot, update telego.ChatMemberUpdated) {
	err := deleteSubscriptionsByChatId(update.Chat.ID)
	if err != nil {
		bot.Logger().Errorf("Failed to delete subscriptions: %v", err)
	}
}
