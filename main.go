package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"encoding/json"
	"time"
	"bytes"
)

type ReloadContinuationData struct {
	Continuation string
}
type Continuation struct {
	ReloadContinuationData ReloadContinuationData
}

type SubMenuItems struct {
	Title string
	Continuation Continuation
}

type ConfigInfo struct {
	AppInstallData string `json:"appInstallData"`
}

type Client struct {
	Hl string `json:"hl"`
	Gl string	`json:"gl"`
	RemoteHost string `json:"remoteHost"`
	DeviceMake string `json:"deviceMake"`
	DeviceModel string `json:"deviceModel"`
	VisitorData string `json:"visitorData"`
	UserAgent string `json:"userAgent"`
	ClientName string `json:"clientName"`
	ClientVersion string	`json:"clientVersion"`
	OsName string	`json:"osName"`
	OsVersion string `json:"osVersion"`
	OriginalUrl string `json:"originalUrl"`
	Platform string `json:"platform"`
	ClientFormFactor string `json:"clientFormFactor"`
	ConfigInfo ConfigInfo `json:"configInfo"`
}
type InnerTubeContext struct {
	Client Client `json:"client"`
}

type YtCfg struct {
	INNERTUBE_API_KEY string
	INNERTUBE_CONTEXT InnerTubeContext
	INNERTUBE_CONTEXT_CLIENT_NAME string
	INNERTUBE_CLIENT_VERSION string
	ID_TOKEN string
}

type Context struct {
	Context InnerTubeContext `json:"context"`
	Continuation string `json:"continuation"`
}

type ContinuationContents struct {
	LiveChatContinuation LiveChatContinuation `json:"liveChatContinuation"`
}

type LiveChatContinuation struct {
	Actions []Actions `json:"actions"`
	// Actions map[string]interface{} `json:"actions"`
}

type Actions struct {
	AddChatItemAction AddChatItemAction `json:"addChatItemAction"`
}

type AddChatItemAction struct {
	Item Item `json:"item"`
}

type Item struct {
	LiveChatTextMessageRenderer	LiveChatTextMessageRenderer	  `json:"liveChatTextMessageRenderer"`
}

type LiveChatTextMessageRenderer struct {
	Message Message `json:"message"`
}

type Message struct {
	Runs []Runs `json:"runs"`
}

type Runs struct {
	Text string `json:"text"`
}

type ChatMessagesResponse struct {
	ContinuationContents ContinuationContents `json:"continuationContents"`
}

const (
	API_TYPE = "live_chat"
	leftMatchRegex = `"INNERTUBE_API_KEY":"`
	rightMatchRegex = `",`
	ytCfgRe = `ytcfg\.set\s*\(\s*({.+?})\s*\)\s*;`
    _YT_INITIAL_DATA_RE = `(?:window\s*\[\s*["\']ytInitialData["\']\s*\]|ytInitialData)\s*=\s*({.+?})\s*;\s*(?:var\s+meta|</script|\n)`
	_YT_INITIAL_PLAYER_RESPONSE = `ytInitialPlayerResponse\s*=\s*({.+?})\s*;\s*(?:var\s+meta|</script|\n)`
)

func regexSearch(regex string, str string) []string {
	r, _ := regexp.Compile(regex)
	matches := r.FindAllString(str, -1)
	return matches
}

func FetchInitialData(videoUrl string) (string, string, string) {
	resp, err := http.Get(videoUrl)
	if err != nil {
		panic(err)
	}
	
	defer resp.Body.Close()

	intArr, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	html := string(intArr)
	// TODO ::  work on regex and do not use trims
	initialDataArr := regexSearch(_YT_INITIAL_DATA_RE, html)
	initialData := strings.Trim(initialDataArr[0], "ytInitialData = ")
	initialData = strings.Trim(initialData, ";</script")
	playerResponse := regexSearch(_YT_INITIAL_PLAYER_RESPONSE, html)[0]
	playerResponse = strings.Trim(playerResponse, "ytInitialPlayerResponse = ")
	playerResponse = strings.Trim(playerResponse, ";</sc")
	ytCfg := regexSearch(ytCfgRe, html)[0]
	ytCfg = strings.Trim(ytCfg, "ytcfg.set(")
	ytCfg = strings.Trim(ytCfg, ");")
	return initialData, playerResponse, ytCfg
}

