package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	infoMsg := "This is a simple API to download videos from YouTube and Instagram.\n" +
		"Usage:\n" +
		"/yt?url=<video_url> - download video from YouTube\n" +
		"for yt use can use video_url, video_id, playlist_url or playlist_id\n" +
		"Instagram: /ig?url=<post_id> - get download video url from Instagram\n" +
		"Made with ❤️ by Abishnoi69\n" +
		"Source code: " + SourceCodeURL + "\n"

	switch r.URL.Path {
	case "/":
		if Socks5Proxy == "" {
			infoMsg += "NOTE: No SOCKS5 proxy configured, maybe you get rate limited by YouTube :("
		}

		_, err := fmt.Fprint(w, infoMsg)
		if err != nil {
			http.Error(w, "Error writing response: "+err.Error(), http.StatusInternalServerError)
			return
		}

	case "/yt":
		video, err := HandlerYouTube(w, r)
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if video == nil {
			http.Error(w, "Error: video is nil", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err = json.NewEncoder(w).Encode(video); err != nil {
			http.Error(w, "Error encoding JSON response: "+err.Error(), http.StatusInternalServerError)
		}

	case "/ig":
		data, err := Handle(w, r)
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err = json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, "Error encoding JSON response: "+err.Error(), http.StatusInternalServerError)
		}

	default:
		http.NotFound(w, r)
	}
}
