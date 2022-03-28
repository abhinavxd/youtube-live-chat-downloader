package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"

	YtChat "github.com/abhinavxd/youtube-live-chat-downloader/v2"
)

func main() {
	// Add your cookies here
	// Google would sometimes ask you to solve a CAPTCHA before accessing it's websites
	// or ask for your CONSENT if you are an EU user
	// You can add those cookies here.
	customCookies := []*http.Cookie{
		{Name: "PREF",
			Value:  "tz=Europe.Rome",
			MaxAge: 300},
		{Name: "CONSENT",
			Value:  fmt.Sprintf("YES+yt.432048971.it+FX+%d", 100+rand.Intn(999-100+1)),
			MaxAge: 300},
	}

	continuation, cfg, error := YtChat.ParseInitialData("https://www.youtube.com/watch?v=5qap5aO4i9A", customCookies)
	if error != nil {
		log.Fatal(error)
	}
	for {
		chat, newContinuation, error := YtChat.FetchContinuationChat(continuation, cfg)
		if error == YtChat.ErrLiveStreamOver {
			log.Fatal("Live stream over")
		}
		if error != nil {
			log.Print(error)
			continue
		}
		// set the newly received continuation
		continuation = newContinuation

		for _, msg := range chat {
			fmt.Print(msg.Timestamp, " | ")
			fmt.Println(msg.AuthorName, ": ", msg.Message)
		}
	}
}
