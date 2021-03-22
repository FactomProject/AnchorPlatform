package config

import (
	"flag"
	"fmt"
	"os/user"
	"path/filepath"

	"github.com/jinzhu/configor"
	"github.com/mcuadros/go-defaults"
)

// Config structure
type Config struct {
	HTTPPort int `default:"8082" json:"httpport" form:"httpport" query:"httpport" required:"true"`
	Factom   struct {
		Server    string `default:"" json:"server" form:"server" query:"server" required:"false"`
		User      string `default:"" json:"user" form:"user" query:"user" required:"false"`
		Password  string `default:"" json:"password" form:"password" query:"password" required:"false"`
		EsAddress string `default:"" json:"esaddress" form:"esaddress" query:"esaddress" required:"false"`
	}
	Ledger struct {
		Bitcoin     Ledger
		Ethereum    Ledger
		BitcoinCash Ledger
	}
}

// Ledger is a generic sub-structure that reflects configuration for each ledger in the config
type Ledger struct {
	Endpoint   string `default:"" json:"endpoint" form:"endpoint" query:"endpoint" required:"false"`
	PrivateKey string `default:"" json:"privatekey" form:"privatekey" query:"privatekey" required:"false"`
}

// GetConfig returns config
func GetConfig() *Config {

	// Load AnchorPlatform config
	config := new(Config)
	defaults.SetDefaults(config)

	usr, err := user.Current()
	if err != nil {
		fmt.Println(err)
	}

	// Default config location
	configFile := filepath.Join(usr.HomeDir, ".AnchorMaker/config.yaml")

	// Check if custom config location passed as flag
	flag.StringVar(&configFile, "c", configFile, "config.yaml path")

	flag.Parse()

	fmt.Printf("Using config: %s\n", configFile)

	// Load custom configuration
	if err := configor.Load(config, configFile); err != nil {
		fmt.Printf("%s\n", err)
	}

	return config

}
