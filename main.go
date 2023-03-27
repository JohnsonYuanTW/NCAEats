// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/joho/godotenv"
	"github.com/line/line-bot-sdk-go/v7/linebot"
	"golang.org/x/exp/slices"
)

var bot *linebot.Client

var restaurants []string = []string{"A", "B", "C"}
var menuItems []string = []string{"AA", "BB", "CC"}

func main() {
	var myEnv map[string]string
	var err error

	myEnv, err = godotenv.Read()
	if err != nil {
		log.Println("Env read err:", err)
	}

	bot, err = linebot.New(myEnv["ChannelSecret"], myEnv["ChannelAccessToken"])
	if err != nil {
		log.Println("Line bot err:", err)
	}
	log.Println("Bot Created:", bot)

	http.HandleFunc("/callback", callbackHandler)
	addr := fmt.Sprintf(":%s", myEnv["PORT"])
	http.ListenAndServeTLS(addr, myEnv["SSLCertfilePath"], myEnv["SSLKeyPath"], nil)
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
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
						quota := getQuota()
						replyString = fmt.Sprintf("這個官方帳號尚有 %d 則訊息額度\n", quota)
					case "吃", "開":
						if !slices.Contains(restaurants, args[0]) {
							replyString = invalidInputHandler("無此餐廳，請重新輸入")
							break
						}
						restaurant := args[0]
						replyString = fmt.Sprintf("開單囉，今天吃 %s", restaurant)
					case "點":
						if args[0] == "" {
							replyString = invalidInputHandler("指令輸入錯誤，請重新輸入")
							break
						}
						username := getDisplaynameFromID(event.Source.UserID)
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
					default:
						replyString = fmt.Sprint("無此指令，請重新輸入")
					}
					sendReply(event, replyString)
				}
			}
		}
	}
}

func getDisplaynameFromID(userID string) string {
	res, err := bot.GetProfile(userID).Do()
	if err != nil {
		log.Println("UserID not Valid: ", err)
	}
	return res.DisplayName
}

func sendReply(event *linebot.Event, msg string) {
	if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(msg)).Do(); err != nil {
		log.Print(err)
	}
}

func getQuota() int64 {
	quota, err := bot.GetMessageQuota().Do()
	if err != nil {
		log.Println("Get quota err:", err)
	}
	return quota.Value
}

func invalidInputHandler(res string) string {
	return res
}
