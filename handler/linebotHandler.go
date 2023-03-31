package handler

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/JohnsonYuanTW/NCAEats/models"
	"github.com/line/line-bot-sdk-go/v7/linebot"
	"golang.org/x/exp/slices"
)

var bot *linebot.Client

var menuItems []string = []string{"AA", "BB", "CC"}

func CreateLineBot(channelSecret string, channelAccessToken string) {
	b, err := linebot.New(channelSecret, channelAccessToken)
	if err != nil {
		log.Println("Bot creation err:")
		panic(err)
	}
	bot = b
	log.Println("Bot Created:", GetBot())
}

func GetBot() *linebot.Client {
	return bot
}

func CallbackHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	events, err := bot.ParseRequest(r)
	if err != nil {
		log.Println("Line msg read err:", err)
	}
	if err != nil {
		if err == linebot.ErrInvalidSignature {
			w.WriteHeader(400)
		} else {
			w.WriteHeader(500)
		}
		return
	}

	for _, event := range events {
		if event.Type == linebot.EventTypeMessage {
			// This is a message event
			if message, ok := event.Message.(*linebot.TextMessage); ok {
				// This is a text message event
				if strings.Contains(message.Text, "/") {
					// Only deal with text messages containing "/"
					args := strings.Split(message.Text, "/")
					command, args := args[0], args[1:]
					var replyString string
					switch command {
					case "額度":
						if len(args) > 1 || args[0] != "" {
							replyString = invalidInputHandler("指令輸入錯誤，請重新輸入")
							break
						}
						quota := getQuota(bot)
						replyString = fmt.Sprintf("這個官方帳號尚有 %d 則訊息額度\n", quota)
					case "吃", "開":
						if args[0] == "" {
							replyString = invalidInputHandler("指令輸入錯誤，請重新輸入")
							break
						}
						restaurant, _, ok := models.GetRestaurantByName(args[0])
						if !ok {
							replyString = invalidInputHandler("無此餐廳，請重新輸入")
							break
						}
						replyString = fmt.Sprintf("開單囉，今天吃 %s\n", restaurant.Name)
						menuItems, _ := models.GetMenuItemsByRestaurantID(restaurant.ID)
						if menuItems == nil {
							replyString = invalidInputHandler("系統有誤，請重新輸入")
							break
						}
						for _, item := range menuItems {
							replyString += fmt.Sprintf("%s: %d 元\n", item.Name, item.Price)
						}
					case "點":
						if args[0] == "" {
							replyString = invalidInputHandler("指令輸入錯誤，請重新輸入")
							break
						}
						username := getDisplaynameFromID(bot, event.Source.UserID)
						replyString = fmt.Sprintf("%s 點餐:\n", username)
						var tailReplyString string
						for _, item := range args {
							if item == "" {
								continue
							} else if !slices.Contains(menuItems, item) {
								tailReplyString += invalidInputHandler(fmt.Sprintf("※菜單中不包含餐點 %s，請重新輸入※\n", item))
								continue
							} else {
								replyString += fmt.Sprintf("%s 點餐成功\n", item)
							}
						}
						replyString += tailReplyString
					case "加餐廳":
						if len(args) > 2 {
							replyString = invalidInputHandler("指令輸入錯誤，請重新輸入")
							break
						}
						newRestaurant := &models.Restaurant{}
						newRestaurant.Name, newRestaurant.Tel = args[0], args[1]
						newRestaurant.CreateRestaurant()
						replyString = fmt.Sprintf("餐廳 %s 建立成功", newRestaurant.Name)
					case "加餐點":
						restaurantName, items := args[0], args[1:]
						var restaurant *models.Restaurant
						if r, _, ok := models.GetRestaurantByName(restaurantName); !ok {
							replyString = invalidInputHandler("無此餐廳，請重新輸入")
							break
						} else {
							restaurant = r
						}
						if len(items) < 1 {
							replyString = invalidInputHandler("指令輸入錯誤，請重新輸入")
							break
						}
						replyString = fmt.Sprintf("增加餐點至 %s\n", restaurantName)
						for _, item := range items {
							itemArgs := strings.Split(item, ",")
							name := itemArgs[0]
							if len(itemArgs) < 1 {
								replyString = invalidInputHandler("指令輸入錯誤，請重新輸入")
								break
							}
							price, err := strconv.Atoi(itemArgs[1])
							if err != nil {
								replyString = invalidInputHandler("價格輸入錯誤，請重新輸入")
							} else {
								newMenuItem := &models.MenuItem{}
								newMenuItem.Name, newMenuItem.Price, newMenuItem.Restaurant = name, price, restaurant
								newMenuItem.CreateMenuItem()
								replyString += fmt.Sprintf("餐點 %s %d 元\n", name, price)
							}
						}
					case "餐廳":
						if len(args) > 1 || args[0] != "" {
							replyString = invalidInputHandler("指令輸入錯誤，請重新輸入")
							break
						}
						replyString = fmt.Sprintf("餐廳列表:\n")
						restaurants := models.GetAllRestaurants()
						for _, restrestaurant := range restaurants {
							replyString += fmt.Sprintln(restrestaurant.Name)
						}
					default:
						replyString = fmt.Sprint("無此指令，請重新輸入")
					}
					// Remove trailing newline
					replyString = strings.TrimSuffix(replyString, "\n")
					sendReply(bot, event, replyString)
				}
			}
		}
	}
}
