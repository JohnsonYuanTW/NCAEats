package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/line/line-bot-sdk-go/v7/linebot"
	"gorm.io/gorm"
)

var bot *linebot.Client
var db *gorm.DB
var log = logrus.New()
var SITE_URL string

const templateDir = "./templates"

var (
	ErrInputError         = errors.New("指令輸入錯誤，請重新輸入")
	ErrSystemError        = errors.New("系統有誤，請重新輸入")
	ErrRestaurantNotFound = errors.New("無此餐廳，請重新輸入")
	ErrMenuItemNotFound   = errors.New("無此品項，請重新輸入")
	ErrOrderInProgress    = errors.New("目前有正在進行中的訂單，請重新輸入")
	ErrNoOrderInProgress  = errors.New("目前沒有正在進行中的訂單，請重新輸入")
	ErrNewRestaurantError = errors.New("無法新增餐廳")
	ErrNewMenuItemError   = errors.New("無法新增餐點")
)

func init() {
	loadTemplates(templateDir)
}

func CreateLineBot(channelSecret string, channelAccessToken string, siteUrl string) {
	SITE_URL = siteUrl
	b, err := linebot.New(channelSecret, channelAccessToken)
	if err != nil {
		log.WithError(err).Error("Bot creation error")
		panic(err)
	}
	bot = b
	log.WithField("bot", GetBot().GetBotInfo()).Info("Bot Created")
}

func GetBot() *linebot.Client {
	return bot
}

func CallbackHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	events, err := bot.ParseRequest(r)
	if err != nil {
		if err == linebot.ErrInvalidSignature {
			w.WriteHeader(400)
		} else {
			log.WithError(err).Error("Bot 有誤，無法解析請求")
			w.WriteHeader(500)
		}
		return
	}

	for _, event := range events {
		message, ok := event.Message.(*linebot.TextMessage)
		if !ok || !strings.Contains(message.Text, "/") {
			continue
		}
		// This is a text message event and containing "/"
		args := strings.Split(message.Text, "/")
		command, args := args[0], args[1:]
		var replyString string
		ID := event.Source.UserID
		switch command {
		case "吃", "開":
			if container, err := handleNewOrder(args, ID); err != nil {
				replyString = err.Error()
			} else {
				sendReply(bot, event, "開單", container)
				continue
			}
		case "點":
			if rs, err := handleNewOrderItem(args, ID); err != nil {
				replyString = err.Error()
			} else {
				replyString = rs
			}
		case "加餐廳":
			if rs, err := handleNewRestaurant(args); err != nil {
				replyString = err.Error()
			} else {
				replyString = rs
			}
		case "加餐點":
			if rs, err := handleNewMenuItem(args); err != nil {
				replyString = err.Error()
			} else {
				replyString = rs
			}
		case "餐廳":
			if container, err := handleGetAllRestaurants(args); errors.Is(err, gorm.ErrRecordNotFound) {
				replyString = "無此餐廳，請重新輸入"
			} else if err != nil {
				replyString = err.Error()
			} else {
				sendReply(bot, event, "餐廳列表", container)
				continue
			}
		case "清除":
			if rs, err := handleClearOrder(args, ID); err != nil {
				replyString = err.Error()
			} else {
				replyString = rs
			}
		case "統計":
			if rs, err := handleStatistic(args, ID); err != nil {
				replyString = err.Error()
			} else {
				replyString = rs
			}
		case "訂單":
			if rs, err := handleGetAllOrders(args, ID); err != nil {
				replyString = err.Error()
			} else {
				replyString = rs
			}
		default:
			replyString = "無此指令，請重新輸入"
		}
		// Remove trailing newline
		replyString = strings.TrimSuffix(replyString, "\n")
		sendReply(bot, event, replyString)
	}
}

func getDisplayNameFromID(userID string) string {
	res, err := bot.GetProfile(userID).Do()
	if err != nil {
		log.WithError(err).WithField("User", userID).Error("無法取得使用者 ID，請使用者加入好友")
		return userID
	}
	return res.DisplayName
}

func sendReply(bot *linebot.Client, event *linebot.Event, msg ...interface{}) {
	var err error

	switch len(msg) {
	case 1:
		// We expect a single string argument for a text message.
		text, ok := msg[0].(string)
		if !ok {
			log.Errorf("sendReply: 文字訊息僅可傳送字串，原文: %v", msg[0])
			return
		}
		_, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(text)).Do()
	case 2:
		// We expect two arguments (a string and a FlexContainer) for a flex message.
		altText, ok1 := msg[0].(string)
		flexContainer, ok2 := msg[1].(linebot.FlexContainer)
		if !ok1 || !ok2 {
			log.Errorf("sendReply: flex 訊息需有 altText 與 FlexContainer，原文: %v, %v", altText, flexContainer)
			return
		}
		_, err = bot.ReplyMessage(event.ReplyToken, linebot.NewFlexMessage(altText, flexContainer)).Do()
	default:
		log.Errorf("sendReply: 參數數量不正確，原文數量: %d，參數 ß0: %v", len(msg), msg[0])
		return
	}

	if err != nil {
		log.WithError(err).Error("無法傳送回覆")
	}
}
