# youtube-live-chat-downloader

Fetches Youtube live chat messages without any authentication 


## Getting started 
```go
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
```

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
