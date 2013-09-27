package ensemble

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dacort/choirmaster/choir"
)

type Campfire struct {
	Url     string
	Orgname string
	Token   string
	Choir   *choir.Choir

	Rooms map[int]string
	Users map[int]string
}

type Room struct {
	Id            int
	Name          string
	LastMessageId int
}

type RoomResponse struct {
	Room Room
}
type UserResponse struct {
	User struct {
		Name string
	}
}

type CampfireConfig struct {
	Type    string
	Key     string
	Rooms   []int
	Token   string
	Orgname string
}

type CampfireMessage struct {
	Room_Id   int
	CreatedAt time.Time
	Body      string
	Id        int
	User_Id   int
	Type      string
	Starred   bool
}

func (c *Campfire) GetUser(id int) string {
	if name, ok := c.Users[id]; ok {
		return name
	}

	log.Printf("Looking up user %d", id)
	userUrl := fmt.Sprintf("https://%s.campfirenow.com/users/%d.json", c.Orgname, id)
	req, err := http.NewRequest("GET", userUrl, nil)
	if err != nil {
		log.Fatalf("error building request: %s", err)
	}

	req.SetBasicAuth(c.Token, "x")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("error making request: %s", err)
	}

	var user UserResponse
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&user)
	if err != nil {
		return fmt.Sprintf("%d", id)
	}
	resp.Body.Close()

	c.Users[id] = user.User.Name
	return user.User.Name
}

func (c *Campfire) GetRoom(id int) string {
	if name, ok := c.Rooms[id]; ok {
		return name
	}

	log.Printf("Looking up room %d", id)
	roomUrl := fmt.Sprintf("https://%s.campfirenow.com/room/%d.json", c.Orgname, id)
	req, err := http.NewRequest("GET", roomUrl, nil)
	if err != nil {
		log.Fatalf("error building request: %s", err)
	}

	req.SetBasicAuth(c.Token, "x")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("error making request: %s", err)
	}

	var room RoomResponse
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&room)
	if err != nil {
		return fmt.Sprintf("%d", id)
	}
	resp.Body.Close()

	c.Rooms[id] = room.Room.Name
	return room.Room.Name
}

func (c *Campfire) Configure(config interface{}) {
	jsonString, _ := json.Marshal(config)
	var configObject CampfireConfig
	if err := json.Unmarshal(jsonString, &configObject); err != nil {
		fmt.Println(err.Error())
		return
	}

	c.Url = fmt.Sprintf("https://streaming.campfirenow.com/room/%d/live.json", configObject.Rooms[0])
	c.Token = configObject.Token
	c.Orgname = configObject.Orgname

	c.Rooms = make(map[int]string)
	c.Users = make(map[int]string)

	fmt.Printf("Configured Campfire: %d\n", configObject.Rooms[0])
}

func (c *Campfire) Run(conductor chan *choir.Note) {
	req, err := http.NewRequest("GET", c.Url, nil)
	if err != nil {
		log.Fatalf("error building request: %s", err)
	}

	req.SetBasicAuth(c.Token, "x")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("error making campfire request: %s", err)
	}
	defer resp.Body.Close()

	var reader *bufio.Reader
	reader = bufio.NewReader(resp.Body)

	for {
		line, err := reader.ReadBytes('}')
		if err != nil {
			fmt.Println("campfire connection broken :(", err)
			break
		}

		line = bytes.TrimSpace(line)

		if len(line) == 0 {
			continue
		}

		message := new(CampfireMessage)
		json.Unmarshal(line, &message)

		if message.Body == "" {
			continue
		}

		// TODO Pull all rooms, fire off every minute with the number of people talking in each room
		fmt.Printf("(%s) %s: %s\n", c.GetRoom(message.Room_Id), c.GetUser(message.User_Id), message.Body)
	}
}

func init() {
	fmt.Println("Registered Campfire")
	RegisterService("campfire", &Campfire{})
}
