# Downloader API

## Features: 

* Instagram Data Retrieval:  
Fetches Instagram post-data using GraphQL queries.
Extracts and returns various details about an Instagram post, such as ID, caption, shortcode, dimensions, video URL, author, and more.

* YouTube Data Retrieval:  
Fetches YouTube video or playlist data using the [kkdai/youtube](https://github.com/kkdai/youtube) library.
Extracts and returns various details about a YouTube video, such as ID, author, duration, thumbnail, description, stream URL, title, and view count.
Supports fetching data for both individual videos and playlists.

* Proxy Support:  
Supports the use of a SOCKS5 proxy for YouTube data retrieval, configurable via the config.Socks5Proxy setting.

## Prerequisites:

- Go 1.23.0 or higher

## Configuration:

* `SOCKS5_PROXY` : SOCKS5 proxy URL (optional)

## Setup:

1. Clone the repository:
   ```sh
   git clone https://github.com/Abishnoi69/dl-api.git
   cd dl-api
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
<li>Configure your Environment Variables: <b>SOCKS5_PROXY</b></li>
<li>Tap on Deploy</li>
<li>Use your api and enjoy!</li>
</ol>
</section>


## Endpoints
> You can also use Video ID / Playlist ID instead of URL.

### Get Video Information:
* URL: /yt?url={video_url}/{playlist_url}  
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
"stream_url": "stream_url", // you can use this to download the video
"title": "video_title",
"view_count": "view_count"
...
}
```

### Get Instagram Post Information:
* URL: /ig?url={post_id} 
* Method: GET
* Description: Instagram post-information.
* Response:
```
{
"ID": "post_id",
"caption": "post_caption",
"shortcode": "post_shortcode",
"dimensions": "post_dimensions",
"video_url": "video_url", // you can use this to download the video
"author": "post_author"
...
}
```

## Contributing
Contributions are welcome! For bug reports, feature requests, or pull requests, please open an issue or submit your changes directly


## License
This project is licensed under the MIT License—see the [LICENSE](/LICENSE) file for details.
