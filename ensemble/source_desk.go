package ensemble

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dacort/choirmaster/choir"
)

const deskComUrl = "https://%s.desk.com/api/v2"

type Desk struct {
	Url      string
	Username string
	Password string
	Choir    *choir.Choir

	LastUpdate time.Time

	Users map[int]string
}

type DeskConfig struct {
	Type string
	Key  string
	Http struct {
		Domain   string
		Username string
		Password string
	}
}

type DeskFeed struct {
	Total_Entries int
	Embedded      struct {
		Entries []DeskEntry
	} `json:"_embedded"`
}

type DeskEntry struct {
	Subject    string
	Status     string
	Created_At time.Time
	Updated_At time.Time
	Links      map[string]DeskLink `json:"_links"`
}

type DeskUser struct {
	Name string
}

type CaseHistory struct {
	Total_Entries int
	Embedded      struct {
		Entries []HistoryEntry
	} `json:"_embedded"`
	Links map[string]DeskLink `json:"_links"`
}

type HistoryEntry struct {
	Type       string
	Created_At time.Time
	Links      map[string]DeskLink `json:"_links"`
}

func (de *DeskEntry) Id() string {
	caseRef := de.Links["self"].Href
	parts := strings.Split(caseRef, "/")
	return parts[len(parts)-1]
}

func (de *DeskEntry) GetHistory(d *Desk) *CaseHistory {
	historyPath := fmt.Sprintf("/cases/%s/history?per_page=100", de.Id())
	full_history := new(CaseHistory)

	for {
		history := new(CaseHistory)

		// TODO: OK response
		d.GetUrl(historyPath, history)
		full_history.Total_Entries += history.Total_Entries
		full_history.Embedded.Entries = append(full_history.Embedded.Entries, history.Embedded.Entries...)

		if next_link, ok := history.Links["next"]; ok {
			historyPath = next_link.Href
			if historyPath == "" {
				break
			}
		} else {
			break
		}
	}

	return full_history
}

// TODO:
// Build a map of:
// User -> unique actions
// Actions to verbs
// Figure out a way to:
// Build a sentence from above map
// If there is no user, do we just say "The case was reopened"?
func (de *DeskEntry) BuildDescription(d *Desk, since_updated_at time.Time) string {
	descriptions := make([]string, 0)
	history_items := de.GetHistory(d)
	users := make([]string, 0)

	for _, item := range history_items.Embedded.Entries {
		if item.Created_At.Before(since_updated_at) || item.Type == "rule_applied" {
			continue
		}
		user := item.GetUserName(d)
		description_string := fmt.Sprintf("%s by %s", item.Type, user)
		descriptions = append(descriptions, description_string)
		if len(user) > 0 {
			users = append(users, user)
		}
	}

	user_actions := strings.Join(descriptions, "<br />")
	case_name := de.Subject
	return fmt.Sprintf("Case %s: \"%s\"<br />%s", de.Id(), case_name, user_actions)

	// if len(users) == 0 {
	// 	return fmt.Sprintf("Updated case %s: \"%s\"", de.Id(), case_name)
	// } else {
	// 	all_users := strings.Join(users, ", and ")
	// 	return fmt.Sprintf("%s updated case %s: \"%s\"", all_users, de.Id(), case_name)
	// }
}
func (he *HistoryEntry) UserId() (user_id int) {
	if invoker, ok := he.Links["invoker"]; ok {
		parts := strings.Split(invoker.Href, "/")
		user_id, _ = strconv.Atoi(parts[len(parts)-1])
		return user_id
	}

	return
}

func (he *HistoryEntry) GetUserName(d *Desk) (name string) {
	if name, ok := d.Users[he.UserId()]; ok {
		return name
	}

	userPath := fmt.Sprintf("/users/%d", he.UserId())
	user := new(DeskUser)

	// TODO: OK response
	d.GetUrl(userPath, user)

	d.Users[he.UserId()] = user.Name
	return user.Name
}

type DeskLink struct {
	Class string
	Href  string
}

func (d *Desk) Configure(config interface{}) {
	// This seems innane but it's the only way I can figure it out
	jsonString, _ := json.Marshal(config)
	var configObject DeskConfig
	if err := json.Unmarshal(jsonString, &configObject); err != nil {
		fmt.Println(err.Error())
		return
	}

	d.Url = fmt.Sprintf(deskComUrl, configObject.Http.Domain)
	d.Username = configObject.Http.Username
	d.Password = configObject.Http.Password
	d.Choir = &choir.Choir{configObject.Key}

	d.Users = make(map[int]string)

	fmt.Printf("Configured Desk: %s\n", configObject.Http.Domain)
}

func (d *Desk) Run(conductor chan *choir.Note) {
	for {
		last_update := d.LastUpdate
		feed := d.FetchUpdates()

		fmt.Println(feed.Total_Entries)

		for _, entry := range feed.Embedded.Entries {
			go func(e DeskEntry) {
				note := &choir.Note{
					Label: "Customer",
					Sound: "n/1",
					Text:  e.BuildDescription(d, last_update),
					Choir: d.Choir,
				}
				conductor <- note
			}(entry)
		}

		time.Sleep(60 * time.Second)
	}
}

func init() {
	fmt.Println("Registered Desk")
	RegisterService("desk", &Desk{LastUpdate: time.Now()})
}

// ####################################
// API IMPLEMENTATION
// ####################################
func (d *Desk) FetchUpdates() (feed DeskFeed) {
	feedUrl := fmt.Sprintf("%s/cases/search?since_updated_at=%d", d.Url, d.LastUpdate.Unix())
	fmt.Println(feedUrl)
	req, err := http.NewRequest("GET", feedUrl, nil)
	if err != nil {
		log.Printf("ERR building request: %s", err)
		return
	}

	req.SetBasicAuth(d.Username, d.Password)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("ERR making request: %s", err)
		return
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&feed)
	if err != nil {
		log.Printf("ERR decoding json from Desk: %s\n%s", err, resp.Body)
		return
	}

	d.LastUpdate = time.Now()

	return
}

// Generic API Getter
func (d *Desk) GetUrl(path string, decode_object interface{}) {
	url := fmt.Sprintf("%s%s", d.Url, path)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("ERR building request: %s", err)
		return
	}

	req.SetBasicAuth(d.Username, d.Password)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("ERR making request: %s", err)
		return
	}

	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&decode_object)
	if err != nil {
		log.Printf("ERR decoding json from Desk: %s\n%s", err, resp.Body)
		return
	}

	return
}
