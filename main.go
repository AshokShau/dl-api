package main

import (
	"fmt"
	"github.com/Abishnoi69/ytdl-api/api"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", api.Handler)
	http.HandleFunc("/dl", api.Handler)
	http.HandleFunc("/playlist", api.Handler)
	http.HandleFunc("/instagram", api.Handler)

	fmt.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
