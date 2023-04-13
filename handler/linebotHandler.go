package handler

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/JohnsonYuanTW/NCAEats/models"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

var bot *linebot.Client

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
						// Error handling
						restaurantName := args[0]
						if restaurantName == "" {
							replyString = invalidInputHandler("指令輸入錯誤，請重新輸入")
							break
						}
						restaurant, ok := models.GetRestaurantByName(restaurantName)
						if !ok {
							replyString = invalidInputHandler("無此餐廳，請重新輸入")
							break
						}
						menuItems := models.GetMenuItemsByRestaurantName(restaurantName)
						if menuItems == nil {
							replyString = invalidInputHandler("系統有誤，請重新輸入")
							break
						}

						// Check order existance
						count, ok := models.CountActiveOrderOfID(event.Source.UserID)
						if !ok {
							replyString = invalidInputHandler("系統有誤，請重新輸入")
							break
						}
						if count > 0 {
							replyString = invalidInputHandler("目前有正在進行中的訂單，請重新輸入")
							break
						}

						// Create order
						newOrder := &models.Order{}
						newOrder.Owner, newOrder.Restaurant = event.Source.UserID, restaurant
						newOrder.CreateOrder()

						// Generating output
						replyString = fmt.Sprintf("開單囉，今天吃 %s\n", restaurant.Name)
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
						var order *models.Order
						if count, ok := models.CountActiveOrderOfID(event.Source.UserID); ok && count == 1 {
							order = models.GetActiveOrderOfID(event.Source.UserID)
						} else {
							replyString = invalidInputHandler("目前沒有正在進行中的訂單，請重新輸入")
							break
						}
						var tailReplyString string
						for _, itemName := range args {
							if itemName == "" {
								continue
							} else if menuItem, ok := models.GetMenuItemByNameAndRestaurantName(itemName, order.Restaurant.Name); !ok {
								tailReplyString += invalidInputHandler(fmt.Sprintf("※菜單中不包含餐點 %s，請重新輸入※\n", itemName))
								continue
							} else {
								newOrderDetail := &models.OrderDetail{}
								newOrderDetail.Owner, newOrderDetail.Order, newOrderDetail.MenuItem = event.Source.UserID, order, menuItem
								newOrderDetail.CreateOrderDetail()
								replyString += fmt.Sprintf("%s 點餐成功\n", itemName)
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
						restaurant, found := models.GetRestaurantByName(restaurantName)
						if !found {
							replyString = invalidInputHandler("無此餐廳，請重新輸入")
							break
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
								newMenuItem := &models.MenuItem{Name: name, Price: price, Restaurant: restaurant}
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
					case "清除":
						if len(args) > 1 || args[0] != "" {
							replyString = invalidInputHandler("指令輸入錯誤，請重新輸入")
							break
						}
						count, ok := models.CountActiveOrderOfID(event.Source.UserID)
						if count == 0 {
							replyString = invalidInputHandler("目前沒有正在進行中的訂單，請重新輸入")
							break
						} else if !ok {
							replyString = invalidInputHandler("系統錯誤，請重新輸入")
							break
						}
						if count, ok := models.CountActiveOrderOfID(event.Source.UserID); ok && count == 1 {
							order := models.GetActiveOrderOfID(event.Source.UserID)
							models.DeleteOrderDetailsOfOrderID(order.ID)
							models.DeleteOrderOfID(event.Source.UserID)
						} else {
							replyString = invalidInputHandler("目前沒有正在進行中的訂單，請重新輸入")
							break
						}

						replyString = fmt.Sprintf("已清除訂單")
					case "訂單":
						if len(args) == 1 && args[0] == "" {
							replyString = "訂單列表:\n"
							orders := models.GetActiveOrders()
							for _, order := range orders {
								username := getDisplaynameFromID(bot, event.Source.UserID)
								replyString += fmt.Sprintf("%s: %s\n", username, order.Restaurant.Name)
							}
						} else {
							replyString = invalidInputHandler("指令輸入錯誤，請重新輸入")
							break
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
