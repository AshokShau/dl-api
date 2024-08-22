# YouTube Downloader API

Welcome to the YouTube Downloader API. This API allows you to download video information and stream URLs from YouTube videos and playlists.

## Prerequisites

- Go 1.22.6 or higher

## Configuration
* `SOCKS5_PROXY` : SOCKS5 proxy URL (optional)

## Setup

1. Clone the repository:
   ```sh
   git clone https://github.com/Abishnoi69/ytDl-Api.git
   cd ytDl-Api
   ```
   
2. Run the server:
   ```sh
    go run main.go
    ```
   
3. The server will start on `http://localhost:8080`.


<section>
<h2>Deploy to Vercel</h2>
<ol>
<li>Fork this repository 🍴</li>
<li>Login your <a href="https://vercel.com/">Vercel</a> account </li>
<li>Go to your <a href="https://vercel.com/new">Add New Project</a></li>
<li>Choose the repository you forked</li>
<li>Tap on Deploy</li>
<li>Use your api and enjoy!</li>
</ol>
</section>


## Endpoints
> You can also use Video ID / Playlist ID instead of URL.

### Get Video Information:
* URL: /dl?url={video_url} 
* Method: GET
* Description: Download video information and stream URL.
* Response:
```
{
"ID": "video_id",
"author": "video_author",
"duration": "video_duration",
"thumbnail": "thumbnail_url",
"description": "video_description",
"stream_url": "stream_url",
"title": "video_title",
"view_count": "view_count"
}
```

### Get Playlist Information:
* URL: /playlist?url={playlist_url}
* Method: GET
* Description: Download playlist information and stream URLs.

## Contributing
Contributions are welcome! For bug reports, feature requests, or pull requests, please open an issue or submit your changes directly


## License
This project is licensed under the MIT License—see the [LICENSE](/LICENSE) file for details.
