package YtChat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

type ReloadContinuationData struct {
	Continuation string
}
type Continuation struct {
	ReloadContinuationData ReloadContinuationData
}

type SubMenuItems struct {
	Title        string
	Continuation Continuation
}

type ConfigInfo struct {
	AppInstallData string `json:"appInstallData"`
}

type Client struct {
	Hl               string     `json:"hl"`
	Gl               string     `json:"gl"`
	RemoteHost       string     `json:"remoteHost"`
	DeviceMake       string     `json:"deviceMake"`
	DeviceModel      string     `json:"deviceModel"`
	VisitorData      string     `json:"visitorData"`
	UserAgent        string     `json:"userAgent"`
	ClientName       string     `json:"clientName"`
	ClientVersion    string     `json:"clientVersion"`
	OsName           string     `json:"osName"`
	OsVersion        string     `json:"osVersion"`
	OriginalUrl      string     `json:"originalUrl"`
	Platform         string     `json:"platform"`
	ClientFormFactor string     `json:"clientFormFactor"`
	ConfigInfo       ConfigInfo `json:"configInfo"`
}
type InnerTubeContext struct {
	Client Client `json:"client"`
}

type YtCfg struct {
	INNERTUBE_API_KEY             string
	INNERTUBE_CONTEXT             InnerTubeContext
	INNERTUBE_CONTEXT_CLIENT_NAME string
	INNERTUBE_CLIENT_VERSION      string
	ID_TOKEN                      string
}

type Context struct {
	Context      InnerTubeContext `json:"context"`
	Continuation string           `json:"continuation"`
}

type ContinuationContents struct {
	LiveChatContinuation LiveChatContinuation `json:"liveChatContinuation"`
}

type ContinuationChat struct {
	TimedContinuationData TimedContinuationData `json:"timedContinuationData"`
}

type TimedContinuationData struct {
	Continuation string
	TimeoutMs    int
}
type LiveChatContinuation struct {
	Actions       []Actions          `json:"actions"`
	Continuations []ContinuationChat `json:"continuations"`
}

type Actions struct {
	AddChatItemAction AddChatItemAction `json:"addChatItemAction"`
}

type AddChatItemAction struct {
	Item Item `json:"item"`
}

type Item struct {
	LiveChatTextMessageRenderer LiveChatTextMessageRenderer `json:"liveChatTextMessageRenderer"`
}

type LiveChatTextMessageRenderer struct {
	Message    Message `json:"message"`
	AuthorName struct {
		SimpleText string `json:"simpleText"`
	}
	ContextMenuEndPoint ContextMenuEndPoint `json:"contextMenuEndPoint"`
}

type ContextMenuEndPoint struct {
	TimestampUsec int `json:"timestampUsec"`
}

type Message struct {
	Runs []Runs `json:"runs"`
}

type Runs struct {
	Text  string `json:"text,omitempty"`
	Emoji Emoji  `json:"emoji,omitempty"`
}

type Emoji struct {
	EmojiId string `json:"emojiId"`
}
type ChatMessagesResponse struct {
	ContinuationContents ContinuationContents `json:"continuationContents"`
}

var (
	LIVE_CHAT_URL = `https://www.youtube.com/youtubei/v1/live_chat/get_%s?key=%s`
)

const (
	API_TYPE              = "live_chat"
	YT_CFG_REGEX          = `ytcfg\.set\s*\(\s*({.+?})\s*\)\s*;`
	INITIAL_DATA_REGEX    = `(?:window\s*\[\s*["\']ytInitialData["\']\s*\]|ytInitialData)\s*=\s*({.+?})\s*;\s*(?:var\s+meta|</script|\n)`
	PLAYER_RESPONSE_REGEX = `ytInitialPlayerResponse\s*=\s*({.+?})\s*;\s*(?:var\s+meta|</script|\n)`
)

func regexSearch(regex string, str string) []string {
	r, _ := regexp.Compile(regex)
	matches := r.FindAllString(str, -1)
	return matches
}

