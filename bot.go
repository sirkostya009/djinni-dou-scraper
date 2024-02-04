package main

import (
	"fmt"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	"os"
	"regexp"
	"slices"
	"strings"
)

var commands = []telego.BotCommand{
	{Command: "add", Description: "Add subscription"},
	{Command: "remove", Description: "Remove subscription"},
	{Command: "list", Description: "List subscriptions"},
	{Command: "cancel", Description: "Cancel current command"},
}

var addCommand []int64

func isAdding(update telego.Update) bool {
	return slices.Contains(addCommand, update.Message.Chat.ID)
}

var removeCommand []int64

func isRemoving(update telego.Update) bool {
	return slices.Contains(removeCommand, update.Message.Chat.ID)
}

var urlRegex = regexp.MustCompile(`https://(djinni\.co|jobs\.dou\.ua)/(jobs|vacancies)/\?([a-zA-Z_-]+=[.0-9a-zA-Z]+&?)+`)

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

var bot *telego.Bot

func startBot() {
	var err error
	bot, err = telego.NewBot(os.Getenv("TELEGRAM_BOT_TOKEN"))
	if err != nil {
		panic(err)
	}
	go scrapingJob()

	_ = bot.SetMyCommands(&telego.SetMyCommandsParams{Commands: commands})

	updates, err := bot.UpdatesViaLongPolling(nil)
	if err != nil {
		panic(err)
	}
	defer bot.StopLongPolling()

	bh, err := th.NewBotHandler(bot, updates)

	bh.HandleMessage(func(bot *telego.Bot, message telego.Message) {
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
			_, _ = bot.SendMessage(&telego.SendMessageParams{
				ChatID: message.Chat.ChatID(),
				Text:   response,
			})
		}
	}, th.CommandEqual("cancel"))

	bh.HandleMessage(func(bot *telego.Bot, message telego.Message) {
		var response string
		chatId := message.Chat.ChatID()
		defer func() {
			bot.SendMessage(&telego.SendMessageParams{
				ChatID: chatId,
				Text:   response,
			})
		}()
		if isScraping {
			response = "Cannot add subscriptions at this time. Try again later."
			return
		}
		url := urlRegex.FindString(message.Text)
		if url == "" {
			response = "Invalid link, but go ahead, try, try again"
			return
		}
		sub, _ := findByUrl(url)
		if sub.Url == "" {
			sub.Url = url
		}
		sub.Subscribers = append(sub.Subscribers, chatId.ID)
		_, err = updateSubscription(*sub)
		if err != nil {
			response = err.Error()
			return
		}
		response = "Subscription added"
		i := slices.Index(addCommand, message.Chat.ID)
		addCommand = append(addCommand[:i], addCommand[i+1:]...)
	}, isAdding)

	bh.HandleMessage(func(bot *telego.Bot, message telego.Message) {
		var response string
		chatId := message.Chat.ChatID()
		defer func() {
			bot.SendMessage(&telego.SendMessageParams{
				ChatID: chatId,
				Text:   response,
			})
		}()
		if isScraping {
			response = "Cannot remove subscriptions at this time. Try again later."
			return
		}
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
		_, err = updateSubscription(*sub)
		if err != nil {
			response = err.Error()
			return
		}
		response = "Subscription removed"
		i := slices.Index(removeCommand, message.Chat.ID)
		removeCommand = append(removeCommand[:i], removeCommand[i+1:]...)
	}, isRemoving)

	bh.HandleMessage(func(bot *telego.Bot, message telego.Message) {
		var response string
		if subs := listSubscriptions(message.Chat.ID); len(subs) >= 2 {
			response = "Subscription limit reached"
		} else {
			response = "Sure, let's add another subscription, just drop the link here!"
			addCommand = append(addCommand, message.Chat.ID)
		}
		_, _ = bot.SendMessage(&telego.SendMessageParams{
			ChatID: message.Chat.ChatID(),
			Text:   response,
		})
	}, th.CommandEqual("add"))

	bh.HandleMessage(func(bot *telego.Bot, message telego.Message) {
		response, yuh := listIntoString(message.Chat.ID, "Sure, what subscriptions would you like to remove?")
		if yuh {
			removeCommand = append(removeCommand, message.Chat.ID)
		}
		_, _ = bot.SendMessage(&telego.SendMessageParams{
			ChatID: message.Chat.ChatID(),
			Text:   response,
		})
	}, th.CommandEqual("remove"))

	bh.HandleMessage(func(bot *telego.Bot, message telego.Message) {
		response, _ := listIntoString(message.Chat.ID, "List of active subscriptions:")
		_, _ = bot.SendMessage(&telego.SendMessageParams{
			ChatID: message.Chat.ChatID(),
			Text:   response,
		})
	}, th.CommandEqual("list"))

	bh.Start()
}
