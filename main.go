package main

import (
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	"os"
)

func main() {
	defer db.Close()

	bot, err := telego.NewBot(os.Getenv("TELEGRAM_BOT_TOKEN"))
	if err != nil {
		panic(err)
	}

	var updates <-chan telego.Update
	if url := os.Getenv("WEBHOOK_URL"); url != "" {
		webhookEndpoint := "/" + bot.Token()
		err = bot.SetWebhook(&telego.SetWebhookParams{
			URL: "https://" + url + webhookEndpoint,
		})
		if err != nil {
			panic(err)
		}
		updates, err = bot.UpdatesViaWebhook(webhookEndpoint)
		go func() {
			err = bot.StartWebhook("0.0.0.0:" + os.Getenv("PORT"))
			if err != nil {
				panic(err)
			}
		}()
	} else {
		err = bot.DeleteWebhook(&telego.DeleteWebhookParams{})
		if err != nil {
			bot.Logger().Errorf("Error deleting webhook: %v", err)
		}
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

	bh.Start()
}
