package YtChat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"
	"strconv"
)

type SubMenuItems struct {
	Title        string
	Continuation struct {
		ReloadContinuationData struct {
			Continuation string
		}
	}
}

type InnerTubeContext struct {
	Client struct {
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
		ConfigInfo       struct {
			AppInstallData string `json:"appInstallData"`
		} `json:"configInfo"`
	} `json:"client"`
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

type ContinuationChat struct {
	TimedContinuationData struct {
		Continuation string `json:"continuation"`
		TimeoutMs    int    `json:"timeoutMs"`
	} `json:"timedContinuationData"`
	InvalidationContinuationData struct {
		Continuation string `json:"continuation"`
		TimeoutMs    int    `json:"timeoutMs"`
	} `json:"invalidationContinuationData"`
}

type Actions struct {
	AddChatItemAction struct {
		Item struct {
			LiveChatTextMessageRenderer struct {
				Message struct {
					Runs []Runs `json:"runs"`
				} `json:"message"`
				AuthorName struct {
					SimpleText string `json:"simpleText"`
				}
				TimestampUsec string `json:"timestampUsec"`
			} `json:"liveChatTextMessageRenderer"`
		} `json:"item"`
	} `json:"addChatItemAction"`
}

type Runs struct {
	Text  string `json:"text,omitempty"`
	Emoji struct {
		EmojiId string `json:"emojiId"`
		IsCustomEmoji bool `json:"isCustomEmoji,omitempty"`
		Image struct {
			Thumbnails []struct {
				Url string `json:"url,omitempty"`
			}
		}
	}  `json:"emoji,omitempty"`
}

type ChatMessagesResponse struct {
	ContinuationContents struct {
		LiveChatContinuation struct {
			Actions       []Actions          `json:"actions"`
			Continuations []ContinuationChat `json:"continuations"`
		} `json:"liveChatContinuation"`
	} `json:"continuationContents"`
}

type ChatMessage struct {
	AuthorName string
	Message    string
	Timestamp time.Time
}

var (
	LIVE_CHAT_URL = `https://www.youtube.com/youtubei/v1/live_chat/get_%s?key=%s`
)

const (
	API_TYPE           = "live_chat"
	YT_CFG_REGEX       = `ytcfg\.set\s*\(\s*({.+?})\s*\)\s*;`
	INITIAL_DATA_REGEX = `(?:window\s*\[\s*["\']ytInitialData["\']\s*\]|ytInitialData)\s*=\s*({.+?})\s*;\s*(?:var\s+meta|</script|\n)`
)

func regexSearch(regex string, str string) []string {
	r, _ := regexp.Compile(regex)
	matches := r.FindAllString(str, -1)
	return matches
}

func parseVideoData(initialDataMap map[string]interface{}) []SubMenuItems {
	_subMenuItems := initialDataMap["contents"].(map[string]interface{})["twoColumnWatchNextResults"].(map[string]interface{})["conversationBar"].(map[string]interface{})["liveChatRenderer"].(map[string]interface{})["header"].(map[string]interface{})["liveChatHeaderRenderer"].(map[string]interface{})["viewSelector"].(map[string]interface{})["sortFilterSubMenuRenderer"].(map[string]interface{})["subMenuItems"]
	_json, _ := json.Marshal(_subMenuItems)
	jsonString := string(_json)
	var subMenuItems []SubMenuItems
	json.Unmarshal([]byte(jsonString), &subMenuItems)
	return subMenuItems
}

func parseMicoSeconds(timeStampStr string) time.Time {
	tm, _ := strconv.ParseInt(timeStampStr, 10, 64)
	tm = tm / 1000
    sec := tm / 1000
    msec := tm % 1000
    return time.Unix(sec, msec*int64(time.Millisecond))
}

func fetchChatMessages(initialContinuationInfo string, ytCfg YtCfg) ([]ChatMessage, string, int) {
	apiKey := ytCfg.INNERTUBE_API_KEY
	continuationUrl := fmt.Sprintf(LIVE_CHAT_URL, API_TYPE, apiKey)
	innertubeContext := ytCfg.INNERTUBE_CONTEXT

	context := Context{innertubeContext, initialContinuationInfo}
	b, _ := json.Marshal(context)
	var jsonData = []byte(b)
	request, error := http.NewRequest("POST", continuationUrl, bytes.NewBuffer(jsonData))
	if error != nil {
		panic(error)
	}
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	response, error := client.Do(request)
	if error != nil {
		panic(error)
	}
	if response.StatusCode != 200 {
		panic(fmt.Sprintf("Error fetching chat messages. Status code: %d", response.StatusCode))
	}
	body, _ := ioutil.ReadAll(response.Body)
	response.Body.Close()
	var chatMsgResp ChatMessagesResponse
	json.Unmarshal([]byte(string(body)), &chatMsgResp)
	actions := chatMsgResp.ContinuationContents.LiveChatContinuation.Actions
	chatMessages := []ChatMessage{}
	for _, action := range actions {
		liveChatTextMessageRenderer := action.AddChatItemAction.Item.LiveChatTextMessageRenderer
		runs := liveChatTextMessageRenderer.Message.Runs
		if len(runs) > 0 {
			chatMessage := ChatMessage{}
			authorName := liveChatTextMessageRenderer.AuthorName.SimpleText
			chatMessage.Timestamp = parseMicoSeconds(liveChatTextMessageRenderer.TimestampUsec)
			chatMessage.AuthorName = authorName
			text := ""
			for _, run := range runs {
				if run.Text != "" {
					text += run.Text
				} else {
					if run.Emoji.IsCustomEmoji {
						numberOfThumbnails := len(run.Emoji.Image.Thumbnails)
						// Adding some whitespace after custom image URLs
						// without the whitespace it would be difficult to parse these URLs
						if numberOfThumbnails > 0 && numberOfThumbnails == 2 {
							text += run.Emoji.Image.Thumbnails[1].Url + " "
						} else if numberOfThumbnails == 1 {
							text += run.Emoji.Image.Thumbnails[0].Url + " "
						}
					} else {
						text += run.Emoji.EmojiId
					}
				}
			}
			chatMessage.Message = text
			chatMessages = append(chatMessages, chatMessage)
		}
	}
	// get new continuation and timeout
	timeoutMs := 5
	continuations := chatMsgResp.ContinuationContents.LiveChatContinuation.Continuations[0]
	if continuations.TimedContinuationData.Continuation == "" {
		initialContinuationInfo = continuations.InvalidationContinuationData.Continuation
		timeoutMs = continuations.InvalidationContinuationData.TimeoutMs
	} else {
		initialContinuationInfo = continuations.TimedContinuationData.Continuation
		timeoutMs = continuations.TimedContinuationData.TimeoutMs
	}
	return chatMessages, initialContinuationInfo, timeoutMs
}

func ParseInitialData(videoUrl string) (string, YtCfg) {
	resp, err := http.Get(videoUrl)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	intArr, _ := ioutil.ReadAll(resp.Body)

	html := string(intArr)

	// TODO ::  work on regex and do not use trims
	initialDataArr := regexSearch(INITIAL_DATA_REGEX, html)
	initialData := strings.Trim(initialDataArr[0], "ytInitialData = ")
	initialData = strings.Trim(initialData, ";</script")
	ytCfg := regexSearch(YT_CFG_REGEX, html)[0]
	ytCfg = strings.Trim(ytCfg, "ytcfg.set(")
	ytCfg = strings.Trim(ytCfg, ");")

	var _ytCfg YtCfg
	json.Unmarshal([]byte(ytCfg), &_ytCfg)

	var initialDataMap map[string]interface{}
	json.Unmarshal([]byte(initialData), &initialDataMap)

	subMenuItems := parseVideoData(initialDataMap)
	initialContinuationInfo := subMenuItems[1].Continuation.ReloadContinuationData.Continuation
	return initialContinuationInfo, _ytCfg
}

func FetchContinuationChat(continuation string, ytCfg YtCfg) ([]ChatMessage, string) {
	chatMessages, continuation, timeoutMs := fetchChatMessages(continuation, ytCfg)
	// Sleep for timeoutMs milliseconds sent by the server
	if timeoutMs > 0 {
		time.Sleep(time.Duration(timeoutMs) * time.Millisecond)
	} else {
		time.Sleep(time.Second * 5)
	}
	return chatMessages, continuation
}
