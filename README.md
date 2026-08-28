# youtube-live-chat-downloader

Fetches Youtube live chat messages with no authentication required.

## How does it work?
* The request for fetching live chat is re-created by parsing the HTML content.
* Youtube API returns a `continuation` with new chat messages. This continuation is sent in the next API request to receive new messages and new continuation.

## Getting started 

	go get -u github.com/abhinavxd/youtube-live-chat-downloader/v2

```go
package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"

	YtChat "github.com/abhinavxd/youtube-live-chat-downloader/v2"
)

func main() {
	customCookies := []*http.Cookie{
		{Name: "PREF",
			Value:  "tz=Europe.Rome",
			MaxAge: 300},
		{Name: "CONSENT",
			Value:  fmt.Sprintf("YES+yt.432048971.it+FX+%d", 100+rand.Intn(999-100+1)),
			MaxAge: 300},
	}
	// Google would sometimes ask user to solve a CAPTCHA before accessing it's websites.
	// or ask for your CONSENT if you are an EU user
	// You can add those cookies here.
	// Adding cookies is OPTIONAL
	YtChat.AddCookies(customCookies)

	continuation, cfg, error := YtChat.ParseInitialData("https://www.youtube.com/watch?v=5qap5aO4i9A")
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


```

## Controlling the polling interval

`FetchContinuationChat` sleeps for the interval YouTube suggests (usually 10 s)
*before* returning, so messages reach your code later than they arrived over the
wire. If you need them as soon as possible, use `FetchContinuationChatNoWait`,
which returns immediately and hands you the suggested `timeoutMs` so you can
pace requests yourself:

```go
// How long to wait between requests. Lower means fresher messages.
const pollInterval = time.Second

for {
	start := time.Now()

	chat, newContinuation, timeoutMs, err := YtChat.FetchContinuationChatNoWait(continuation, cfg)
	if errors.Is(err, YtChat.ErrLiveStreamOver) {
		log.Fatal("Live stream over")
	}
	if err != nil {
		log.Print(err)
		time.Sleep(time.Second)
		continue
	}
	continuation = newContinuation

	for _, msg := range chat {
		fmt.Print(msg.Timestamp, " | ")
		fmt.Println(msg.AuthorName, ": ", msg.Message)
	}

	// Poll at pollInterval, or sooner if the server suggests a shorter wait.
	wait := pollInterval
	if suggested := time.Duration(timeoutMs) * time.Millisecond; suggested < wait {
		wait = suggested
	}
	if d := wait - time.Since(start); d > 0 {
		time.Sleep(d)
	}
}
```

Sleeping for the full `timeoutMs` instead reproduces the behaviour of
`FetchContinuationChat`.

Note that `timeoutMs` is a hint, not an enforced limit. Polling faster is
possible, but going too aggressive risks HTTP 429 responses, so handle that
case if you lower the interval.

## Screenshot
![Screenshot from 2021-10-25 09-40-04](https://user-images.githubusercontent.com/48166553/138645792-03baeb42-3eb9-4685-85f2-12c5ee694720.png)




<!-- LICENSE -->
## License

Distributed under the MIT License. See `LICENSE` for more information.


Based on chat-downloader by Xenova (https://github.com/xenova/chat-downloader)
