package main

import YtChat "github.com/abhinavxd/youtube-live-chat-downloader"
import "fmt"

func main() {
	continuation, cfg := YtChat.ParseInitialData("https://www.youtube.com/watch?v=5qap5aO4i9A")
	for {
		chat, newContinuation := YtChat.FetchContinuationChat(continuation, cfg)
		continuation = newContinuation
		for _, msg := range chat {
			fmt.Println(msg.AuthorName, ": ", msg.Message)
		}
	}
}
