package handler

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/JohnsonYuanTW/NCAEats/models"
	"github.com/line/line-bot-sdk-go/v7/linebot"
	"gorm.io/gorm"
)

var bot *linebot.Client
var db *gorm.DB

// WIP: Refactor log to logrus
// TODO: change llog back to log
var llog = logrus.New()

var (
	ErrInputError         = errors.New("指令輸入錯誤，請重新輸入")
	ErrSystemError        = errors.New("系統有誤，請重新輸入")
	ErrRestaurantNotFound = errors.New("無此餐廳，請重新輸入")
	ErrOrderInProgress    = errors.New("目前有正在進行中的訂單，請重新輸入")
)

func CreateLineBot(channelSecret string, channelAccessToken string) {
	b, err := linebot.New(channelSecret, channelAccessToken)
	if err != nil {
		llog.WithError(err).Error("Bot creation error")
		panic(err)
	}
	bot = b
	llog.WithField("bot", GetBot().GetBotInfo()).Info("Bot Created")
}

func GetBot() *linebot.Client {
	return bot
}

func SetDB(d *gorm.DB) {
	db = d
}

func handleQuota(args []string) (string, error) {
	// Error handling
	if len(args) != 1 || args[0] != "" {
		return "", ErrInputError
	}

	// Get quota
	quota, err := getQuota(bot)
	if err != nil {
		llog.WithError(err).Error("無法獲取訊息額度")
		return "", errors.New("無法獲取訊息額度")
	}

	replyString := fmt.Sprintf("這個官方帳號尚有 %d 則訊息額度\n", quota)
	return replyString, nil
}

func handleNewOrder(args []string, ID string) (linebot.FlexContainer, error) {
	if len(args) != 1 || args[0] == "" {
		return nil, ErrInputError
	}

	restaurantName := args[0]
	restaurant, err := models.GetRestaurantByName(db, restaurantName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRestaurantNotFound
		}
		llog.WithError(err).Error("無法取得 %s 餐廳資訊", restaurantName)
		return nil, ErrSystemError
	}

	menuItems, err := models.GetMenuItemsByRestaurantName(db, restaurantName)
	if err != nil {
		llog.WithError(err).Error("無法取得 %s 的餐點項目", restaurantName)
		return nil, ErrSystemError
	}

	order, err := getActiveOrderOfIDWithErrorHandling(ID)
	if err != nil {
		return nil, err
	}
	if order != nil {
		return nil, ErrOrderInProgress
	}

	newOrder := &models.Order{
		Owner:      ID,
		Restaurant: restaurant,
	}
	if err = newOrder.CreateOrder(db); err != nil {
		llog.WithError(err).WithField("User", getDisplayNameFromID(ID)).Error("系統問題，無法開單")
		return nil, ErrSystemError
	}

	// Get and parse menuItemListFlexContainer.json
	menuItemListFlexContainer, err := getMenuItemListFlexContainer(restaurant)
	if err != nil {
		llog.WithError(err).WithField("Path", menuItemListFlexContainerPath).Error("無法解析 JSON")
		return nil, ErrSystemError
	}

	// Access the contents array in menuItemListFlexContainer
	bubbleContainer, ok := menuItemListFlexContainer.(*linebot.BubbleContainer)
	if !ok {
		return nil, ErrSystemError
	}

	// Add menuItems into container
	for _, item := range menuItems {
		newMenuItemBox, err := getMenuItemListBoxComponent(&item)
		if err != nil {
			llog.WithError(err).WithField("Path", menuItemListBoxComponentPath).Error("無法解析 JSON")
			return nil, ErrSystemError
		}

		bubbleContainer.Body.Contents = append(bubbleContainer.Body.Contents, &newMenuItemBox)
	}
	return menuItemListFlexContainer, nil
}

