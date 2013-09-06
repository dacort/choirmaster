package ensemble

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dacort/choirmaster/choir"
)

type Jira struct {
	Url        string
	Username   string
	Password   string
	Choir      *choir.Choir
	LastUpdate time.Time
}

type JiraConfig struct {
	Type string
	Key  string
	Http struct {
		Domain   string
		Username string
		Password string
	}
}

// XML Activity Feed
type JiraFeed struct {
	XMLName        xml.Name    `xml:"feed"`
	Updated        string      `xml:"updated"`
	TimezoneOffset string      `xml:"timezone-offset"`
	Entry          []JiraEntry `xml:"entry"`
}

type JiraEntry struct {
	Published time.Time `xml:"published"`
	Category  struct {
		Term string `xml:"term,attr"`
	} `xml:"category"`
	AuthorName string `xml:"author>name"`
	Title      string `xml:"title"`
}

func (j *Jira) Configure(config interface{}) {
	// This seems innane but it's the only way I can figure it out
	jsonString, _ := json.Marshal(config)
	var configObject JiraConfig
	if err := json.Unmarshal(jsonString, &configObject); err != nil {
		fmt.Println(err.Error())
		return
	}

	j.Url = fmt.Sprintf("https://%s/activity?maxResults=20&os_authType=basic&title=undefined", configObject.Http.Domain)
	j.Username = fmt.Sprintf("%s", configObject.Http.Username)
	j.Password = fmt.Sprintf("%s", configObject.Http.Password)
	j.Choir = &choir.Choir{configObject.Key}

	fmt.Printf("Configured JIRA: %s\n", configObject.Http.Domain)
}

func (j *Jira) FetchUpdates() (feed JiraFeed) {
	req, err := http.NewRequest("GET", j.Url, nil)
	if err != nil {
		log.Printf("ERR building request: %s", err)
	}

	req.SetBasicAuth(j.Username, j.Password)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("ERR making request: %s", err)
	}

	dec := xml.NewDecoder(resp.Body)
	err = dec.Decode(&feed)
	if err != nil {
		log.Printf("ERR decoding json from JIRA: %s\n%s", err, resp.Body)
	}

	resp.Body.Close()

	return
}

func SoundClass(category string) string {
	switch category {
	case "comment":
		return "n/1"
	case "resolved":
		return "g/1"
	case "closed":
		return "g/1"
	case "started":
		return "g/1"
	case "created":
		return "b/1"
	case "reopened":
		return "b/2"
	default:
		return "n/0"
	}
}

func (j *Jira) Run(conductor chan *choir.Note) {
	for {
		feed := j.FetchUpdates()

		for _, entry := range feed.Entry {

			if entry.Published.Before(j.LastUpdate) {
				continue
			} else {
				j.LastUpdate = entry.Published.Add(1 * time.Second)
			}

			if entry.Category.Term == "" {
				entry.Category.Term = "changed"
			}

			note := &choir.Note{
				Label: fmt.Sprintf("JIRA:%s", entry.Category.Term),
				Sound: SoundClass(entry.Category.Term),
				Text:  entry.Title,
				Choir: j.Choir,
			}

			go func() {
				conductor <- note
			}()
		}

		time.Sleep(5 * time.Second)
	}
}

func init() {
	fmt.Println("Registered JIRA")
	RegisterService("jira", &Jira{LastUpdate: time.Now()})
}