func parseVideoData(playerResponseMap, initialDataMap  map[string]interface{}) ([]SubMenuItems) {
	_subMenuItems := initialDataMap["contents"].(map[string]interface{})["twoColumnWatchNextResults"].
	(map[string]interface{})["conversationBar"].(map[string]interface{})["liveChatRenderer"].(map[string]interface{})["header"].
	(map[string]interface{})["liveChatHeaderRenderer"].(map[string]interface{})["viewSelector"].
	(map[string]interface{})["sortFilterSubMenuRenderer"].(map[string]interface{})["subMenuItems"]
	_json, err := json.Marshal(_subMenuItems)
	if err != nil {
		panic(err)
	}
	jsonString := string(_json)
	var subMenuItems []SubMenuItems
	json.Unmarshal([]byte(jsonString), &subMenuItems)
	if err != nil {
		panic(err)
	}
	return subMenuItems
}

// Fetch chat messages from the given Youtube URL
func FetchChatMessages(initialContinuationInfo string, ytCfg YtCfg) () {
	// initialPageUrl := fmt.Sprintf("https://www.youtube.com/%s?continuation=%s", API_TYPE, initialContinuationInfo)
	// fmt.Println(initialPageUrl)
	apiKey := ytCfg.INNERTUBE_API_KEY
	continuationUrl := fmt.Sprintf("https://www.youtube.com/youtubei/v1/live_chat/get_%s?key=%s", API_TYPE, apiKey)
	innertubeContext := ytCfg.INNERTUBE_CONTEXT
	context := Context{innertubeContext, initialContinuationInfo}
	b, err := json.Marshal(context)
    if err != nil {
        fmt.Println(err)
    }
	// Now loop through all the chat messages
	for {
		var jsonData = []byte(b)
		request, error := http.NewRequest("POST", continuationUrl, bytes.NewBuffer(jsonData))
		request.Header.Set("Content-Type", "application/json; charset=UTF-8")

		client := &http.Client{}
		response, error := client.Do(request)
		if error != nil {
			panic(error)
		}
		defer response.Body.Close()

		if response.StatusCode != 200 {
			panic("Some error fetching chat messages")
		}
		body, _ := ioutil.ReadAll(response.Body)
		var bird ChatMessagesResponse
		err := json.Unmarshal([]byte(string(body)), &bird)
		if err != nil {
			panic(err)
		}
		actions := bird.ContinuationContents.LiveChatContinuation.Actions
		// iterate over actions 
		for _, action := range actions {
			runs := action.AddChatItemAction.Item.LiveChatTextMessageRenderer.Message.Runs
			if len(runs) > 0 && len(runs[0].Text) > 0 {
				fmt.Println(runs[0].Text)
			}
		}
		time.Sleep(time.Second * 5)
	}
}

func main() {
	// TODO:: Take youtube video videoUrl from CLI
	videoUrl := "https://www.youtube.com/watch?v=5qap5aO4i9A"
	initialData, playerResponse, _ytCfg := FetchInitialData(videoUrl)
	// parse the responses
	var ytCfg YtCfg
	json.Unmarshal([]byte(_ytCfg), &ytCfg)

	var playerResponseMap map[string]interface{}
	json.Unmarshal([]byte(playerResponse), &playerResponseMap)

	var initialDataMap map[string]interface{}
	json.Unmarshal([]byte(initialData), &initialDataMap)

	subMenuItems := parseVideoData(playerResponseMap, initialDataMap)
	initialContinuationInfo := subMenuItems[1].Continuation.ReloadContinuationData.Continuation

	FetchChatMessages(initialContinuationInfo, ytCfg)
}
	