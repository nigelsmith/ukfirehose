package main

import (
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	T_CONSUMER_K       = "TWITTER_CONSUMER_KEY"
	T_CONSUMER_SEC_K   = "TWITTER_CONSUMER_SEC_KEY"
	T_ACCESS_TOKEN     = "TWITTER_ACCESS_TOKEN"
	T_ACCESS_TOKEN_SEC = "TWITTER_ACCESS_TOKEN_SECRET"
)

func main() {
	consumerKey := os.Getenv(T_CONSUMER_K)
	consumerSecretKey := os.Getenv(T_CONSUMER_SEC_K)
	accessToken := os.Getenv(T_ACCESS_TOKEN)
	accessTokenSecret := os.Getenv(T_ACCESS_TOKEN_SEC)

	config := oauth1.NewConfig(consumerKey, consumerSecretKey)
	token := oauth1.NewToken(accessToken, accessTokenSecret)
	httpClient := config.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)

	countC := make(chan struct{})

	demux := twitter.NewSwitchDemux()
	demux.Tweet = func(t *twitter.Tweet) {
		countC <- struct{}{}
	}

	demux.Warning = func(w *twitter.StallWarning) {
		log.Printf("Stalling: %s, %s, %d%% full\n", w.Code, w.Message, w.PercentFull)
	}


	params := &twitter.StreamFilterParams{
		// Crude uk bounding box
		Locations:     []string{"-5.948959,49.939341,0.950455,58.735828",},
		StallWarnings: twitter.Bool(true),
	}

	stream, err := client.Streams.Filter(params)
	if err != nil {
		log.Fatal(err.Error())
	}

	go demux.HandleChan(stream.Messages)

	go func() {
		var counter int
		t := time.NewTicker(3 * time.Second)

		for {
			select {
			case <-countC:
				counter++
			case <-t.C:
				log.Printf("Saw %d tweets.\n", counter)
			}
		}
	}()

	stopC := make(chan os.Signal)
	signal.Notify(stopC, syscall.SIGINT, syscall.SIGTERM)
	<-stopC

	log.Println("Stopping...")
	stream.Stop()
}
