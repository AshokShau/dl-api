package api

import (
	"encoding/json"
	"fmt"
	"github.com/Abishnoi69/ytdl-api/api/config"
	"github.com/Abishnoi69/ytdl-api/api/instagram"
	"github.com/kkdai/youtube/v2"
	"net/http"
	"net/url"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	ytClient := youtube.Client{}
	if config.Socks5Proxy != "" {
		proxyURL, _ := url.Parse(config.Socks5Proxy)
		ytClient = youtube.Client{HTTPClient: &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyURL)}}}
	}

	switch r.URL.Path {
	case "/":
		infoMsg := fmt.Sprintf("Welcome to ytDl API\n/dl?url=<video_url> - Download a single video\n/playlist?url=<playlist_url> - Download a playlist\nExample:\n/dl?url=https://www.youtube.com/watch?v=video_id\n/dl?url=video_id\n/playlist?url=https://www.youtube.com/playlist?list=playlist_id\n/playlist?url=playlist_id\n/instagram?id=instagram_video/reel/photo_id\nMade with ❤ by @Abishnoi69\nGolang API for downloading YouTube videos and playlists\n")
		if config.Socks5Proxy == "" {
			infoMsg += "No SOCKS5 proxy configured, maybe you get rate limited by YouTube :("
		}

		w.Header().Set("Content-Type", "application/json")
		_, err := fmt.Fprint(w, infoMsg)
		if err != nil {
			http.Error(w, "Error writing response: "+err.Error(), http.StatusInternalServerError)
			return
		}

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

		streamURL, err := ytClient.GetStreamURL(video, &video.Formats.WithAudioChannels()[0])
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

			videos = append(videos, map[string]string{
				"ID":          video.ID,
				"author":      video.Author,
				"duration":    video.Duration.String(),
				"thumbnail":   video.Thumbnails[0].URL,
				"description": video.Description,
				"stream_url":  streamURL,
				"title":       video.Title,
				"view_count":  fmt.Sprintf("%d", video.Views),
			})
		}

		w.Header().Set("Content-Type", "application/json")
		if err = json.NewEncoder(w).Encode(videos); err != nil {
			http.Error(w, "Error encoding JSON response: "+err.Error(), http.StatusInternalServerError)
		}

	case "/instagram":
		data, caption, err := instagram.Handle(w, r)
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		response := map[string]string{
			"ID":                       data.ID,
			"caption":                  caption,
			"shortCode":                data.Shortcode,
			"dimensions":               fmt.Sprintf("%dx%d", data.Dimensions.Width, data.Dimensions.Height),
			"is_video":                 fmt.Sprintf("%t", data.IsVideo),
			"title":                    data.Title,
			"video_url":                data.VideoURL,
			"author":                   data.Owner.Username,
			"displayURL":               data.DisplayURL,
			"display_resources":        fmt.Sprintf("%v", data.DisplayResources),
			"edge_media_to_caption":    fmt.Sprintf("%v", data.EdgeMediaToCaption.Edges),
			"edge_sidecar_to_children": fmt.Sprintf("%v", data.EdgeSidecarToChildren.Edges),
			"coauthor_producers":       fmt.Sprintf("%v", data.CoauthorProducers),
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Error encoding JSON response: "+err.Error(), http.StatusInternalServerError)
		}

	default:
		http.NotFound(w, r)
	}
}
