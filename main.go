package main

import (
	"fmt"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	"os"
)

func main() {
	initMongo()
	bot := createBot(telego.WithDefaultDebugLogger())
	webhookEndpoint := "/" + bot.Token()

	err := bot.SetWebhook(&telego.SetWebhookParams{
		URL: "https://" + os.Getenv("WEBHOOK_URL") + webhookEndpoint,
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(bot.GetWebhookInfo())

	updates, err := bot.UpdatesViaWebhook(webhookEndpoint)
	if err != nil {
		panic(err)
	}

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

	go bh.Start()
	defer bh.Stop()
	_ = bot.StartWebhook("0.0.0.0:443")
	_ = bot.StopWebhook()
	_ = bot.DeleteWebhook(&telego.DeleteWebhookParams{})
}
