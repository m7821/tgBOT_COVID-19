package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/buger/jsonparser"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func usage(program string) {
	fmt.Println("Usage: ", program, " [Telegram BOT token]")
}

func main() {

	if len(os.Args) != 2 {
		usage(os.Args[0])
		os.Exit(1)
	}

	bot, err := tgbotapi.NewBotAPI(os.Args[1])
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	var sent bool
	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		sent = false
		rsp, _ := http.Get("https://corona-stats.online/?format=json")
		stat, _ := ioutil.ReadAll(rsp.Body)

		jsonparser.ArrayEach(stat, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			country, _ := jsonparser.GetString(value, "country")
			cases, _ := jsonparser.GetInt(value, "cases")

			if country == update.Message.Text {
				flagPath, _ := jsonparser.GetString(value, "countryInfo", "flag")
				flag, _ := http.Get(flagPath)
				photo, _ := ioutil.ReadAll(flag.Body)
				tgPhoto := tgbotapi.FileBytes{Name: country, Bytes: photo}
				photomsg := tgbotapi.NewPhotoUpload(update.Message.Chat.ID, tgPhoto)
				bot.Send(photomsg)

				replyStr := fmt.Sprint(country, ",", cases)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, replyStr)

				bot.Send(msg)
				sent = true
			}

		}, "data")

		if sent == false {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Wrong Country")
			bot.Send(msg)
		}

	}
}
