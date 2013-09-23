package ensemble

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dacort/choirmaster/choir"
)

const yammerActivityUrl = "https://www.yammer.com/api/v1/messages.json?access_token=%s&newer_than=%s"

type Yammer struct {
	Url         string
	AccessToken string
	Choir       *choir.Choir
	LastId      string
}

type YammerConfig struct {
	Type string
	Key  string
	Http struct {
		Access_Token string
	}
}

type YammerFeed struct {
	Messages   []YammerMessage
	References []YammerReference
}

type HammerTime struct {
	time.Time
}

func (t *HammerTime) UnmarshalJSON(data []byte) error {
	// Yammer format: "2013/09/20 18:40:57 +0000""
	const longForm = "2006/01/02 15:04:05 +0000"
	timestamp := strings.Replace(fmt.Sprintf("%s", data), "\"", "", -1)

	parsed, err := time.Parse(longForm, timestamp)
	if err != nil {
		fmt.Println("Could not unmarshall %s", timestamp)
		return err
	}
	t.Time = parsed
	return nil
}

type YammerMessage struct {
	Id              int
	Created_at      HammerTime
	Replied_To_Id   int
	Content_Excerpt string
	Body            struct {
		Rich  string
		Plain string
	}
	Sender_Id int
}

type YammerReference struct {
	Type      string
	Id        int
	Full_Name string
}

func (yf *YammerFeed) LookupUser(user_id int) (full_name string) {
	for _, item := range yf.References {
		if item.Type == "user" && item.Id == user_id {
			full_name = item.Full_Name
		}
	}

	if full_name == "" {
		full_name = "Unknown User"
	}

	return
}

func (ym *YammerMessage) SoundClass() string {
	if ym.Replied_To_Id == 0 {
		return "n/1"
	} else {
		return "n/0"
	}
}

func (ym *YammerMessage) GetCategory() string {
	if ym.Replied_To_Id == 0 {
		return "Yammer:update"
	} else {
		return "Yammer:reply"
	}
}

func (ym *YammerMessage) GetText() string {
	if len(ym.Body.Rich) < 500 {
		return ym.Body.Rich
	} else if len(ym.Content_Excerpt) > 500 {
		return ym.Content_Excerpt[0:500]
	} else {
		return ym.Content_Excerpt
	}
}

func (y *Yammer) FetchUpdates() (feed YammerFeed) {
	url := fmt.Sprintf(yammerActivityUrl, y.AccessToken, y.LastId)
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("ERR making request for %s: %s", url, err)
		return
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&feed)
	if err != nil {
		log.Printf("ERR decoding json from Yammer: %s\n%s", err, resp.Body)
		return
	}

	if len(feed.Messages) > 0 {
		y.LastId = fmt.Sprintf("%d", feed.Messages[0].Id)
	}

	return
}

func (y *Yammer) Configure(config interface{}) {
	// This seems innane but it's the only way I can figure it out
	jsonString, _ := json.Marshal(config)
	var configObject YammerConfig
	if err := json.Unmarshal(jsonString, &configObject); err != nil {
		fmt.Println(err.Error())
		return
	}
	y.Url = fmt.Sprintf(yammerActivityUrl, configObject.Http.Access_Token, "1")
	y.AccessToken = configObject.Http.Access_Token
	y.Choir = &choir.Choir{configObject.Key}

	// Prime the LastId
	y.LastId = "1"
	_ = y.FetchUpdates()

	fmt.Println("Configured Yammer")
}

func (y *Yammer) Run(conductor chan *choir.Note) {
	for {
		feed := y.FetchUpdates()

		for _, message := range feed.Messages {
			note := &choir.Note{
				Label: message.GetCategory(),
				Sound: message.SoundClass(),
				Text:  fmt.Sprintf("%s: %s", feed.LookupUser(message.Sender_Id), message.GetText()),
				Choir: y.Choir,
			}

			go func() {
				conductor <- note
			}()
		}

		time.Sleep(60 * time.Second)
	}
}

func init() {
	fmt.Println("Registered Yammer")
	RegisterService("yammer", &Yammer{})
}
