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

func (a *AppHandler) handleNewOrder(args []string, ID string) (linebot.FlexContainer, error) {
	if len(args) != 1 || args[0] == "" {
		return nil, ErrInputError
	}

	restaurantName := args[0]
	restaurant, err := a.RestaurantRepo.GetRestaurantByName(restaurantName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRestaurantNotFound
		}
		a.Logger.WithError(err).Error("無法取得 %s 餐廳資訊", restaurantName)
		return nil, ErrSystemError
	}

	menuItems, err := a.MenuItemRepo.GetMenuItemsByRestaurantName(restaurantName)
	if err != nil {
		a.Logger.WithError(err).Error("無法取得 %s 的餐點項目", restaurantName)
		return nil, ErrSystemError
	}

	order, err := a.getActiveOrderOfIDWithErrorHandling(ID)
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
	if err = a.OrderRepo.CreateOrder(newOrder); err != nil {
		a.Logger.WithError(err).WithField("User", a.getDisplayNameFromID(ID)).Error("系統問題，無法開單")
		return nil, ErrSystemError
	}

	// Get and parse menuItemListFlexContainer.json
	menuItemListFlexContainer, err := a.Templates.generateFlexContainer("menuItemListFlexContainer", restaurant.Name, restaurant.Tel)
	if err != nil {
		a.Logger.WithError(err).WithField("File", "menuItemListFlexContainer").Error("無法解析 JSON")
		return nil, ErrSystemError
	}

	// Access the contents array in menuItemListFlexContainer
	bubbleContainer, ok := menuItemListFlexContainer.(*linebot.BubbleContainer)
	if !ok {
		return nil, ErrSystemError
	}

	// Add menuItems into container
	for _, menuItem := range menuItems {
		newMenuItemBox, err := a.Templates.generateBoxComponent("menuItemListBoxComponent", menuItem.Name, menuItem.Price, menuItem.Name, menuItem.Name)
		if err != nil {
			a.Logger.WithError(err).WithField("File", "menuItemListBoxComponent").Error("無法解析 JSON")
			return nil, ErrSystemError
		}

		bubbleContainer.Body.Contents = append(bubbleContainer.Body.Contents, &newMenuItemBox)
	}
	return menuItemListFlexContainer, nil
}

func (a *AppHandler) handleNewOrderItem(args []string, ID string) (string, error) {
	var replyString string

	// Error handling
	if len(args) < 1 {
		return "", ErrInputError
	}

	// get username and display first part of response
	username := a.getDisplayNameFromID(ID)
	replyString = fmt.Sprintf("%s 點餐:\n", username)

	// Count number of orders
	if count, err := a.OrderRepo.CountActiveOrderOfOwnerID(ID); err != nil {
		a.Logger.WithError(err).WithField("User", a.getDisplayNameFromID(ID)).Error("無法計算訂單數量")
		return "", ErrSystemError
	} else if count != 1 {
		if count < 1 {
			return "", ErrNoOrderInProgress
		} else {
			// count > 1
			a.Logger.WithField("User", a.getDisplayNameFromID(ID)).Errorf("使用者目前有 %d 筆訂單", count)
			return "", ErrSystemError
		}
	}

	// Get active order
	order, err := a.getActiveOrderOfIDWithErrorHandling(ID)
	if err != nil {
		return "", err
	}

	// Create order details
	var tailReplyString string
	for _, itemName := range args {
		if itemName == "" {
			continue
		} else if menuItem, err := a.MenuItemRepo.GetMenuItemByNameAndRestaurantName(itemName, order.Restaurant.Name); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return "", ErrMenuItemNotFound
			}
			a.Logger.WithField("User", a.getDisplayNameFromID(ID)).Errorf("無法取得 %s 餐點資訊", itemName)
			return "", ErrSystemError
		} else {
			newOrderDetail := &models.OrderDetail{}
			newOrderDetail.Owner, newOrderDetail.Order, newOrderDetail.MenuItem = ID, order, menuItem
			a.OrderDetailRepo.CreateOrderDetail(newOrderDetail)
			replyString += fmt.Sprintf("%s 點餐成功\n", itemName)
		}
	}
	replyString += tailReplyString
	return replyString, nil
}

func (a *AppHandler) handleNewRestaurant(args []string) (string, error) {
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
		err := a.RestaurantRepo.CreateRestaurant(newRestaurant)
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

func (a *AppHandler) handleNewMenuItem(args []string) (string, error) {
	if len(args) < 2 {
		return "", ErrInputError
	}

	// Get restaurant
	restaurantName, items := args[0], args[1:]
	restaurant, err := a.RestaurantRepo.GetRestaurantByName(restaurantName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrRestaurantNotFound
		}
		a.Logger.WithError(err).Errorf("無法取得 %s 餐廳資訊", restaurantName)
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
		if err := a.MenuItemRepo.CreateMenuItem(newMenuItem); err != nil {
			return "", ErrNewMenuItemError
		}

		sb.WriteString(fmt.Sprintf("餐點 %s %d 元\n", name, price))

	}
	return sb.String(), nil
}

