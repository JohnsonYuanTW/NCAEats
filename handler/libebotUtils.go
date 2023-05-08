package handler

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/JohnsonYuanTW/NCAEats/models"
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
		log.Println("Send Text Replay err:", err)
	}
}

func sendReplyFlexMessage(bot *linebot.Client, event *linebot.Event, altText string, contents linebot.FlexContainer) {
	if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewFlexMessage(altText, contents)).Do(); err != nil {
		log.Println("Send Flex Replay err:", err)
	}
}

func getQuota(bot *linebot.Client) (int64, error) {
	quota, err := bot.GetMessageQuota().Do()
	if err != nil {
		return 0, err
	}
	return quota.Value, nil
}

// 餐廳
const restaurantListFlexContainerPath = "./templates/restaurantListFlexContainer.json"
const restaurantListBoxComponentPath = "./templates/restaurantListBoxComponent.json"

func getRestaurantListFlexContainer() (linebot.FlexContainer, error) {
	jsonData, err := ioutil.ReadFile(restaurantListFlexContainerPath)
	if err != nil {
		return nil, err
	}
	restaurantListFlexContainer, err := linebot.UnmarshalFlexMessageJSON(jsonData)
	if err != nil {
		return nil, err
	}

	return restaurantListFlexContainer, nil
}

func getRestaurantListBoxComponent(restaurant *models.Restaurant) (linebot.BoxComponent, error) {
	jsonData, err := ioutil.ReadFile(restaurantListBoxComponentPath)
	if err != nil {
		return linebot.BoxComponent{}, err
	}
	template := string(jsonData)
	template = fmt.Sprintf(template, restaurant.Name, restaurant.Tel, restaurant.Name, restaurant.Name)

	restaurantBoxComponent := linebot.BoxComponent{}
	err = restaurantBoxComponent.UnmarshalJSON([]byte(template))
	if err != nil {
		return linebot.BoxComponent{}, err
	}

	return restaurantBoxComponent, nil
}

// 開
const menuItemListFlexContainerPath = "./templates/menuItemListFlexContainer.json"
const menuItemListBoxComponentPath = "./templates/menuItemListBoxComponent.json"

func getMenuItemListFlexContainer(restaurant *models.Restaurant) (linebot.FlexContainer, error) {
	jsonData, err := ioutil.ReadFile(menuItemListFlexContainerPath)
	if err != nil {
		return nil, err
	}
	template := string(jsonData)
	template = fmt.Sprintf(template, restaurant.Name, restaurant.Tel)
	menuItemListFlexContainer, err := linebot.UnmarshalFlexMessageJSON([]byte(template))
	if err != nil {
		return nil, err
	}

	return menuItemListFlexContainer, nil
}

func getMenuItemListBoxComponent(menuItem *models.MenuItem) (linebot.BoxComponent, error) {
	jsonData, err := ioutil.ReadFile(menuItemListBoxComponentPath)
	if err != nil {
		return linebot.BoxComponent{}, err
	}
	template := string(jsonData)
	template = fmt.Sprintf(template, menuItem.Name, menuItem.Price, menuItem.Name, menuItem.Name)

	menuItemBoxComponent := linebot.BoxComponent{}
	err = menuItemBoxComponent.UnmarshalJSON([]byte(template))
	if err != nil {
		return linebot.BoxComponent{}, err
	}

	return menuItemBoxComponent, nil
}
