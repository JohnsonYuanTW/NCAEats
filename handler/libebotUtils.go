package handler

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

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
// Load JSON template
var templates map[string]string

func loadTemplates(dir string) error {
	templates = make(map[string]string)

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// load the file content
		content, err := ioutil.ReadFile(filepath.Join(dir, file.Name()))
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
