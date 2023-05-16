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

// Flex response
const (
	restaurantListFlexContainerPath = "./templates/restaurantListFlexContainer.json"
	restaurantListBoxComponentPath  = "./templates/restaurantListBoxComponent.json"
	menuItemListFlexContainerPath   = "./templates/menuItemListFlexContainer.json"
	menuItemListBoxComponentPath    = "./templates/menuItemListBoxComponent.json"
)

func getFlexContainer(path string, data interface{}) (linebot.FlexContainer, error) {
	jsonData, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	template := string(jsonData)
	template = fmt.Sprintf(template, data)

	flexContainer, err := linebot.UnmarshalFlexMessageJSON([]byte(template))
	if err != nil {
		return nil, err
	}

	return flexContainer, nil
}

func getBoxComponent(path string, data interface{}) (linebot.BoxComponent, error) {
	jsonData, err := ioutil.ReadFile(path)
	if err != nil {
		return linebot.BoxComponent{}, err
	}

	template := string(jsonData)
	template = fmt.Sprintf(template, data)

	boxComponent := linebot.BoxComponent{}
	err = boxComponent.UnmarshalJSON([]byte(template))
	if err != nil {
		return linebot.BoxComponent{}, err
	}

	return boxComponent, nil
}

func getRestaurantListFlexContainer() (linebot.FlexContainer, error) {
	return getFlexContainer(restaurantListFlexContainerPath, "")
}

func getRestaurantListBoxComponent(restaurant *models.Restaurant) (linebot.BoxComponent, error) {
	return getBoxComponent(restaurantListBoxComponentPath, struct {
		Name string
		Tel  string
	}{
		Name: restaurant.Name,
		Tel:  restaurant.Tel,
	})
}

func getMenuItemListFlexContainer(restaurant *models.Restaurant) (linebot.FlexContainer, error) {
	return getFlexContainer(menuItemListFlexContainerPath, struct {
		Name string
		Tel  string
	}{
		Name: restaurant.Name,
		Tel:  restaurant.Tel,
	})
}

func getMenuItemListBoxComponent(menuItem *models.MenuItem) (linebot.BoxComponent, error) {
	return getBoxComponent(menuItemListBoxComponentPath, struct {
		Name  string
		Price int
	}{
		Name:  menuItem.Name,
		Price: menuItem.Price,
	})
}
