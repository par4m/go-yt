# Microservice that powers https://dsa90.vercel.app

This is a Go-based microservice designed to automatically fetch videos from a YouTube playlist

```
const (
	APIKey     = "your_api_key_here"
	PlaylistID = "your_playlist_id" // Playlist ID
)
```

Edit main.go and add your YouTube API and Playlist_Id

- Automatically pull the videos from the playlist
- Filter
- Tags
- Leetcode Links
- Video URL
- Topic

Save videos to "youtube_videos_links.json" which is then used by dsa90days and automatically saved to the database

```
git clone https://github.com/par4m/go-yt
cd go-yt
go mod tidy  # Clean and update dependencies
go run main.go
```