func handleNewOrderItem(args []string, ID string) (string, error) {
	var replyString string

	// Error handling
	if len(args) < 1 {
		return "", fmt.Errorf("指令輸入錯誤，請重新輸入")
	}

	// get username and display first part of response
	username := getDisplayNameFromID(ID)
	replyString = fmt.Sprintf("%s 點餐:\n", username)

	// Count number of orders
	if count, err := models.CountActiveOrderOfOwnerID(db, ID); err != nil {
		log.Printf("無法計算 %s 的訂單數量: %v", getDisplayNameFromID(ID), err)
		return "", fmt.Errorf("系統有誤，請重新輸入")
	} else if count != 1 {
		if count < 1 {
			return "", fmt.Errorf("目前沒有正在進行重的訂單")
		} else {
			// count > 1
			log.Printf("使用者 %s 目前有 %d 筆訂單", getDisplayNameFromID(ID), count)
			return "", fmt.Errorf("訂單狀態有誤，請重新輸入")
		}
	}

	// Get active order
	order, err := getActiveOrderOfIDWithErrorHandling(ID)
	if err != nil {
		return "", err
	}

	// Create order details
	var tailReplyString string
	for _, itemName := range args {
		if itemName == "" {
			continue
		} else if menuItem, err := models.GetMenuItemByNameAndRestaurantName(db, itemName, order.Restaurant.Name); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return "", fmt.Errorf("菜單中不包含餐點 %s，請重新輸入", itemName)
			}
			log.Printf("無法取得 %s 餐點資訊: %v", itemName, err)
			return "", fmt.Errorf("系統有誤，請重新輸入")
		} else {
			newOrderDetail := &models.OrderDetail{}
			newOrderDetail.Owner, newOrderDetail.Order, newOrderDetail.MenuItem = ID, order, menuItem
			newOrderDetail.CreateOrderDetail(db)
			replyString += fmt.Sprintf("%s 點餐成功\n", itemName)
		}
	}
	replyString += tailReplyString
	return replyString, nil
}

func handleNewRestaurant(args []string) (string, error) {
	// Error handling: check if there are any arguments
	if len(args) == 0 {
		return "", fmt.Errorf("指令輸入錯誤，請重新輸入")
	}

	var restaurants []*models.Restaurant
	// Loop through each argument and create a new restaurant
	for _, item := range args {
		// Split the argument into name and telephone number
		itemArgs := strings.Split(item, ",")
		// Check if the argument is valid
		if len(itemArgs) < 2 {
			return "", fmt.Errorf("指令輸入錯誤，請重新輸入")
		}
		name, tel := itemArgs[0], itemArgs[1]
		newRestaurant := &models.Restaurant{Name: name, Tel: tel}
		// Create the new restaurant in the database
		err := newRestaurant.CreateRestaurant(db)
		if err != nil {
			return "", fmt.Errorf("無法新增餐廳，請重新輸入")
		}
		// Add the new restaurant to the list of created restaurants
		restaurants = append(restaurants, newRestaurant)
	}

	// Check if any restaurants were actually created
	if len(restaurants) == 0 {
		return "", fmt.Errorf("沒有新增任何餐廳")
	}

	// Concatenate the reply string using a strings.Builder
	var sb strings.Builder
	for _, r := range restaurants {
		sb.WriteString(fmt.Sprintf("餐廳 %s 建立成功\n", r.Name))
	}
	return sb.String(), nil
}

func handleNewMenuItem(args []string) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("指令輸入錯誤，請重新輸入")
	}

	// Get restaurant
	restaurantName, items := args[0], args[1:]
	restaurant, err := models.GetRestaurantByName(db, restaurantName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", fmt.Errorf("無此餐廳 %s，請重新輸入", restaurantName)
		}
		log.Printf("無法取得 %s 餐廳資訊: %v", restaurantName, err)
		return "", fmt.Errorf("系統有誤，請重新輸入")
	}

	if len(items) == 0 {
		return "", fmt.Errorf("沒有新增任何餐點")
	}

	// Create menuitem
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("增加餐點至 %s\n", restaurantName))
	for _, item := range items {
		itemArgs := strings.Split(item, ",")
		if len(itemArgs) < 2 {
			return "", fmt.Errorf("指令輸入錯誤，請重新輸入")
		}
		name := itemArgs[0]
		price, err := strconv.Atoi(itemArgs[1])
		if err != nil {
			return "", fmt.Errorf("價格輸入錯誤，請重新輸入")
		}

		newMenuItem := &models.MenuItem{Name: name, Price: price, Restaurant: restaurant}
		if err := newMenuItem.CreateMenuItem(db); err != nil {
			return "", fmt.Errorf("無法新增餐點，請重新輸入")
		}

		sb.WriteString(fmt.Sprintf("餐點 %s %d 元\n", name, price))

	}
	return sb.String(), nil
}

func handleGetAllRestaurants(args []string) (linebot.FlexContainer, error) {
	// Error handling
	if len(args) > 1 || args[0] != "" {
		return nil, fmt.Errorf("指令輸入錯誤，請重新輸入")
	}

	// Get restaurant list
	restaurants, err := models.GetAllRestaurants(db)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		log.Printf("無法取得餐廳資訊: %v", err)
		return nil, fmt.Errorf("系統有誤，請重新輸入")
	}

	// Get and parse restaurantListFlexContainer.json
	restaurantListFlexContainer, err := getRestaurantListFlexContainer()
	if err != nil {
		log.Printf("無法解析 %s: %v", restaurantListFlexContainerPath, err)
		return nil, fmt.Errorf("系統有誤，請重新輸入")
	}

	// Add restaurant box into container
	for _, restaurant := range restaurants {
		newRestaurantBox, err := getRestaurantListBoxComponent(restaurant)
		if err != nil {
			log.Printf("無法解析 %s: %v", restaurantListBoxComponentPath, err)
			return nil, fmt.Errorf("系統有誤，請重新輸入")
		}

		restaurantListFlexContainer.(*linebot.BubbleContainer).Body.Contents = append(restaurantListFlexContainer.(*linebot.BubbleContainer).Body.Contents, &newRestaurantBox)
	}

	return restaurantListFlexContainer, nil
}

