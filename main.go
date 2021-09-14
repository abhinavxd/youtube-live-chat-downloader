package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
)

const (
	leftMatchRegex = `"INNERTUBE_API_KEY":"`
	rightMatchRegex = `",`
)

// Fetches API key from the given Youtube URL
func FetchAPIKey(videoUrl string) (string, error) {
	resp, err := http.Get(videoUrl)
	if err != nil {
		panic(err)
	}
	
	defer resp.Body.Close()

	html, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	rx := regexp.MustCompile(`(?s)` + regexp.QuoteMeta(leftMatchRegex) + `(.*?)` + regexp.QuoteMeta(rightMatchRegex))
	matches := rx.FindAllStringSubmatch(string(html), -1)

	if len(matches) == 0 || len(matches[0]) == 0 {
		return "", fmt.Errorf("no matches found")
	}

	apiKey := matches[0][1]
	return apiKey, nil
}

// Fetch chat messages from the given Youtube URL
func FetchLiveStreamChat(videoId string) (string, error) {
	return "", nil
}

func main() {
	// TODO:: Take youtube video videoUrl from CLI
	videoUrl := "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
	key, err := FetchAPIKey(videoUrl)
	if err != nil {
		panic(err)
	}
	fmt.Printf("API key: %s\n", key)
}
		

