package main

import (
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	"os"
)

func main() {
	initDB()
	bot := createBot()
	webhookEndpoint := "/" + bot.Token()

	var err error
	var updates <-chan telego.Update
	if os.Getenv("WEBHOOK_URL") != "" {
		err = bot.SetWebhook(&telego.SetWebhookParams{
			URL: "https://" + os.Getenv("WEBHOOK_URL") + webhookEndpoint,
		})
		updates, err = bot.UpdatesViaWebhook(webhookEndpoint)
	} else {
		updates, err = bot.UpdatesViaLongPolling(nil)
	}
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
	bh.HandleMyChatMemberUpdated(stopHandler)

	go bh.Start()
	defer bh.Stop()
	if os.Getenv("WEBHOOK_URL") != "" {
		_ = bot.StartWebhook("0.0.0.0:" + os.Getenv("PORT"))
		_ = bot.StopWebhook()
		_ = bot.DeleteWebhook(&telego.DeleteWebhookParams{})
	} else {
		select {}
	}
}
