package handler

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/JohnsonYuanTW/NCAEats/models"
	"github.com/line/line-bot-sdk-go/v7/linebot"
	"gorm.io/gorm"
)

// TODO: Remove this and use dependency injection.
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

// This function handles the statistic of an active order for a given ID.
func handleStatistic(args []string, ID string) (string, error) {
	// Check if input is valid
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
	orderDetails, err := models.GetActiveOrderDetailsOfID(db, order.ID)
	if err != nil {
		log.WithError(err).Errorf("無法取得 ID %d 的訂單細項", order.ID)
		return "", ErrSystemError
	}

	// Calculate totals
	totals := calculateTotals(orderDetails)

	// Generate userReport
	var userReport strings.Builder
	fmt.Fprintf(&userReport, "%s<br>", order.Restaurant.Name)
	for _, od := range orderDetails {
		userName := getDisplayNameFromID(od.Owner)
		fmt.Fprintf(&userReport, "%s / %s / %d<br>", userName, od.MenuItem.Name, od.MenuItem.Price)
	}

	// Save userReport to a static HTML file
	userReportPath := "./static/userReport.html"
	if err := writeReportToFile(userReportPath, userReport.String()); err != nil {
		return "", err
	}

	// Generate restaurantReport
	var restaurantReport strings.Builder
	totalItemCount := 0
	totalPrice := 0

	fmt.Fprintf(&restaurantReport, "%s:\n", order.Restaurant.Name)
	for itemName, details := range totals {
		count := len(details)
		price := 0
		if count > 0 {
			price = details[0].MenuItem.Price * count
		}
		fmt.Fprintf(&restaurantReport, "%s / %d 份 / 共 %d 元\n", itemName, count, price)

		totalItemCount += count
		totalPrice += price
	}

	fmt.Fprintf(&restaurantReport, "總計: 共 %d 份 / 共 %d 元\n", totalItemCount, totalPrice)

	// return restaurantReport
	return SITE_URL + "/userReport\n\n" + restaurantReport.String(), nil
}

// writeReportToFile writes the provided content to a file at the specified path
func writeReportToFile(path string, content string) error {
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		log.Printf("Could not write report to HTML file: %v", err)
		return err
	}
	return nil
}

// This function calculates the totals for each item in the order.
func calculateTotals(orderDetails []*models.OrderDetail) map[string][]*models.OrderDetail {
	totals := make(map[string][]*models.OrderDetail)
	for _, od := range orderDetails {
		totals[od.MenuItem.Name] = append(totals[od.MenuItem.Name], od)
	}
	return totals
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
