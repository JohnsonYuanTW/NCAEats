package handler

import (
	"errors"
	"fmt"
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
var log = logrus.New()

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

func CreateLineBot(channelSecret string, channelAccessToken string) {
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

func SetDB(d *gorm.DB) {
	db = d
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
		log.WithError(err).Error("無法取得 %s 餐廳資訊", restaurantName)
		return nil, ErrSystemError
	}

	menuItems, err := models.GetMenuItemsByRestaurantName(db, restaurantName)
	if err != nil {
		log.WithError(err).Error("無法取得 %s 的餐點項目", restaurantName)
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
		log.WithError(err).WithField("User", getDisplayNameFromID(ID)).Error("系統問題，無法開單")
		return nil, ErrSystemError
	}

	// Get and parse menuItemListFlexContainer.json
	menuItemListFlexContainer, err := generateFlexContainer(templates["menuItemListFlexContainer"], restaurant.Name, restaurant.Tel)
	if err != nil {
		log.WithError(err).WithField("File", "menuItemListFlexContainer").Error("無法解析 JSON")
		return nil, ErrSystemError
	}

	// Access the contents array in menuItemListFlexContainer
	bubbleContainer, ok := menuItemListFlexContainer.(*linebot.BubbleContainer)
	if !ok {
		return nil, ErrSystemError
	}

	// Add menuItems into container
	for _, menuItem := range menuItems {
		newMenuItemBox, err := generateBoxComponent(templates["menuItemListBoxComponent"], menuItem.Name, menuItem.Price, menuItem.Name, menuItem.Name)
		if err != nil {
			log.WithError(err).WithField("File", "menuItemListBoxComponent").Error("無法解析 JSON")
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
		return "", ErrInputError
	}

	// get username and display first part of response
	username := getDisplayNameFromID(ID)
	replyString = fmt.Sprintf("%s 點餐:\n", username)

	// Count number of orders
	if count, err := models.CountActiveOrderOfOwnerID(db, ID); err != nil {
		log.WithError(err).WithField("User", getDisplayNameFromID(ID)).Error("無法計算訂單數量")
		return "", ErrSystemError
	} else if count != 1 {
		if count < 1 {
			return "", ErrNoOrderInProgress
		} else {
			// count > 1
			log.WithField("User", getDisplayNameFromID(ID)).Errorf("使用者目前有 %d 筆訂單", count)
			return "", ErrSystemError
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
				return "", ErrMenuItemNotFound
			}
			log.WithField("User", getDisplayNameFromID(ID)).Errorf("無法取得 %s 餐點資訊", itemName)
			return "", ErrSystemError
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
		return "", ErrInputError
	}

	var restaurants []*models.Restaurant
	// Loop through each argument and create a new restaurant
	for _, item := range args {
		// Split the argument into name and telephone number
		itemArgs := strings.Split(item, ",")
		// Check if the argument is valid
		if len(itemArgs) < 2 {
			return "", ErrInputError
		}
		name, tel := itemArgs[0], itemArgs[1]
		newRestaurant := &models.Restaurant{Name: name, Tel: tel}
		// Create the new restaurant in the database
		err := newRestaurant.CreateRestaurant(db)
		if err != nil {
			return "", ErrNewRestaurantError
		}
		// Add the new restaurant to the list of created restaurants
		restaurants = append(restaurants, newRestaurant)
	}

	// Check if any restaurants were actually created
	if len(restaurants) == 0 {
		return "", ErrNewRestaurantError
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
		return "", ErrInputError
	}

	// Get restaurant
	restaurantName, items := args[0], args[1:]
	restaurant, err := models.GetRestaurantByName(db, restaurantName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrRestaurantNotFound
		}
		log.WithError(err).Errorf("無法取得 %s 餐廳資訊", restaurantName)
		return "", ErrSystemError
	}

	if len(items) == 0 {
		return "", ErrNewMenuItemError
	}

	// Create menuitem
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("增加餐點至 %s\n", restaurantName))
	for _, item := range items {
		itemArgs := strings.Split(item, ",")
		if len(itemArgs) < 2 {
			return "", ErrInputError
		}
		name := itemArgs[0]
		price, err := strconv.Atoi(itemArgs[1])
		if err != nil {
			return "", ErrInputError
		}

		newMenuItem := &models.MenuItem{Name: name, Price: price, Restaurant: restaurant}
		if err := newMenuItem.CreateMenuItem(db); err != nil {
			return "", ErrNewMenuItemError
		}

		sb.WriteString(fmt.Sprintf("餐點 %s %d 元\n", name, price))

	}
	return sb.String(), nil
}

