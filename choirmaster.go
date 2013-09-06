package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/dacort/choirmaster/choir"
	"github.com/dacort/choirmaster/ensemble"
)

type Config struct {
	Sources []map[string]interface{}
}

func getConfig(filename string) *Config {
	config := new(Config)
	file, e := ioutil.ReadFile(filename)
	if e != nil {
		fmt.Printf("File error: %v\n", e)
		os.Exit(1)
	}

	json.Unmarshal(file, &config)

	return config
}

func main() {
	// All services should be registered at this point.
	// Read in the config file and then FindService(type) and Configure(config_item)
	config := getConfig("config.json")

	// Now, create a channel to listen on.
	// Each time a service gets updated, this channel gets called
	var conductorChan = make(chan *choir.Note)

	// Configure each service
	for _, source := range config.Sources {
		source_type := fmt.Sprintf("%s", source["type"])
		if service, ok := ensemble.FindService(source_type); ok {
			service.Configure(source)
			go service.Run(conductorChan)
		}
	}

	// Let's make this sucker sing!
	for {
		b := <-conductorChan
		go b.Choir.Sing(*b)
	}

}
