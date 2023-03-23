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
	"strconv"

	"github.com/joho/godotenv"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

var bot *linebot.Client

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

	log.Print("Request Received. ")
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
			switch message := event.Message.(type) {
			// Handle only on text message
			case *linebot.TextMessage:
				// GetMessageQuota: Get how many remain free tier push message quota you still have this month. (maximum 500)
				quota, err := bot.GetMessageQuota().Do()
				if err != nil {
					log.Println("Quota err:", err)
				}
				// message.ID: Msg unique ID
				// message.Text: Msg text
				if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("msg ID:"+message.ID+":"+"Get:"+message.Text+" , \n OK! remain message:"+strconv.FormatInt(quota.Value, 10))).Do(); err != nil {
					log.Print(err)
				}

			// Handle only on Sticker message
			case *linebot.StickerMessage:
				var kw string
				for _, k := range message.Keywords {
					kw = kw + "," + k
				}

				outStickerResult := fmt.Sprintf("收到貼圖訊息: %s, pkg: %s kw: %s  text: %s", message.StickerID, message.PackageID, kw, message.Text)
				if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(outStickerResult)).Do(); err != nil {
					log.Print(err)
				}
			}
		}
	}
}
