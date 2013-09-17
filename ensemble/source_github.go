package ensemble

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dacort/choirmaster/choir"
)

type Github struct {
	Url        string
	LastUpdate time.Time
	Choir      *choir.Choir
}

type GithubConfig struct {
	Type string
	Key  string
	Http struct {
		Username     string
		Orgname      string
		Access_Token string
	}
}

// XML Activity Feed
type GithubFeed struct {
	XMLName xml.Name      `xml:"feed"`
	Updated string        `xml:"updated"`
	Entry   []GithubEntry `xml:"entry"`
}

type GithubEntry struct {
	Published  time.Time `xml:"published"`
	Id         string    `xml:"id"`
	AuthorName string    `xml:"author>name"`
	Title      string    `xml:"title"`
	Content    string    `xml:"content"`
}

func (je *GithubEntry) GetCategory() string {
	firstSplit := strings.Split(je.Id, ":")
	tagDirty := firstSplit[len(firstSplit)-1]

	tag := strings.Split(tagDirty, "/")[0]

	return fmt.Sprintf("GitHub:%s", tag)
}

func (g *Github) Configure(config interface{}) {
	jsonString, _ := json.Marshal(config)
	var configObject GithubConfig
	if err := json.Unmarshal(jsonString, &configObject); err != nil {
		fmt.Println(err.Error())
		return
	}

	// We could use the API, but the feed gives us pretty titles
	// https://github.com/organizations/%s/%s.private.atom?token=%s
	// https://api.github.com/users/%s/events/orgs/%s?access_token=%s
	g.Url = fmt.Sprintf("https://github.com/organizations/%s/%s.private.atom?token=%s",
		configObject.Http.Orgname,
		configObject.Http.Username,
		configObject.Http.Access_Token,
	)
	g.Choir = &choir.Choir{configObject.Key}

	fmt.Printf("Configured Github: %s\n", configObject.Http.Orgname)
}

func (g *Github) FetchUpdates() (feed GithubFeed) {
	resp, err := http.Get(g.Url)
	if err != nil {
		log.Printf("ERR making request for %s: %s", g.Url, err)
		return
	}

	dec := xml.NewDecoder(resp.Body)
	err = dec.Decode(&feed)
	if err != nil {
		log.Printf("ERR decoding xml from GitHub: %s\n%s", err, resp.Body)
		return
	}

	resp.Body.Close()

	return
}

func (ge *GithubEntry) SoundClass() string {
	switch ge.GetCategory() {
	case "PublicEvent":
		return "g/3"
	case "TeamAddEvent":
		return "g/3"
	case "PullRequestEvent":
		return "n/2"
	default:
		return "n/0"
	}
}

func (g *Github) Run(conductor chan *choir.Note) {
	for {
		feed := g.FetchUpdates()

		for _, entry := range feed.Entry {
			if entry.Published.Before(g.LastUpdate) {
				continue
			} else {
				g.LastUpdate = entry.Published.Add(1 * time.Second)
			}

			note := &choir.Note{
				Label: entry.GetCategory(),
				Sound: entry.SoundClass(),
				Text:  entry.Title,
				Choir: g.Choir,
			}

			// fmt.Println(note)
			go func() {
				conductor <- note
			}()

		}

		time.Sleep(5 * time.Second)
	}
}

func init() {
	fmt.Println("Registered GitHub")
	RegisterService("github", &Github{LastUpdate: time.Now()})
}
