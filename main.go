package main

import (
	"fmt"
	"net/http"

	"github.com/JohnsonYuanTW/NCAEats/config"
	"github.com/JohnsonYuanTW/NCAEats/handler"
)

func main() {
	// Load environment variables
	env := config.Env

	// Create LineBot and serve http server
	handler.CreateLineBot(env["ChannelSecret"], env["ChannelAccessToken"])
	http.HandleFunc("/callback", handler.CallbackHandler)
	addr := fmt.Sprintf(":%s", env["PORT"])
	if err := http.ListenAndServeTLS(addr, env["SSLCertfilePath"], env["SSLKeyPath"], nil); err != nil {
		panic(err)
	}

}
