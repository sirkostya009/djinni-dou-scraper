package main

import (
	tg "github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	"os"
)

func main() {
	defer db.Close()

	bot, err := tg.NewBot(os.Getenv("TELEGRAM_BOT_TOKEN"))
	if err != nil {
		panic(err)
	}

	var updates <-chan tg.Update
	if url := os.Getenv("WEBHOOK_URL"); url != "" {
		webhookEndpoint := "/" + bot.Token()
		err = bot.SetWebhook(&tg.SetWebhookParams{
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
		err = bot.DeleteWebhook(&tg.DeleteWebhookParams{})
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

	h := &Handlers{}
	bh.HandleMessage(h.CancelHandler, th.CommandEqual("cancel"))
	bh.HandleMessage(h.AddMessage, h.IsAdding)
	bh.HandleMessage(h.RemoveMessage, h.IsRemoving)
	bh.HandleMessage(h.AddHandler, th.CommandEqual("add"))
	bh.HandleMessage(h.RemoveHandler, th.CommandEqual("remove"))
	bh.HandleMessage(h.ListHandler, th.CommandEqual("list"))
	bh.HandleMyChatMemberUpdated(h.StopHandler)

	bh.Start()
}
