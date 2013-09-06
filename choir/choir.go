package choir

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
)

const choirSingUrl = "http://api.choir.io/%s"

type Choir struct {
	Key string
}

type Note struct {
	Label string
	Sound string
	Text  string
	Choir *Choir
}

func NewChoir(key string) *Choir {
	return &Choir{Key: key}
}

func (c *Choir) PostUrl() (link string) {
	return fmt.Sprintf(choirSingUrl, c.Key)
}

func (c *Choir) Sing(note Note) {

	resp, err := http.PostForm(c.PostUrl(), url.Values{
		"label": {note.Label},
		"sound": {note.Sound},
		"text":  {note.Text},
	})
	defer resp.Body.Close()

	if err != nil {
		log.Printf("ERR making choir request: %s", err)
		log.Print(note)
	}
}
