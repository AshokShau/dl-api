package api

import (
	"os"

	_ "github.com/joho/godotenv/autoload"
)

var (
	Socks5Proxy   = os.Getenv("SOCKS5_PROXY")
	SourceCodeURL = "https://github.com/Abishnoi69/ytdl-api"
)
