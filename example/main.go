package main

import (
	YtChat "github.com/abhinavxd/youtube-live-chat-downloader/v2"
	"fmt"
	"log"
)

func main() {
	continuation, cfg, error := YtChat.ParseInitialData("https://www.youtube.com/watch?v=5qap5aO4i9A")
	if error != nil {
		log.Fatal(error)
	}
	for {
		chat, newContinuation, error := YtChat.FetchContinuationChat(continuation, cfg)
		if error != nil {
			log.Print(error)
			continue
		}
		continuation = newContinuation
		for _, msg := range chat {
			fmt.Print(msg.Timestamp, " | ")
			fmt.Println(msg.AuthorName, ": ", msg.Message)
		}
	}
}
