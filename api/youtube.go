package api

import (
	"fmt"
	"github.com/kkdai/youtube/v2"
	"net/http"
	"net/url"
)

func HandlerYouTube(_ http.ResponseWriter, r *http.Request) ([]map[string]string, error) {
	videoURL := r.URL.Query().Get("url")
	if videoURL == "" {
		return nil, fmt.Errorf("please provide a video URL\nUsage: /youtube?url=videoUrl")
	}

	ytClient := youtube.Client{}
	if Socks5Proxy != "" {
		proxyURL, _ := url.Parse(Socks5Proxy)
		ytClient = youtube.Client{HTTPClient: &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyURL)}}}
	}

	var (
		playlist   *youtube.Playlist
		isPlaylist bool
		streamURL  string
		videos     []map[string]string
	)

	video, err := ytClient.GetVideo(videoURL)
	if err != nil {
		playlist, err = ytClient.GetPlaylist(videoURL)
		isPlaylist = true
		if err != nil {
			return nil, fmt.Errorf("error: %v", err)
		}
	}

	if isPlaylist {
		for _, entry := range playlist.Videos {
			video, err = ytClient.VideoFromPlaylistEntry(entry)
			if err != nil {
				return nil, fmt.Errorf("error: %v", err)
			}

			streamURL, err = ytClient.GetStreamURL(video, &video.Formats[0])
			if err != nil {
				return nil, fmt.Errorf("error: %v", err)
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

		return videos, nil

	} else {
		streamURL, err = ytClient.GetStreamURL(video, &video.Formats[0])
		if err != nil {
			return nil, fmt.Errorf("error: %v", err)
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

		return videos, nil
	}
}
