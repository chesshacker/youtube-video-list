package main

import (
	"bufio"
	"context"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"html"
	"os"
	"strconv"
	"strings"

	"golang.org/x/exp/constraints"
	"google.golang.org/api/option"
	youtube "google.golang.org/api/youtube/v3"
)

type programInputs struct {
	apiKey          string
	channelId       string
	publishedBefore string
	publishedAfter  string
}

type videosResult struct {
	Videos []*videoDetails `json:"videos"`
}

type videoDetails struct {
	VideoId   string `json:"id"`
	Title     string `json:"title"`
	ViewCount uint64 `json:"viewCount"`
}

const maxResults int = 50

func main() {
	inputs := getProgramInputs()

	ctx := context.Background()
	service, err := youtube.NewService(ctx, option.WithAPIKey(inputs.apiKey))
	check(err)
	result := getVideos(service, inputs)
	updateVideoStats(service, &result)
	printVideos(result)
	// content, _ := json.Marshal(result)
	// fmt.Printf(string(content))
}

func getProgramInputs() (result programInputs) {
	result.apiKey = os.Getenv("YOUTUBE_APIKEY")
	flag.StringVar(&result.channelId, "channel", "", "YouTube Channel ID")
	flag.StringVar(&result.publishedBefore, "before", "", "Published before time, i.e. 2019-12-04T00:00:00Z")
	flag.StringVar(&result.publishedAfter, "after", "", "Published after time, i.e. 2019-12-03T00:00:00Z")
	flag.Parse()
	if result.apiKey == "" {
		check(errors.New("missing APIKEY environment variable"))
	}
	if result.channelId == "" {
		check(errors.New("missing -channelId argument"))
	}
	return result
}

func getVideos(service *youtube.Service, inputs programInputs) (result videosResult) {
	var nextPageToken string
	for {
		call := service.Search.List("snippet").
			Type("video").
			MaxResults(int64(maxResults)).
			Order("viewCount").
			ChannelId(inputs.channelId)
		if inputs.publishedBefore != "" {
			call.PublishedBefore(inputs.publishedBefore)
		}
		if inputs.publishedAfter != "" {
			call.PublishedAfter(inputs.publishedAfter)
		}
		if nextPageToken != "" {
			call.PageToken(nextPageToken)
		}

		response, err := call.Do()
		check(err)
		for _, item := range response.Items {
			result.Videos = append(result.Videos, &videoDetails{
				VideoId: item.Id.VideoId,
				Title:   html.UnescapeString(item.Snippet.Title),
			})
		}
		nextPageToken = response.NextPageToken
		if nextPageToken == "" {
			break
		}
	}
	return result
}

func updateVideoStats(service *youtube.Service, videoResults *videosResult) {
	videoIds := make([]string, len(videoResults.Videos))
	for i, video := range videoResults.Videos {
		videoIds[i] = video.VideoId
	}

	// Iterate through the video IDs in chunks of maxResults
	for i := 0; i < len(videoIds); i += maxResults {
		videoIdsJoined := strings.Join(videoIds[i:min(i+maxResults, len(videoIds))], ",")
		videosListCall := service.Videos.List("statistics").Id(videoIdsJoined)
		videosListResponse, err := videosListCall.Do()
		check(err)

		// Iterate through the video data and update the view count for each video
		for j, video := range videosListResponse.Items {
			videoResults.Videos[i+j].ViewCount = video.Statistics.ViewCount
		}
	}
}

func printVideos(result videosResult) {
	writer := csv.NewWriter(bufio.NewWriter(os.Stdout))
	headers := []string{
		"Views",
		"Title",
		"URL",
	}
	err := writer.Write(headers)
	check(err)
	for _, video := range result.Videos {
		row := []string{
			strconv.FormatUint(video.ViewCount, 10),
			video.Title,
			"https://www.youtube.com/watch?v=" + video.VideoId,
		}
		err := writer.Write(row)
		check(err)
	}
	writer.Flush()
}

func check(err error) {
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
}

func min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}
