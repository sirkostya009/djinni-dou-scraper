package main

import (
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	"os"
)

func main() {
	initMongo()
	bot := createBot()
	webhookUrl := os.Getenv("WEBHOOK_URL") + "/" + bot.Token()

	err := bot.SetWebhook(&telego.SetWebhookParams{
		URL: webhookUrl,
	})
	if err != nil {
		panic(err)
	}

	go func() {
		err = bot.StartWebhook(":443")
		if err != nil {
			panic(err)
		}
	}()

	updates, err := bot.UpdatesViaWebhook(webhookUrl)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = bot.StopWebhook()
		_ = bot.DeleteWebhook(&telego.DeleteWebhookParams{})
	}()

	bh, err := th.NewBotHandler(bot, updates)
	if err != nil {
		panic(err)
	}

	bh.HandleMessage(cancelHandler, th.CommandEqual("cancel"))
	bh.HandleMessage(addMessage, isAdding)
	bh.HandleMessage(removeMessage, isRemoving)
	bh.HandleMessage(addHandler, th.CommandEqual("add"))
	bh.HandleMessage(removeHandler, th.CommandEqual("remove"))
	bh.HandleMessage(listHandler, th.CommandEqual("list"))

	bh.Start()
}
