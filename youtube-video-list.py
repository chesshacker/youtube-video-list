import csv
import os
import argparse
import sys

from more_itertools import grouper
from googleapiclient.discovery import build


MAX_RESULTS = 50


def get_video_ids(youtube, channel_id, published_after=None, published_before=None):
    """
    Get a list of video IDs for a channel published between specified dates (optional).
    """
    video_ids = []
    next_page_token = ""

    while True:
        search_list = youtube.search().list(
            part="id",
            channelId=channel_id,
            type="video",
            order="date",
            publishedAfter=published_after,
            publishedBefore=published_before,
            maxResults=MAX_RESULTS,
            pageToken=next_page_token
        ).execute()
        video_ids.extend([item["id"]["videoId"]
                         for item in search_list["items"]])
        next_page_token = search_list.get("nextPageToken")

        # If there is no next page, break the loop
        if not next_page_token:
            break
    return video_ids


def get_video_data(youtube, video_ids):
    """
    Get the view count, title, and video URL for a list of video IDs.
    """
    video_data = []

    # Iterate through the video IDs in chunks of MAX_RESULTS
    for video_ids_chunk in grouper(video_ids, MAX_RESULTS):
        # Remove None values from the final chunk
        video_ids_joined = ",".join(
            [x for x in video_ids_chunk if x is not None])
        videos_list = youtube.videos().list(
            id=video_ids_joined,
            part="snippet,statistics"
        ).execute()

        # Iterate through the video data and add it to the list
        for video in videos_list["items"]:
            video_data.append({
                "Views": video["statistics"]["viewCount"],
                "Title": video["snippet"]["title"],
                "URL": f"https://www.youtube.com/watch?v={video['id']}"
            })

    # Sort the video data by view count in descending order
    video_data.sort(key=lambda x: int(x["Views"]), reverse=True)

    return video_data


def write_csv(video_data):
    """
    Write the video data to STDOUT as CSV.
    """
    with sys.stdout as csv_file:
        writer = csv.DictWriter(csv_file, fieldnames=["Views", "Title", "URL"])
        writer.writeheader()
        writer.writerows(video_data)


def get_inputs():
    """
    Get inputs from command line and environment variables.
    """
    parser = argparse.ArgumentParser(
        description="Get YouTube video data for a channel between and write it out in CSV format.")
    parser.add_argument(
        "--channel", help="The ID of the YouTube channel to search.")
    parser.add_argument(
        "--after", help="The published after date to search from in the format YYYY-MM-DDTHH:MM:SSZ.", default=None)
    parser.add_argument(
        "--before", help="The published before date to search from in the format YYYY-MM-DDTHH:MM:SSZ.", default=None)
    args = parser.parse_args()

    return {
        "api_key": os.environ.get("APIKEY"),
        "channel": args.channel,
        "after": args.after,
        "before": args.before,
    }


def main():
    inputs = get_inputs()
    youtube = build("youtube", "v3", developerKey=inputs['api_key'])
    video_ids = get_video_ids(
        youtube, inputs['channel'], inputs['after'], inputs['before'])
    video_data = get_video_data(youtube, video_ids)
    write_csv(video_data)


if __name__ == "__main__":
    main()
