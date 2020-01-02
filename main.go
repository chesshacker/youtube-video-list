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

func main() {
	inputs := getProgramInputs()

	ctx := context.Background()
	service, err := youtube.NewService(ctx, option.WithAPIKey(inputs.apiKey))
	check(err)
	result := getVideos(service, inputs)
	for _, video := range result.Videos {
		updateVideoStats(service, video)
	}
	printVideos(result)
	// content, _ := json.Marshal(result)
	// fmt.Printf(string(content))
}

func getProgramInputs() (result programInputs) {
	result.apiKey = os.Getenv("APIKEY")
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
			MaxResults(50).
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

func updateVideoStats(service *youtube.Service, video *videoDetails) {
	call := service.Videos.List("statistics").Id(video.VideoId)

	response, err := call.Do()
	check(err)

	video.ViewCount = response.Items[0].Statistics.ViewCount
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
