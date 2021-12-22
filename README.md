# youtube-live-chat-downloader

Fetches Youtube live chat messages with no authentication required.

## How does it work?
* The request for fetching live chat is re-created by parsing the HTML content.
* Youtube API returns a `continuation` with new chat messages. This continuation is sent in the next API request to receive new messages and new continuation.

## Getting started 
```go
package main

import (
	YtChat "github.com/abhinavxd/youtube-live-chat-downloader"
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
```

![Screenshot from 2021-10-25 09-40-04](https://user-images.githubusercontent.com/48166553/138645792-03baeb42-3eb9-4685-85f2-12c5ee694720.png)


<!-- CONTRIBUTING -->
## Contributing

Contributions are what make the open source community such an amazing place to learn, inspire, and create. Any contributions you make are **greatly appreciated**.

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request



<!-- LICENSE -->
## License

Distributed under the MIT License. See `LICENSE` for more information.