func (a *AppHandler) handleGetAllRestaurants(args []string) (linebot.FlexContainer, error) {
	// Error handling
	if len(args) > 1 || args[0] != "" {
		return nil, ErrInputError
	}

	// Get restaurant list
	restaurants, err := a.RestaurantRepo.GetAllRestaurants()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		a.Logger.WithError(err).Error("無法取得餐廳列表")
		return nil, ErrSystemError
	}

	// Get and parse restaurantListFlexContainer.json
	restaurantListFlexContainer, err := a.Templates.generateFlexContainer("restaurantListFlexContainer")
	if err != nil {
		a.Logger.WithError(err).Error("無法解析 restaurantListFlexContainer")
		return nil, ErrSystemError
	}

	// Add restaurant box into container
	for _, restaurant := range restaurants {
		restaurantListBoxComponent, err := a.Templates.generateBoxComponent("restaurantListBoxComponent", restaurant.Name, restaurant.Tel, restaurant.Name, restaurant.Name)
		if err != nil {
			a.Logger.WithError(err).Error("無法解析 restaurantListBoxComponent")
			return nil, ErrSystemError
		}

		restaurantListFlexContainer.(*linebot.BubbleContainer).Body.Contents = append(restaurantListFlexContainer.(*linebot.BubbleContainer).Body.Contents, &restaurantListBoxComponent)
	}

	return restaurantListFlexContainer, nil
}

func (a *AppHandler) getActiveOrderOfIDWithErrorHandling(ID string) (*models.Order, error) {
	// Get active order
	var orders []models.Order
	var err error
	username := a.getDisplayNameFromID(ID)
	if orders, err = a.OrderRepo.GetActiveOrdersOfID(ID); err != nil {
		a.Logger.WithError(err).Errorf("無法取得 %s 的訂單資訊", username)
		return nil, ErrSystemError
	}

	if count := len(orders); count == 1 {
		return &orders[0], nil
	} else if count < 1 {
		return nil, nil
	} else {
		// count > 1
		a.Logger.WithField("User", username).Errorf("使用者目前有 %d 筆訂單", count)
		return nil, ErrSystemError
	}
}

func (a *AppHandler) handleClearOrder(args []string, ID string) (string, error) {
	// Error handling
	if len(args) > 1 || args[0] != "" {
		return "", ErrInputError
	}

	// Get active order
	order, err := a.getActiveOrderOfIDWithErrorHandling(ID)
	if err != nil {
		return "", err
	}
	if order == nil {
		return "", ErrNoOrderInProgress
	}

	// Delete orderDetails and order
	err = a.OrderDetailRepo.DeleteOrderDetailsOfOrderID(order.ID)
	if err != nil {
		a.Logger.WithError(err).Errorf("無法刪除 ID %d 的訂單細項", order.ID)
		return "", ErrSystemError
	}
	err = a.OrderRepo.DeleteOrderOfID(order.ID)
	if err != nil {
		a.Logger.WithError(err).Errorf("無法刪除 ID %d 的訂單", order.ID)
		return "", ErrSystemError
	}

	return "已清除訂單", nil
}

// This function handles the statistic of an active order for a given ID.
func (a *AppHandler) handleStatistic(args []string, ID string) (string, error) {
	// Check if input is valid
	if len(args) > 1 || args[0] != "" {
		return "", ErrInputError
	}

	// Get active order
	order, err := a.getActiveOrderOfIDWithErrorHandling(ID)
	if err != nil {
		return "", err
	}
	if order == nil {
		return "", ErrNoOrderInProgress
	}

	// Get order details
	orderDetails, err := a.OrderDetailRepo.GetActiveOrderDetailsOfID(order.ID)
	if err != nil {
		a.Logger.WithError(err).Errorf("無法取得 ID %d 的訂單細項", order.ID)
		return "", ErrSystemError
	}

	// Calculate totals
	totals := calculateTotals(orderDetails)

	// Generate userReport
	var userReport strings.Builder
	fmt.Fprintf(&userReport, "%s<br>", order.Restaurant.Name)
	for _, od := range orderDetails {
		userName := a.getDisplayNameFromID(od.Owner)
		fmt.Fprintf(&userReport, "%s / %s / %d<br>", userName, od.MenuItem.Name, od.MenuItem.Price)
	}

	// Save userReport to a static HTML file
	userReportPath := "./static/userReport.html"
	if err := writeReportToFile(userReportPath, userReport.String()); err != nil {
		a.Logger.Printf("Could not write report to HTML file: %v", err)
		return "", err
	}
	userReportURL := "https://" + a.Env["SITE_URL"] + ":" + a.Env["PORT"] + "/userReport"

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

	return userReportURL + "\n\n" + restaurantReport.String(), nil
}

func (a *AppHandler) handleGetAllOrders(args []string, ID string) (string, error) {
	var replyString string
	if len(args) == 1 && args[0] == "" {
		replyString = "訂單列表:\n"
		orders, err := a.OrderRepo.GetActiveOrders()
		if err != nil {
			a.Logger.WithError(err).Error("無法取得所有訂單")
			return "", ErrSystemError
		}
		for _, order := range orders {
			username := a.getDisplayNameFromID(ID)
			replyString += fmt.Sprintf("%s: %s\n", username, order.Restaurant.Name)
		}
		return replyString, nil
	} else {
		return "", ErrInputError
	}
}

// writeReportToFile writes the provided content to a file at the specified path
func writeReportToFile(path string, content string) error {
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
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
