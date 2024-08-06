package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type journal struct {
	id         int
	user_text  string
	status     string
	open_date  string
	close_date string
}

func telegramBot() {
	// Подключение к базе данных
	db, err := sql.Open("sqlite3", "data/tgbase.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Портируем бота
	bot, err := tgbotapi.NewBotAPI("6731248572:AAHUv0lOq71xDfhLRfHEA8w-lBt5syXEblY")
	if err != nil {
		panic(err)
	}

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60
	updates := bot.GetUpdatesChan(updateConfig)

	currentTime := time.Now()

	for {
		// Обработка сообщений
		for update := range updates {
			if update.Message == nil {
				continue
			}

			// Если отправитель сообщения от администратора, сохраняем сообщение пользователя в базу данных
			if strings.HasPrefix(update.Message.Text, "#Заявка принята") || strings.HasPrefix(update.Message.Text, "#заявка принята") && update.Message.From.IsBot {
				update.Message.MessageID = update.Message.ReplyToMessage.MessageID
				_, err := db.Exec("insert into journal (user_text, open_date, close_date, status) values ($1, $2, '-', 'принято')", update.Message.ReplyToMessage.Text, currentTime.Format("2006-01-02"))
				if err != nil {
					panic(err)
				}
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ваша заявка принята!")
				msg.ReplyToMessageID = update.Message.MessageID
				bot.Send(msg)

			}

			if strings.HasPrefix(update.Message.Text, "#Заявка решена") || strings.HasPrefix(update.Message.Text, "#заявка решена") && update.Message.From.IsBot {
				update.Message.MessageID = update.Message.ReplyToMessage.MessageID
				_, err := db.Exec("update journal set status = 'Завершен', close_date = $1 where user_text = $2", currentTime.Format("2006-01-02"), update.Message.ReplyToMessage.Text)
				if err != nil {
					panic(err)
				}
			}
			username := update.Message.From.UserName
			if strings.HasPrefix(update.Message.Text, "#заявки") || strings.HasPrefix(update.Message.Text, "#Заявки") && strings.EqualFold(username, "AlexanderChekmarev") {
				rows, err := db.Query("SELECT * FROM journal")
				if err != nil {
					panic(err)
				}
				journals := []journal{}
				for rows.Next() {
					j := journal{}
					err := rows.Scan(&j.id, &j.user_text, &j.status, &j.open_date, &j.close_date)
					if err != nil {
						panic(err)
					}
					journals = append(journals, j)
				}
				var response []string
				for _, j := range journals {
					response = append(response, fmt.Sprintf("id: %d\nОписание: %s\nСтатус: %s\nДата открытия: %s\nДата закрытия: %s\n\n", j.id, j.user_text, j.status, j.open_date, j.close_date))
				}
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Список всех заявок:\n"+strings.Join(response, "\n"))
				bot.Send(msg)
			}
		}
	}
}

func main() {
	telegramBot()
}
