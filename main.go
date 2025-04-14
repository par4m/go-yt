package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"unicode"
)

const (
	APIKey     = "your_api_key_here"
	PlaylistID = "PLVItHqpXY_DArKRcfmGWykqV3u4hDaJLo" // Playlist ID
)

type Video struct {
	VideoTitle   string `json:"video_title"`
	VideoID      string `json:"video_id"`
	VideoURL     string `json:"video_url"`
	Description  string `json:"description"`
	LeetcodeLink string `json:"leetcode_link,omitempty"` // Optional field for LeetCode problems
	ProblemName  string `json:"problem_name,omitempty"`  // Optional field for LeetCode problems
	Difficulty   string `json:"difficulty,omitempty"`    // Optional field for LeetCode problems
	Topic        string `json:"topic,omitempty"`         // Optional field for LeetCode problems
}

func formatSlug(title string) string {
	// Normalize Unicode characters
	t := transform.Chain(
		norm.NFD,
		runes.Remove(runes.In(unicode.Mn)),
		norm.NFC,
	)
	normalized, _, _ := transform.String(t, title)

	// Slugify rules
	var slug strings.Builder
	for _, r := range strings.ToLower(normalized) {
		switch {
		case r >= 'a' && r <= 'z':
			slug.WriteRune(r)
		case r >= '0' && r <= '9':
			slug.WriteRune(r)
		case r == ' ' || r == '-':
			if slug.Len() == 0 || slug.String()[slug.Len()-1] != '-' {
				slug.WriteRune('-')
			}
		}
	}

	// Trim leading/trailing hyphens
	result := strings.Trim(slug.String(), "-")

	// Ensure non-empty slug
	if result == "" {
		return "untitled"
	}
	return result
}

func main() {
	ctx := context.Background()
	service, err := youtube.NewService(ctx, option.WithAPIKey(APIKey))
	if err != nil {
		log.Fatalf("Error creating YouTube client: %v", err)
	}

	videos := make([]Video, 0)
	nextPageToken := ""

	for {
		call := service.PlaylistItems.List([]string{"snippet"}).
			PlaylistId(PlaylistID).
			MaxResults(50).
			PageToken(nextPageToken)

		response, err := call.Do()
		if err != nil {
			log.Fatalf("Error fetching playlist items: %v", err)
		}

		for _, item := range response.Items {
			video := parseVideoMetadata(item)
			videos = append(videos, video)
		}

		nextPageToken = response.NextPageToken
		if nextPageToken == "" {
			break
		}
	}

	if err := saveToJSON(videos); err != nil {
		log.Fatal(err)
	}
}

func parseVideoMetadata(item *youtube.PlaylistItem) Video {
	snippet := item.Snippet

	video := Video{
		VideoTitle:  snippet.Title,
		VideoID:     snippet.ResourceId.VideoId,
		VideoURL:    fmt.Sprintf("https://www.youtube.com/watch?v=%s", snippet.ResourceId.VideoId),
		Description: snippet.Description,
	}

	// Extract LeetCode link and problem name from description first
	if leetcodeLink, problemName := extractLeetCodeLink(snippet.Description); leetcodeLink != "" {
		video.LeetcodeLink = leetcodeLink
		video.ProblemName = problemName
		video.Difficulty = extractDifficulty(snippet.Description, snippet.Title)
		video.Topic = extractTopic(snippet.Description)
		return video
	}

	// Fallback to title parsing for LeetCode problem details
	re := regexp.MustCompile(`LeetCode\s+(\d+):?\s+(.*?)\s*-\s*(Easy|Medium|Hard)`)
	if matches := re.FindStringSubmatch(snippet.Title); len(matches) > 3 {
		video.LeetcodeLink = fmt.Sprintf("https://leetcode.com/problems/%s", formatSlug(matches[2]))
		video.ProblemName = matches[2]
		video.Difficulty = matches[3]
		video.Topic = extractTopic(snippet.Description)
	}

	return video
}

func extractLeetCodeLink(description string) (string, string) {
	// Match both formats:
	// 1. Leetcode: https://leetcode.com/problems/minimum-absolute-difference/description/
	// 2. https://leetcode.com/problems/two-sum/
	re := regexp.MustCompile(`(?:Leetcode:\s*)?(https://leetcode\.com/problems/([^/]+)/?[^\s]*)`)
	matches := re.FindStringSubmatch(description)
	if len(matches) > 2 {
		cleanURL := strings.Split(matches[1], "/description")[0] // Remove /description suffix if present
		return cleanURL, formatProblemName(matches[2])
	}
	return "", ""
}

func formatProblemName(slug string) string {
	return strings.Title(strings.ReplaceAll(slug, "-", " "))
}

func extractDifficulty(description, title string) string {
	// Check both title and description for difficulty markers
	re := regexp.MustCompile(`#(easy|medium|hard)`)
	if matches := re.FindStringSubmatch(title + " " + description); len(matches) > 0 {
		return strings.Title(matches[1])
	}
	return "Unknown"
}

func extractTopic(description string) string {
	topics := []string{"Binary Tree", "Sorting", "Dynamic Programming", "Binary Search"}

	for _, topic := range topics {
		if containsIgnoreCase(description, topic) {
			return topic
		}
	}
	return "Other"
}

func containsIgnoreCase(str, substr string) bool {
	return strings.Contains(strings.ToLower(str), strings.ToLower(substr))
}

func saveToJSON(videos []Video) error {
	jsonData, err := json.MarshalIndent(videos, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %v", err)
	}

	err = os.WriteFile("youtube_playlist_videos.json", jsonData, 0644)
	if err != nil {
		return fmt.Errorf("error writing file: %v", err)
	}

	fmt.Printf("Successfully saved %d videos to youtube_playlist_videos.json\n", len(videos))
	return nil
}