func fetchInitialData(videoUrl string) (string, string, string) {
	resp, err := http.Get(videoUrl)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	defer resp.Body.Close()

	intArr, _ := ioutil.ReadAll(resp.Body)

	html := string(intArr)

	// TODO ::  work on regex and do not use trims
	initialDataArr := regexSearch(INITIAL_DATA_REGEX, html)
	initialData := strings.Trim(initialDataArr[0], "ytInitialData = ")
	initialData = strings.Trim(initialData, ";</script")
	playerResponse := regexSearch(PLAYER_RESPONSE_REGEX, html)[0]
	playerResponse = strings.Trim(playerResponse, "ytInitialPlayerResponse = ")
	playerResponse = strings.Trim(playerResponse, ";</sc")
	ytCfg := regexSearch(YT_CFG_REGEX, html)[0]
	ytCfg = strings.Trim(ytCfg, "ytcfg.set(")
	ytCfg = strings.Trim(ytCfg, ");")
	return initialData, playerResponse, ytCfg
}

func parseVideoData(initialDataMap map[string]interface{}) []SubMenuItems {
	_subMenuItems := initialDataMap["contents"].(map[string]interface{})["twoColumnWatchNextResults"].(map[string]interface{})["conversationBar"].(map[string]interface{})["liveChatRenderer"].(map[string]interface{})["header"].(map[string]interface{})["liveChatHeaderRenderer"].(map[string]interface{})["viewSelector"].(map[string]interface{})["sortFilterSubMenuRenderer"].(map[string]interface{})["subMenuItems"]
	_json, _ := json.Marshal(_subMenuItems)
	jsonString := string(_json)
	var subMenuItems []SubMenuItems
	json.Unmarshal([]byte(jsonString), &subMenuItems)
	return subMenuItems
}

func fetchChatMessages(initialContinuationInfo string, ytCfg YtCfg) {
	apiKey := ytCfg.INNERTUBE_API_KEY
	continuationUrl := fmt.Sprintf(LIVE_CHAT_URL, API_TYPE, apiKey)
	innertubeContext := ytCfg.INNERTUBE_CONTEXT
	// Now loop through all the chat messages
	for {
		context := Context{innertubeContext, initialContinuationInfo}
		b, _ := json.Marshal(context)
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
		var chatMsgResp ChatMessagesResponse
		json.Unmarshal([]byte(string(body)), &chatMsgResp)
		actions := chatMsgResp.ContinuationContents.LiveChatContinuation.Actions
		// iterate over actions
		for _, action := range actions {
			runs := action.AddChatItemAction.Item.LiveChatTextMessageRenderer.Message.Runs
			if len(runs) > 0 {
				authorName := action.AddChatItemAction.Item.LiveChatTextMessageRenderer.AuthorName.SimpleText
				// timeStampUSec := action.AddChatItemAction.Item.LiveChatTextMessageRenderer.ContextMenuEndPoint.TimestampUsec
				// timeStamp := time.Unix(int64(timeStampUSec), 0)
				// fmt.Print(timeStamp.String() + " | ")
				fmt.Print(authorName + ": ")
				for _, run := range runs {
					if run.Text != "" {
						fmt.Print(run.Text)
					} else {
						fmt.Print(run.Emoji.EmojiId)
					}
				}
				fmt.Println()
			}
		}
		// set new continuation and timeout
		initialContinuationInfo = chatMsgResp.ContinuationContents.LiveChatContinuation.Continuations[0].TimedContinuationData.Continuation
		timeOutMs := chatMsgResp.ContinuationContents.LiveChatContinuation.Continuations[0].TimedContinuationData.TimeoutMs
		if timeOutMs > 0 {
			time.Sleep(time.Duration(timeOutMs) * time.Millisecond)
		} else {
			time.Sleep(time.Second * 5)
		}
	}
}

func FetchLiveChat(videoUrl string) {
	initialData, playerResponse, _ytCfg := fetchInitialData(videoUrl)
	// parse the responses
	var ytCfg YtCfg
	json.Unmarshal([]byte(_ytCfg), &ytCfg)

	var playerResponseMap map[string]interface{}
	json.Unmarshal([]byte(playerResponse), &playerResponseMap)

	var initialDataMap map[string]interface{}
	json.Unmarshal([]byte(initialData), &initialDataMap)

	subMenuItems := parseVideoData(initialDataMap)
	initialContinuationInfo := subMenuItems[1].Continuation.ReloadContinuationData.Continuation

	fetchChatMessages(initialContinuationInfo, ytCfg)
}
