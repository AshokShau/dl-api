package config

import (
	"os"

	_ "github.com/joho/godotenv/autoload"
)

var Socks5Proxy = os.Getenv("SOCKS5_PROXY")
