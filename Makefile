build:
	go build -o youtube-video-list main.go

clean:
	rm youtube-video-list

all: build
