package api

import (
	"encoding/json"
	"fmt"
	"github.com/kkdai/youtube/v2"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpproxy"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	_ "github.com/joho/godotenv/autoload"
)

var (
	Socks5Proxy   = os.Getenv("SOCKS5_PROXY")
	SourceCodeURL = "https://github.com/Abishnoi69/ytdl-api"
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
		video, err := handlerYouTube(w, r)
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
		data, err := handle(w, r)
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

var (
	mediaTypeRegex = regexp.MustCompile(`(?s)data-media-type="(.*?)"`)
	mainMediaRegex = regexp.MustCompile(`class="Content(.*?)src="(.*?)"`)
	captionRegex   = regexp.MustCompile(`(?s)class="Caption"(.*?)class="CaptionUsername".*data-log-event="captionProfileClick" target="_blank">(.*?)</a>(.*?)<div`)
	htmlTagRegex   = regexp.MustCompile(`<[^>]*>`)
)

type InstagramData *struct {
	ShortcodeMedia *ShortcodeMedia `json:"shortcode_media"`
	Data           struct {
		XDTShortcodeMedia *ShortcodeMedia `json:"xdt_shortcode_media"`
	} `json:"data,omitempty"`
}

type ShortcodeMedia struct {
	Typename              string                `json:"__typename"`
	ID                    string                `json:"id"`
	Shortcode             string                `json:"shortcode"`
	Dimensions            Dimensions            `json:"dimensions"`
	DisplayResources      []DisplayResources    `json:"display_resources"`
	IsVideo               bool                  `json:"is_video"`
	Title                 string                `json:"title"`
	VideoURL              string                `json:"video_url"`
	Owner                 Owner                 `json:"owner"`
	DisplayURL            string                `json:"display_url"`
	EdgeMediaToCaption    EdgeMediaToCaption    `json:"edge_media_to_caption"`
	EdgeSidecarToChildren EdgeSidecarToChildren `json:"edge_sidecar_to_children"`
	CoauthorProducers     *[]CoauthorProducers  `json:"coauthor_producers"`
}

type Owner struct {
	Username string `json:"username"`
}

type CoauthorProducers struct {
	Username string `json:"username"`
}

type Dimensions struct {
	Height int `json:"height"`
	Width  int `json:"width"`
}

type DisplayResources struct {
	ConfigWidth  int    `json:"config_width"`
	ConfigHeight int    `json:"config_height"`
	Src          string `json:"src"`
}

type EdgeMediaToCaption struct {
	Edges []Edges `json:"edges"`
}
type Edges struct {
	Node struct {
		Typename         string             `json:"__typename"`
		Text             string             `json:"text"`
		ID               string             `json:"id"`
		Shortcode        string             `json:"shortcode"`
		CommenterCount   int                `json:"commenter_count"`
		Dimensions       Dimensions         `json:"dimensions"`
		DisplayResources []DisplayResources `json:"display_resources"`
		IsVideo          bool               `json:"is_video"`
		VideoURL         string             `json:"video_url,omitempty"`
		DisplayURL       string             `json:"display_url"`
	} `json:"node"`
}

type EdgeSidecarToChildren struct {
	Edges []Edges `json:"edges"`
}

type StoriesData struct {
	URL string `json:"url"`
}

func getCaption(instagramData *ShortcodeMedia) string {
	if len(instagramData.EdgeMediaToCaption.Edges) == 0 {
		return ""
	}

	var sb strings.Builder
	if username := instagramData.Owner.Username; username != "" {
		sb.WriteString(fmt.Sprintf("<b>%v</b>", username))
	}

	if coauthors := instagramData.CoauthorProducers; coauthors != nil && len(*coauthors) > 0 {
		if sb.Len() > 0 {
			sb.WriteString(" <b>&</b> ")
		}
		for i, coauthor := range *coauthors {
			if i > 0 {
				sb.WriteString(" <b>&</b> ")
			}
			sb.WriteString(fmt.Sprintf("<b>%v</b>", coauthor.Username))
		}
	}

	if sb.Len() > 0 {
		sb.WriteString("<b>:</b>\n")
	}
	sb.WriteString(instagramData.EdgeMediaToCaption.Edges[0].Node.Text)

	return sb.String()
}

func getInstagramData(postID string) *ShortcodeMedia {
	if data := getEmbedData(postID); data != nil && data.ShortcodeMedia != nil {
		return data.ShortcodeMedia
	} else if data := getGQLData(postID); data != nil && data.Data.XDTShortcodeMedia != nil {
		return data.Data.XDTShortcodeMedia
	}
	return nil
}

func getEmbedData(postID string) InstagramData {
	var instagramData InstagramData

	body := Request(fmt.Sprintf("https://www.instagram.com/p/%v/embed/captioned/", postID), RequestParams{
		Method: "GET",
		Headers: map[string]string{
			"accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
			"accept-language": "en-US,en;q=0.9",
			"connection":      "close",
			"sec-fetch-mode":  "navigate",
			"User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36",
			"viewport-width":  "1280",
		},
	}).Body()
	if body == nil {
		return nil
	}

	if match := regexp.MustCompile(`\\"gql_data\\":([\s\S]*)}"}`).FindSubmatch(body); len(match) == 2 {
		s := strings.ReplaceAll(string(match[1]), `\"`, `"`)
		s = strings.ReplaceAll(s, `\\/`, `/`)
		s = strings.ReplaceAll(s, `\\`, `\`)

		if err := json.Unmarshal([]byte(s), &instagramData); err != nil {
			log.Print("[instagram/getEmbed] Error unmarshalling Instagram data: ", err)
		}
	}

	mediaTypeData := mediaTypeRegex.FindAllStringSubmatch(string(body), -1)
	if instagramData == nil && len(mediaTypeData) > 0 && mediaTypeData[0][1] == "GraphImage" {
		mainMediaData := mainMediaRegex.FindAllStringSubmatch(string(body), -1)
		mainMediaURL := strings.ReplaceAll(mainMediaData[0][2], "amp;", "")

		var caption, owner string
		captionData := captionRegex.FindAllStringSubmatch(string(body), -1)
		if len(captionData) > 0 {
			owner = strings.TrimSpace(htmlTagRegex.ReplaceAllString(captionData[0][2], ""))
			caption = strings.TrimSpace(htmlTagRegex.ReplaceAllString(captionData[0][3], ""))
		}

		dataJSON := fmt.Sprintf(`{
			"shortcode_media": {
				"__typename": "GraphImage",
				"display_url": "%v",
				"edge_media_to_caption": {
					"edges": [
						{
							"node": {
								"text": "%v"
							}
						}
					]
				},
				"owner": {
					"username": "%v"
				}
			}
		}`, mainMediaURL, caption, owner)

		if err := json.Unmarshal([]byte(dataJSON), &instagramData); err != nil {
			return nil
		}
	}

	return instagramData
}

func getGQLData(postID string) InstagramData {
	var instagramData InstagramData

	body := Request("https://www.instagram.com/api/graphql", RequestParams{
		Method: "POST",
		Headers: map[string]string{
			"User-Agent":         "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:120.0) Gecko/20100101 Firefox/120.0",
			"Accept":             "*/*",
			"Accept-Language":    "en-US;q=0.5,en;q=0.3",
			"Content-Type":       "application/x-www-form-urlencoded",
			"X-FB-Friendly-Name": "PolarisPostActionLoadPostQueryQuery",
			"X-CSRFToken":        "-m5n6c-w1Z9RmrGqkoGTMq",
			"X-IG-App-ID":        "936619743392459",
			"X-FB-LSD":           "AVp2LurCmJw",
			"X-ASBD-ID":          "129477",
			"DNT":                "1",
			"Sec-Fetch-Dest":     "empty",
			"Sec-Fetch-Mode":     "cors",
			"Sec-Fetch-Site":     "same-origin",
		},
		BodyString: []string{
			"av=0",
			"__d=www",
			"__user=0",
			"__a=1",
			"__req=3",
			"__hs=19734.HYP:instagram_web_pkg.2.1..0.0",
			"dpr=1",
			"__ccg=UNKNOWN",
			"__rev=1010782723",
			"__s=qg5qgx:efei15:ng6310",
			"__hsi=7323030086241513400",
			"__dyn=7xeUjG1mxu1syUbFp60DU98nwgU29zEdEc8co2qwJw5ux609vCwjE1xoswIwuo2awlU-cw5Mx62G3i1ywOwv89k2C1Fwc60AEC7U2czXwae4UaEW2G1NwwwNwKwHw8Xxm16wUxO1px-0iS2S3qazo7u1xwIwbS1LwTwKG1pg661pwr86C1mwrd6goK68jxe6V8",
			"__csr=gps8cIy8WTDAqjWDrpda9SoLHhaVeVEgvhaJzVQ8hF-qEPBV8O4EhGmciDBQh1mVuF9V9d2FHGicAVu8GAmfZiHzk9IxlhV94aKC5oOq6Uhx-Ku4Kaw04Jrx64-0oCdw0MXw1lm0EE2Ixcjg2Fg1JEko0N8U421tw62wq8989EMw1QpV60CE02BIw",
			"__comet_req=7",
			"lsd=AVp2LurCmJw",
			"jazoest=2989",
			"__spin_r=1010782723",
			"__spin_b=trunk",
			"__spin_t=1705025808",
			"fb_api_caller_class=RelayModern",
			"fb_api_req_friendly_name=PolarisPostActionLoadPostQueryQuery",
			"query_hash=b3055c01b4b222b8a47dc12b090e4e64",
			fmt.Sprintf(`variables={"shortcode": "%v","fetch_comment_count":2,"fetch_related_profile_media_count":0,"parent_comment_count":0,"child_comment_count":0,"fetch_like_count":10,"fetch_tagged_user_count":null,"fetch_preview_comment_count":2,"has_threaded_comments":true,"hoisted_comment_id":null,"hoisted_reply_id":null}`, postID),
			"server_timestamps=true",
			"doc_id=10015901848480474",
		},
	}).Body()

	if err := json.Unmarshal(body, &instagramData); err != nil {
		log.Print("Instagram: Error unmarshalling GQLData: ", err)
		return nil
	}

	return instagramData
}

func handle(w http.ResponseWriter, r *http.Request) (map[string]string, error) {
	videoId := r.URL.Query().Get("id")
	if videoId == "" {
		return nil, fmt.Errorf("please provide a video/reel/photo Id\nUsage: /instagram?id=videoId")
	}

	data := getInstagramData(videoId)
	if data == nil {
		return nil, fmt.Errorf("instagram data not found for post-ID: %v", videoId)
	}

	caption := getCaption(data)

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

	return response, nil
}

type RequestParams struct {
	Method     string            // "GET", "OPTIONS" or "POST"
	Proxy      bool              // Use proxy for the request
	Headers    map[string]string // Common headers for both GET and POST
	Query      map[string]string // Query parameters for GET
	BodyString []string          // Body of the request for POST
}

// Request sends a GET, OPTIONS or POST request to the specified link with the provided parameters and returns the response.
// The Link specifies the URL to send the request to.
// The params contain additional parameters for the request, such as headers, query parameters, and body.
// The Method field in params should be "GET" or "POST" to indicate the type of request.
//
// Example usage:
//
//	response := Request("https://api.example.com/users", RequestParams{
//		Method: "GET",
//		Headers: map[string]string{
//			"Authorization": "Bearer your-token",
//		},
//		Query: map[string]string{
//			"page":  "1",
//			"limit": "10",
//		},
//	})
//
//	response := Request("https://example.com/api", RequestParams{
//		Method: "POST",
//		Headers: map[string]string{
//			"Content-Type": "application/json",
//		},
//		BodyString: []string{
//			"param1=value1",
//			"param2=value2",
//		},
//	})
func Request(Link string, params RequestParams) *fasthttp.Response {
	request := fasthttp.AcquireRequest()
	response := fasthttp.AcquireResponse()

	client := &fasthttp.Client{
		ReadBufferSize:  16 * 1024,
		MaxConnsPerHost: 1024,
	}
	if params.Proxy {
		client.Dial = fasthttpproxy.FasthttpSocksDialer(Socks5Proxy)
	}

	request.Header.SetMethod(params.Method)
	for key, value := range params.Headers {
		request.Header.Set(key, value)
	}

	if params.Method == fasthttp.MethodGet {
		request.SetRequestURI(Link)
		for key, value := range params.Query {
			request.URI().QueryArgs().Add(key, value)
		}
	} else if params.Method == fasthttp.MethodOptions {
		request.SetRequestURI(Link)
		for key, value := range params.Query {
			request.URI().QueryArgs().Add(key, value)
		}
	} else if params.Method == fasthttp.MethodPost {
		request.SetBodyString(strings.Join(params.BodyString, "&"))
		request.SetRequestURI(Link)
	} else {
		log.Print("[request/Request] Error: Unsupported method ", params.Method)
		return response
	}

	err := client.Do(request, response)
	if err != nil {
		if strings.Contains(err.Error(), "missing port in address") {
			return response
		}
		log.Print("[request/Request] Error: ", err)
	}
	return response
}

func handlerYouTube(w http.ResponseWriter, r *http.Request) ([]map[string]string, error) {
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
