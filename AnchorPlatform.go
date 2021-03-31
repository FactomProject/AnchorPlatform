package main

import (
	"github.com/FactomProject/AnchorPlatform/api"
	"github.com/FactomProject/AnchorPlatform/config"
	"github.com/FactomProject/AnchorPlatform/factomSync"
)

func main() {
	conf := config.GetConfig()
	apiInstance := api.NewAPI(conf)
	sync := new(factomSync.Sync)
	go sync.Run(conf)
	_=apiInstance.Start()
}
