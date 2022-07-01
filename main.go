/*
 * *******************************
 * Copyright (c) 2022  Luke (github.com/itsLuuke). - All Rights Reserved
 *
 * Unauthorized copying or redistribution of this file in source and binary forms via any medium is strictly prohibited.
 * *******************************
 */

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	token := os.Getenv("TOKEN")
	if token == "" {
		panic("TOKEN environment variable is empty")
	}

	b, err := gotgbot.NewBot(token, &gotgbot.BotOpts{
		Client: http.Client{},
		DefaultRequestOpts: &gotgbot.RequestOpts{
			Timeout: gotgbot.DefaultTimeout,
			APIURL:  gotgbot.DefaultAPIURL,
		},
	})
	if err != nil {
		panic("failed to create new bot: " + err.Error())
	}

	updater := ext.NewUpdater(&ext.UpdaterOpts{
		ErrorLog: nil,
		DispatcherOpts: ext.DispatcherOpts{
			Error: func(b *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
				fmt.Println("an error occurred while handling update:", err.Error())
				return ext.DispatcherActionNoop
			},
			MaxRoutines: ext.DefaultMaxRoutines,
		},
	})
	dispatcher := updater.Dispatcher

	triggers := []rune{'/', '!', '>'}
	startCmd := handlers.NewCommand("start", start)
	startCmd.Triggers = triggers
	dispatcher.AddHandler(startCmd)
	helpCmd := handlers.NewCommand("help", start)
	helpCmd.Triggers = triggers
	dispatcher.AddHandler(helpCmd)
	dispatcher.AddHandler(handlers.NewMessage(func(msg *gotgbot.Message) bool {
		if msg.Sticker != nil && msg.Sticker.PremiumAnimation != nil && (msg.Chat.Type == "supergroup" || msg.Chat.Type == "group") {
			return true
		}
		return false
	}, delStk))

	// Start receiving updates.
	err = updater.StartPolling(b, &ext.PollingOpts{
		DropPendingUpdates: true,
		GetUpdatesOpts: gotgbot.GetUpdatesOpts{
			Timeout: 9,
			RequestOpts: &gotgbot.RequestOpts{
				Timeout: time.Second * 10,
			},
		},
	})
	if err != nil {
		panic("failed to start polling: " + err.Error())
	}
	fmt.Printf("%s has been started...\n", b.User.Username)

	// Idle, to keep updates coming in, and avoid bot stopping.
	updater.Idle()
}

// start introduces the bot.
func start(b *gotgbot.Bot, ctx *ext.Context) error {
	if ctx.EffectiveChat.Type == "private" {
		_, err := ctx.EffectiveMessage.Reply(b,
			fmt.Sprintf("Hello, I'm @%s. Add me to your chat with <b>Delete messages</b> permission so i can instantly delete annoying premium stickers!\n\nFor support join @TheBotsSupport", b.User.Username),
			&gotgbot.SendMessageOpts{
				ParseMode: "html",
			})
		if err != nil {
			return fmt.Errorf("failed to send start message: %w", err)
		}
	} else {
		_, err := ctx.EffectiveMessage.Reply(b,
			"Hello, give me <b>Delete messages</b> permission so i can instantly delete annoying premium stickers!",
			&gotgbot.SendMessageOpts{
				ParseMode: "html",
			})
		if err != nil {
			return fmt.Errorf("failed to send start message: %w", err)
		}
	}

	return nil
}
func delStk(b *gotgbot.Bot, ctx *ext.Context) (err error) {
	_, err = ctx.EffectiveMessage.Delete(b, nil)
	if err != nil {
		if err.Error() != "unable to deleteMessage: Bad Request: message can't be deleted" {
			log.Printf("error deleting sticker in chat: %d\n%s", ctx.EffectiveChat.Id, err.Error())
		}
	}
	return nil
}
