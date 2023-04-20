package handler

import (
	"log"

	"github.com/line/line-bot-sdk-go/v7/linebot"
)

func getDisplayNameFromID(userID string) string {
	res, err := bot.GetProfile(userID).Do()
	if err != nil {
		log.Println("UserID not Valid: ", err)
	}
	return res.DisplayName
}

func sendReply(bot *linebot.Client, event *linebot.Event, msg string) {
	if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(msg)).Do(); err != nil {
		log.Println("Send Replay err:", err)
	}
}

func getQuota(bot *linebot.Client) int64 {
	quota, err := bot.GetMessageQuota().Do()
	if err != nil {
		log.Println("Get quota err:", err)
	}
	return quota.Value
}

func invalidInputHandler(res string) string {
	return res
}