func handleGetAllRestaurants(args []string) (linebot.FlexContainer, error) {
	// Error handling
	if len(args) > 1 || args[0] != "" {
		return nil, ErrInputError
	}

	// Get restaurant list
	restaurants, err := models.GetAllRestaurants(db)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		log.WithError(err).Error("無法取得餐廳列表")
		return nil, ErrSystemError
	}

	// Get and parse restaurantListFlexContainer.json
	restaurantListFlexContainer, err := generateFlexContainer(templates["restaurantListFlexContainer"])
	if err != nil {
		log.WithError(err).Error("無法解析 restaurantListFlexContainer")
		return nil, ErrSystemError
	}

	// Add restaurant box into container
	for _, restaurant := range restaurants {
		restaurantListBoxComponent, err := generateBoxComponent(templates["restaurantListBoxComponent"], restaurant.Name, restaurant.Tel, restaurant.Name, restaurant.Name)
		if err != nil {
			log.WithError(err).Error("無法解析 restaurantListBoxComponent")
			return nil, ErrSystemError
		}

		restaurantListFlexContainer.(*linebot.BubbleContainer).Body.Contents = append(restaurantListFlexContainer.(*linebot.BubbleContainer).Body.Contents, &restaurantListBoxComponent)
	}

	return restaurantListFlexContainer, nil
}

func getActiveOrderOfIDWithErrorHandling(ID string) (*models.Order, error) {
	// Get active order
	var orders []models.Order
	var err error
	username := getDisplayNameFromID(ID)
	if orders, err = models.GetActiveOrdersOfID(db, ID); err != nil {
		log.WithError(err).Errorf("無法取得 %s 的訂單資訊", username)
		return nil, ErrSystemError
	}

	if count := len(orders); count == 1 {
		return &orders[0], nil
	} else if count < 1 {
		return nil, nil
	} else {
		// count > 1
		log.WithField("User", username).Errorf("使用者目前有 %d 筆訂單", count)
		return nil, ErrSystemError
	}
}

func handleClearOrder(args []string, ID string) (string, error) {
	// Error handling
	if len(args) > 1 || args[0] != "" {
		return "", ErrInputError
	}

	// Get active order
	order, err := getActiveOrderOfIDWithErrorHandling(ID)
	if err != nil {
		return "", err
	}
	if order == nil {
		return "", ErrNoOrderInProgress
	}

	// Delete orderDetails and order
	err = models.DeleteOrderDetailsOfOrderID(db, order.ID)
	if err != nil {
		log.WithError(err).Errorf("無法刪除 ID %d 的訂單細項", order.ID)
		return "", ErrSystemError
	}
	err = models.DeleteOrderOfID(db, order.ID)
	if err != nil {
		log.WithError(err).Errorf("無法刪除 ID %d 的訂單", order.ID)
		return "", ErrSystemError
	}

	return "已清除訂單", nil
}

func handleStatistic(args []string, ID string) (string, error) {
	if len(args) > 1 || args[0] != "" {
		return "", ErrInputError
	}

	// Get active order
	order, err := getActiveOrderOfIDWithErrorHandling(ID)
	if err != nil {
		return "", err
	}
	if order == nil {
		return "", ErrNoOrderInProgress
	}

	// Get order details
	replyString := fmt.Sprintf("%s:\n", order.Restaurant.Name)
	orderDetails, err := models.GetActiveOrderDetailsOfID(db, order.ID)
	if err != nil {
		log.WithError(err).Errorf("無法取得 ID %d 的訂單細項", order.ID)
		return "", ErrSystemError
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
			log.WithError(err).Error("無法取得所有訂單")
			return "", ErrSystemError
		}
		for _, order := range orders {
			username := getDisplayNameFromID(ID)
			replyString += fmt.Sprintf("%s: %s\n", username, order.Restaurant.Name)
		}
		return replyString, nil
	} else {
		return "", ErrInputError
	}
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
