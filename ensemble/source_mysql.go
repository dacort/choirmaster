package ensemble

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"database/sql"
	_ "github.com/go-sql-driver/mysql"

	"github.com/dacort/choirmaster/choir"
)

type Mysql struct {
	DatabaseConnection *sql.DB
	DateQuery          string

	Choir      *choir.Choir
	LastUpdate time.Time
}

type MysqlConfig struct {
	Key      string
	Database string
	Username string
	Password string
	Query    string
}

func (m *Mysql) Configure(config interface{}) {
	// This seems innane but it's the only way I can figure it out
	jsonString, _ := json.Marshal(config)
	var configObject MysqlConfig
	if err := json.Unmarshal(jsonString, &configObject); err != nil {
		fmt.Println(err.Error())
		return
	}

	m.DateQuery = configObject.Query
	m.Choir = &choir.Choir{configObject.Key}
	connectStr := fmt.Sprintf("%s:%s@%s", configObject.Username, configObject.Password, configObject.Database)

	db, err := sql.Open("mysql", connectStr)
	if err != nil {
		fmt.Print("Could not connecto to MySQL", err.Error())
		return
	}
	// defer db.Close()

	m.DatabaseConnection = db

	fmt.Printf("Configured Mysql: %s\n", configObject.Database)
}

func (m *Mysql) Run(conductor chan *choir.Note) {
	for {
		var (
			name  string
			price float64
		)
		rows, err := m.DatabaseConnection.Query(m.DateQuery, m.LastUpdate.UTC())
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(&name, &price)
			if err != nil {
				log.Fatal(err)
			}
			go func(n string, p float64) {
				note := &choir.Note{
					Label: "SALE",
					Sound: "bloop/cheer",
					Text:  fmt.Sprintf("Just sold %s at $%v!", n, p),
					Choir: m.Choir,
				}
				conductor <- note
			}(name, price)
		}
		err = rows.Err()
		if err != nil {
			log.Fatal(err)
		}

		m.LastUpdate = time.Now()

		time.Sleep(60 * time.Second)
	}
}

func init() {
	fmt.Println("Registered MySQL")
	RegisterService("mysql", &Mysql{LastUpdate: time.Now()})
}
