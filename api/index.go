package api

import (
	"encoding/json"
	"fmt"
	"github.com/kkdai/youtube/v2"
	"net/http"
	"net/url"
	"os"

	_ "github.com/joho/godotenv/autoload"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	var Socks5Proxy = os.Getenv("SOCKS5_PROXY")

	ytClient := youtube.Client{}
	var client *http.Client
	if Socks5Proxy != "" {
		proxyURL, _ := url.Parse(Socks5Proxy)
		client = &http.Client{Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}}
		ytClient = youtube.Client{HTTPClient: client}
	}

	switch r.URL.Path {
	case "/":
		msg := `Welcome to ytDl API
/dl?url=<video_url> - Download a single video
/playlist?url=<playlist_url> - Download a playlist
			
Example:
/dl?url=https://www.youtube.com/watch?v=video_id
/dl?url=video_id
/playlist?url=https://www.youtube.com/playlist?list=playlist_id
/playlist?url=playlist_id

Made with ❤ by @Abishnoi69
Golang API for downloading YouTube videos and playlists
`
		if Socks5Proxy == "" {
			msg += "No SOCKS5 proxy configured, maybe you get rate limited by YouTube :("
		}
		_, _ = fmt.Fprint(w, msg)

	case "/dl":
		videoURL := r.URL.Query().Get("url")
		if videoURL == "" {
			http.Error(w, "Please provide a video URL", http.StatusBadRequest)
			return
		}

		video, err := ytClient.GetVideo(videoURL)
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		formats := video.Formats.WithAudioChannels()
		streamURL, err := ytClient.GetStreamURL(video, &formats[0])
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		response := map[string]string{
			"ID":          video.ID,
			"author":      video.Author,
			"duration":    video.Duration.String(),
			"thumbnail":   video.Thumbnails[0].URL,
			"description": video.Description,
			"stream_url":  streamURL,
			"title":       video.Title,
			"view_count":  fmt.Sprintf("%d", video.Views),
		}

		w.Header().Set("Content-Type", "application/json")
		if err = json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Error encoding JSON response: "+err.Error(), http.StatusInternalServerError)
		}

	case "/playlist":
		playlistURL := r.URL.Query().Get("url")
		if playlistURL == "" {
			http.Error(w, "Please provide a playlist URL", http.StatusBadRequest)
			return
		}

		playlist, err := ytClient.GetPlaylist(playlistURL)
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		var videos []map[string]string
		for _, entry := range playlist.Videos {
			video, err := ytClient.VideoFromPlaylistEntry(entry)
			if err != nil {
				http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
				return
			}

			streamURL, err := ytClient.GetStreamURL(video, &video.Formats[0])
			if err != nil {
				http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
				return
			}

			videoInfo := map[string]string{
				"ID":          video.ID,
				"author":      video.Author,
				"duration":    video.Duration.String(),
				"thumbnail":   video.Thumbnails[0].URL,
				"description": video.Description,
				"stream_url":  streamURL,
				"title":       video.Title,
				"view_count":  fmt.Sprintf("%d", video.Views),
			}
			videos = append(videos, videoInfo)
		}

		w.Header().Set("Content-Type", "application/json")
		if err = json.NewEncoder(w).Encode(videos); err != nil {
			http.Error(w, "Error encoding JSON response: "+err.Error(), http.StatusInternalServerError)
		}

	default:
		http.NotFound(w, r)
	}
}