func getActiveOrderOfIDWithErrorHandling(ID string) (*models.Order, error) {
	// Get active order
	var orders []models.Order
	var err error
	username := getDisplayNameFromID(ID)
	if orders, err = models.GetActiveOrdersOfID(db, ID); err != nil {
		log.Printf("無法取得 %s 的訂單資訊: %v", username, err)
		return nil, fmt.Errorf("系統有誤，請重新輸入")
	}

	if count := len(orders); count == 1 {
		return &orders[0], nil
	} else if count < 1 {
		return nil, nil
	} else {
		// count > 1
		log.Printf("使用者 %s 目前有 %d 份訂單", username, count)
		return nil, fmt.Errorf("目前訂單數量有誤，請重新輸入")
	}
}

func handleClearOrder(args []string, ID string) (string, error) {
	// Error handling
	if len(args) > 1 || args[0] != "" {
		return "", fmt.Errorf("指令輸入錯誤，請重新輸入")
	}

	// Get active order
	order, err := getActiveOrderOfIDWithErrorHandling(ID)
	if err != nil {
		return "", err
	}
	if order == nil {
		return "", fmt.Errorf("目前沒有正在進行中的訂單，請重新輸入")
	}

	// Delete orderDetails and order
	err = models.DeleteOrderDetailsOfOrderID(db, order.ID)
	if err != nil {
		log.Printf("無法刪除 order_id 為 %d 的訂單細項: %v", order.ID, err)
		return "", fmt.Errorf("目前訂單狀況有誤，請重新輸入")
	}
	err = models.DeleteOrderOfID(db, order.ID)
	if err != nil {
		log.Printf("無法刪除 order_id 為 %d 的訂單: %v", order.ID, err)
		return "", fmt.Errorf("目前訂單狀況有誤，請重新輸入")
	}

	return "已清除訂單", nil
}

func handleStatistic(args []string, ID string) (string, error) {
	if len(args) > 1 || args[0] != "" {
		return "", fmt.Errorf("指令輸入錯誤，請重新輸入")
	}

	// Get active order
	order, err := getActiveOrderOfIDWithErrorHandling(ID)
	if err != nil {
		return "", err
	}
	if order == nil {
		return "", fmt.Errorf("目前沒有正在進行中的訂單，請重新輸入")
	}

	// Get order details
	replyString := fmt.Sprintf("%s:\n", order.Restaurant.Name)
	orderDetails, err := models.GetActiveOrderDetailsOfID(db, order.ID)
	if err != nil {
		log.Printf("無法取得 order_id 為 %d 的訂單細項: %v", order.ID, err)
		return "", fmt.Errorf("目前訂單狀況有誤，請重新輸入")
	}
	for _, od := range orderDetails {
		userName := getDisplayNameFromID(od.Owner)
		replyString += fmt.Sprintf("%s: %s / %d\n", userName, od.MenuItem.Name, od.MenuItem.Price)
	}
	return replyString, nil
}

func handleGetAllOrders(args []string, ID string) (string, error) {
	var replyString string
	if len(args) == 1 && args[0] == "" {
		replyString = "訂單列表:\n"
		orders, err := models.GetActiveOrders(db)
		if err != nil {
			log.Printf("無法取得所有訂單: %v", err)
			return "", fmt.Errorf("無法取得所有訂單，請重新輸入")
		}
		for _, order := range orders {
			username := getDisplayNameFromID(ID)
			replyString += fmt.Sprintf("%s: %s\n", username, order.Restaurant.Name)
		}
		return replyString, nil
	} else {
		return "", fmt.Errorf("指令輸入錯誤，請重新輸入")
	}
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
		case "額度":
			if rs, err := handleQuota(args); err != nil {
				replyString = err.Error()
			} else {
				replyString = rs
			}
		case "吃", "開":
			if container, err := handleNewOrder(args, ID); err != nil {
				replyString = err.Error()
			} else {
				sendReplyFlexMessage(bot, event, "開單", container)
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
				sendReplyFlexMessage(bot, event, "餐廳列表", container)
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
