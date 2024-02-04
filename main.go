package main

import th "github.com/mymmrac/telego/telegohandler"

func main() {
	initMongo()
	bot := createBot()

	updates, err := bot.UpdatesViaLongPolling(nil)
	if err != nil {
		panic(err)
	}
	defer bot.StopLongPolling()

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
