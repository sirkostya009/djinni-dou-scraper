package main

import (
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	"os"
)

func main() {
	initDB()
	bot := createBot()

	updates, err := bot.UpdatesViaLongPolling(nil)
	if os.Getenv("WEBHOOK_URL") != "" {
		bot.StopLongPolling()
		webhookEndpoint := "/" + bot.Token()
		err = bot.SetWebhook(&telego.SetWebhookParams{
			URL: "https://" + os.Getenv("WEBHOOK_URL") + webhookEndpoint,
		})
		updates, err = bot.UpdatesViaWebhook(webhookEndpoint)
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

	if os.Getenv("WEBHOOK_URL") != "" {
		go func() {
			err = bot.StartWebhook("0.0.0.0:" + os.Getenv("PORT"))
			if err != nil {
				panic(err)
			}
		}()
		defer func() {
			err = bot.StopWebhook()
			if err != nil {
				bot.Logger().Errorf("Error stopping webhook: %v", err)
			}
		}()
		defer func() {
			err = bot.DeleteWebhook(&telego.DeleteWebhookParams{})
			if err != nil {
				bot.Logger().Errorf("Error deleting webhook: %v", err)
			}
		}()
	}
	bh.Start()
}
