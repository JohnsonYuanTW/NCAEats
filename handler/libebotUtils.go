package handler

import (
	"fmt"
	"os"

	"path/filepath"
	"strings"

	"github.com/line/line-bot-sdk-go/v7/linebot"
)

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

// Flex response
// Load JSON template
var templates map[string]string

func loadTemplates(dir string) error {
	templates = make(map[string]string)

	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// load the file content
		content, err := os.ReadFile(filepath.Join(dir, file.Name()))
		if err != nil {
			return err
		}

		// get the base name of the file (without extension)
		baseName := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))

		// store it in the map
		templates[baseName] = string(content)
	}

	return nil
}

// Unmarshal JSON template
type unmarshalFunc func([]byte) (interface{}, error)

func getComponent(template string, unmarshal unmarshalFunc, data ...interface{}) (interface{}, error) {
	// Insert data (if any) into the template
	if len(data) != 0 {
		template = fmt.Sprintf(template, data...)
	}

	// Parse JSON to linebot flex container
	component, err := unmarshal([]byte(template))
	if err != nil {
		return nil, err
	}

	return component, nil
}

func unmarshalFlexContainer(template []byte) (interface{}, error) {
	flexContainer, err := linebot.UnmarshalFlexMessageJSON(template)
	return flexContainer, err
}

func unmarshalBoxComponent(data []byte) (interface{}, error) {
	boxComponent := &linebot.BoxComponent{}
	err := boxComponent.UnmarshalJSON(data)
	return *boxComponent, err
}

func generateFlexContainer(template string, data ...interface{}) (linebot.FlexContainer, error) {
	flex, err := getComponent(template, unmarshalFlexContainer, data...)
	if err != nil {
		return nil, err
	}
	return flex.(linebot.FlexContainer), nil
}

func generateBoxComponent(template string, data ...interface{}) (linebot.BoxComponent, error) {
	box, err := getComponent(template, unmarshalBoxComponent, data...)
	if err != nil {
		return linebot.BoxComponent{}, err
	}
	return box.(linebot.BoxComponent), nil
}
