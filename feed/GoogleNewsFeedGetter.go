package feed

import (
	"blogoconsumer/models"
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/streadway/amqp"
)

var newsRefreshInterval = 1 * time.Second
var lastUpdatedTime time.Time
var channelAmqp *amqp.Channel

func init() {
	amqpConnection, err := amqp.Dial(os.Getenv("RABBITMQ_URI"))
	if err != nil {
		log.Fatal(err)
	}
	lastUpdatedTime = time.Now()
	channelAmqp, _ = amqpConnection.Channel()
}

func Dispatch(url string) int {
	if strings.Contains(url, "google") {
		return 0
	}
	return -1
}

func GetFeedFromGoogleNews(url string) ([]models.Entry, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 ( Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.110 Safari/537.36")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	byteValue, _ := ioutil.ReadAll(resp.Body)
	var googleNewsFeed models.GoogleNewsFeed
	err = xml.Unmarshal(byteValue, &googleNewsFeed)
	if err != nil {
		return nil, err
	}

	newsEntries := make([]models.Entry, len(googleNewsFeed.Items))
	for ind, googleNewsEntry := range googleNewsFeed.Items {
		newsEntries[ind].Title = googleNewsEntry.Title
		newsEntries[ind].Link = googleNewsEntry.Link
		newsEntries[ind].Description = googleNewsEntry.Description
	}

	return newsEntries, err
}

func GetFeedEntries(url string) ([]models.Entry, error) {
	dispatch := Dispatch(url)

	var entries []models.Entry
	var err error
	if dispatch == 0 {
		entries, err = GetFeedFromGoogleNews(url)
	}

	return entries, err
}
