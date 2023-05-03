package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/JohnsonYuanTW/NCAEats/config"
	"github.com/JohnsonYuanTW/NCAEats/handler"
	"github.com/JohnsonYuanTW/NCAEats/models"
)

func main() {
	// Load environment variables
	env, err := config.LoadEnvVariables()
	if err != nil {
		log.Fatalf("無法載入 env: %v", err)
	}

	// Initialize DB conn and models
	models.InitDB(env)

	// Create LineBot and serve http server
	handler.CreateLineBot(env["ChannelSecret"], env["ChannelAccessToken"])
	http.HandleFunc("/callback", handler.CallbackHandler)
	addr := fmt.Sprintf(":%s", env["PORT"])
	if err := http.ListenAndServeTLS(addr, env["SSLCertfilePath"], env["SSLKeyPath"], nil); err != nil {
		log.Printf("無法啟動 web server: %v", err)
		os.Exit(1)
	}
}
