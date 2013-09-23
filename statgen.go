package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

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
	config := getConfig("config.json")

	since := time.Now().Add(-(48 + 14) * time.Hour)
	var wg sync.WaitGroup

	// Configure each service
	for _, source := range config.Sources {
		source_type := fmt.Sprintf("%s", source["type"])
		if service, ok := ensemble.FindService(source_type); ok {
			service.Configure(source)

			wg.Add(1)
			go func() {
				// Decrement the counter when the goroutine completes.
				defer wg.Done()

				// Print out the updates
				service.PrintUpdatesSince(since)
			}()
		}
	}

	// Wait for all HTTP fetches to complete.
	wg.Wait()
}
