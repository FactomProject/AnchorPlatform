package main

import (
	"github.com/FactomProject/AnchorPlatform/api"
	"github.com/FactomProject/AnchorPlatform/config"
	"github.com/FactomProject/AnchorPlatform/factomSync"
)

func main() {
	conf := config.GetConfig()
	api := api.NewAPI(conf)
	go factomSync.Sync(conf)
	api.Start()
}
