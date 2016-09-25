package main

import (
	"fmt"
	"os"
	"encoding/json"
	"log"
	"os/signal"
	"syscall"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

type Channels struct {
Title string `json:"title"`
Filters []string `json:"filters"`
}


type Config struct {
	Consumer struct {
		Key string `json:"key"`
		Secret string `json:"secret"`
	} `json:"consumer"`
	Access struct {
		Token string `json:"token"`
		Secret string `json:"secret"`
	} `json:"access"`
	// Represents the filters used for the stream
	Channels []Channels `json:"channels"`
}


// Find the config file and build a File instance with it.
func get_config_file() *os.File {
	//For now. When we will be in the docker, the config file should be looked at
	///opt/app_name/credentials.json
	//file, file_err := os.Open("/opt/twitter-sentiment-elk-go/credentials.json")
	file, file_err := os.Open("credentials.json")
	if file_err != nil {
		panic(file_err)
	}
	return file
}

//Get an instance of Credentials to connect to the twitter app.
func get_config() (*Config, error) {
	c := Config{}
	decoder := json.NewDecoder(get_config_file())
	err := decoder.Decode(&c)
	return &c, err
}

// Build a twitter client with the credentials found in the config file.
func get_twitter_client(config *Config) *twitter.Client {
	twitter_config := oauth1.NewConfig(config.Consumer.Key, config.Consumer.Secret)
	twitter_token := oauth1.NewToken(config.Access.Token, config.Access.Secret)
	httpClient := twitter_config.Client(oauth1.NoContext, twitter_token)
	client := twitter.NewClient(httpClient)
	return client
}

// Init a twitter client and start a stream. The filters of the stream are found
// in the config.
func run_stream(client *twitter.Client, channel *Channels) {
	fmt.Printf("Starting Stream... %v", channel.Title)

	// FILTER
	filterParams := &twitter.StreamFilterParams{
		Track:         channel.Filters,
		StallWarnings: twitter.Bool(true),
	}
	stream, err := client.Streams.Filter(filterParams)
	if err != nil {
		log.Fatal(err)
	}

	// Convenience Demux demultiplexed stream messages
	demux := twitter.NewSwitchDemux()
	demux.Tweet = func(tweet *twitter.Tweet) {
		fmt.Println(tweet.Text)
	}
	demux.DM = func(dm *twitter.DirectMessage) {
		fmt.Println(dm.SenderID)
	}
	demux.Event = func(event *twitter.Event) {
		fmt.Printf("%#v\n", event)
	}
	go demux.HandleChan(stream.Messages)

	// Wait for SIGINT and SIGTERM (HIT CTRL-C)
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Println(<-ch)

	fmt.Println("Stopping Stream...")
	stream.Stop()
}

func main() {
	config, _ := get_config()
	client := get_twitter_client(config)
	run_stream(client, "cat")
	//for _, streamChannel := range config.StreamChannels{
	//}
}
