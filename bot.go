package main

import (
	"fmt"
	"github.com/mymmrac/telego"
	"os"
	"regexp"
	"slices"
	"strings"
)

var addCommand []int64

func isAdding(update telego.Update) bool {
	return slices.Contains(addCommand, update.Message.Chat.ID)
}

var removeCommand []int64

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

	_ = bot.SetMyCommands(&telego.SetMyCommandsParams{Commands: []telego.BotCommand{
		{Command: "add", Description: "Add subscription"},
		{Command: "remove", Description: "Remove subscription"},
		{Command: "list", Description: "List subscriptions"},
		{Command: "cancel", Description: "Cancel current command"},
	}})

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

var urlRegex = regexp.MustCompile(`https://(djinni\.co|jobs\.dou\.ua)/(jobs|vacancies)/\?([a-zA-Z_-]+=[_.0-9a-zA-Z]+&?)+`)

func addMessage(bot *telego.Bot, message telego.Message) {
	var response string
	chatId := message.Chat.ChatID()
	defer func() {
		bot.SendMessage(&telego.SendMessageParams{
			ChatID: chatId,
			Text:   response,
		})
	}()
	url := urlRegex.FindString(message.Text)
	if url == "" {
		response = "Invalid link, but go ahead, try, try again"
		return
	}
	sub, _ := findByUrl(url)
	sub.Subscribers = append(sub.Subscribers, chatId.ID)
	if sub.Url == "" {
		sub.Url = url
		go func(sub *Subscription) {
			var s scraper
			var selector string
			switch {
			case strings.Contains(url, "djinni.co"):
				s, selector = djinniCrawler, ".list-unstyled"
			case strings.Contains(url, "jobs.dou.ua"):
				s, selector = douCrawler, ".lt"
			}
			sub.Data = htmlUlScraper(sub.Url, selector, s)

			_, _ = updateSubscription(*sub)
		}(&sub)
	}
	_, err := updateSubscription(sub)
	if err != nil {
		response = err.Error()
		return
	}
	response = "Subscription added"
	i := slices.Index(addCommand, message.Chat.ID)
	addCommand = append(addCommand[:i], addCommand[i+1:]...)
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
	url := urlRegex.FindString(message.Text)
	if url == "" {
		response = "Invalid link, but go ahead, try, try again"
		return
	}
	sub, err := findByUrl(url)
	if err != nil {
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
		response = err.Error()
		return
	}
	response = "Subscription removed"
	i := slices.Index(removeCommand, message.Chat.ID)
	removeCommand = append(removeCommand[:i], removeCommand[i+1:]...)
}

func addHandler(bot *telego.Bot, message telego.Message) {
	var response string
	if subs := listSubscriptions(message.Chat.ID); len(subs) >= 2 {
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

func stopHandler(bot *telego.Bot, update telego.Update) {
	fmt.Println(update)
	err := deleteSubscriptionsByChatId(update.Message.Chat.ID)
	if err != nil {
		bot.Logger().Errorf("Failed to delete subscriptions: %v", err)
	}
}
