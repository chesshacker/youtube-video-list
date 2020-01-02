# YouTube Video List

After a big conference like AWS re:Invent, there are literally hundreds of
videos available for all the sessions. While it would be great to watch all the
videos, I just don't have that much time. To help find the most interesting
videos and track my progress, I like to make a spreadsheet of the videos. This
tool creates a list of all the videos, including their title, view count and a
link to view the video.

To run this program, you will need the following:

* YouTube Data API Key - See the [YouTube Data API Overview] for more
  information on creating an API Key.
* YouTube Channel ID - See [Understanding your channel URLs] to learn more about
  Channel IDs. For example, the [AWS Events] Channel ID is
  `UCdoadna9HFHsxXWhafhNvKw`.
* Optionally, a range of dates you are interested in limiting results to. At
  this time, it appears re:Invent 2019 videos were posted between December 3 and
  December 26.

Before running, you need to copy `secrets.template.env` to `secrets.env` and
replace the placeholder with your YouTube Data API Key.

Then you can build and run the program:

```
make
set -o allexport; source secrets.env; set +o allexport
./youtube-video-list \
  -channel UCdoadna9HFHsxXWhafhNvKw \
  -after 2019-12-01T00:00:00Z \
  -before 2019-12-28T00:00:00Z
```

Note that this initial version runs each request to YouTube Data API in series,
and this example will run several hundred queries, which could take about a
minute or two to complete.

[YouTube Data API Overview]: https://developers.google.com/youtube/v3/getting-started
[AWS Events]: https://www.youtube.com/channel/UCdoadna9HFHsxXWhafhNvKw/videos
[Understanding your channel URLs]: https://support.google.com/youtube/answer/6180214?hl=en
