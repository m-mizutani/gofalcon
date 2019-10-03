package gofalcon

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// SensorAPI is for /sensor/ prefix APIs
type SensorAPI struct {
	client *Client
}

// EntitiesDatafeedInput is argument of EntitiesDatafeed
type EntitiesDatafeedInput struct {
	AppID  *string
	Format *string
}

// EntitiesDatafeedOutput is a result of EntitiesDatafeed
type EntitiesDatafeedOutput struct {
	BaseResponse
	Resources []DataFeedResource `json:"resources"`
}

type DataFeedSessionToken struct {
	Expiration string `json:"expiration"`
	Token      string `json:"token"`
}

type DataFeedResource struct {
	DataFeedURL                  string `json:"dataFeedURL"`
	RefreshActiveSessionInterval int    `json:"refreshActiveSessionInterval"`
	RefreshActiveSessionURL      string `json:"refreshActiveSessionURL"`
	SessionToken                 DataFeedSessionToken
}

// Partition extracts parition number from DataFeedURL
func (x DataFeedResource) Partition() (int, error) {
	urlArr := strings.Split(strings.Split(x.DataFeedURL, "?")[0], "/")
	partition, err := strconv.Atoi(urlArr[len(urlArr)-1])
	if err != nil {
		return 0, errors.Wrapf(err, "Fail to extract Parition from URL: %s", x.DataFeedURL)
	}
	return partition, nil
}

// EntitiesDatafeed retrieves URL of event stream.
func (x *SensorAPI) EntitiesDatafeed(input *EntitiesDatafeedInput) (*EntitiesDatafeedOutput, error) {
	qs := url.Values{}
	if input.AppID != nil {
		qs.Add("appId", *input.AppID)
	}
	if input.Format != nil {
		qs.Add("format", *input.Format)
	}

	req := request{
		Method:      "GET",
		Path:        "sensors/entities/datafeed/v2",
		QueryString: qs,
	}

	var output EntitiesDatafeedOutput
	if err := x.client.sendRequest(req, &output); err != nil {
		return nil, errors.Wrap(err, "Fail to query detections")
	}

	Logger.WithFields(logrus.Fields{
		"appId":  StringValue(input.AppID),
		"format": StringValue(input.Format),
		"meta":   output.Meta,
	}).Debug("Done SensorAPI.EntitiesDatafeed")

	return &output, nil
}

// EntitiesDatafeedActionInput is argument of EntitiesDatafeedAction
type EntitiesDatafeedActionInput struct {
	ActionName *string
	AppID      *string
	Partition  *int
}

// EntitiesDatafeedActionOutput is a result of EntitiesDatafeedAction
type EntitiesDatafeedActionOutput struct {
	BaseResponse
}

// EntitiesDatafeedAction retrieves URL of event stream.
func (x *SensorAPI) EntitiesDatafeedAction(input *EntitiesDatafeedActionInput) (*EntitiesDatafeedActionOutput, error) {
	qs := url.Values{}
	if input.AppID == nil {
		return nil, fmt.Errorf("Input AppID is required")
	}
	if input.ActionName == nil {
		return nil, fmt.Errorf("Input AtionName is required")
	}
	if input.Partition == nil {
		return nil, fmt.Errorf("Input Partition is required")
	}

	qs.Add("appId", *input.AppID)
	qs.Add("action_name", *input.ActionName)

	req := request{
		Method:      "POST",
		Path:        fmt.Sprintf("sensors/entities/datafeed-actions/v1/%d", *input.Partition),
		QueryString: qs,
		Headers:     []httpHeader{{"Content-Type", "application/json"}},
	}

	var output EntitiesDatafeedActionOutput
	if err := x.client.sendRequest(req, &output); err != nil {
		return nil, errors.Wrapf(err, "Fail to %s DataFeed", *input.ActionName)
	}

	Logger.WithFields(logrus.Fields{
		"appId":       StringValue(input.AppID),
		"action_name": StringValue(input.ActionName),
		"partition":   IntValue(input.Partition),
		"meta":        output.Meta,
	}).Debug("Done SensorAPI.EntitiesDatafeedAction")

	return &output, nil
}

// --------------------------------------------

// StreamEventMetaData is metadata of event stream from Falcon API.
type StreamEventMetaData struct {
	CustomerIDString  string `json:"customerIDString"`
	EventType         string `json:"eventType"`
	Offset            int    `json:"offset"`
	EventCreationTime int64  `json:"eventCreationTime"`
}

type streamEvent struct {
	Meta  StreamEventMetaData    `json:"metadata"`
	Event map[string]interface{} `json:"event"`
}

// StreamQueue is issued from EventStream() including metadata, event and error.
// If error is occurred, Meta and Event must be nil.
type StreamQueue struct {
	Error error
	Meta  *StreamEventMetaData
	Event map[string]interface{}
}

const (
	// StreamEventQueueSize is default queue size
	StreamEventQueueSize = 1024
)

func readEventStreamFeed(feed DataFeedResource) chan *StreamQueue {
	ch := make(chan *StreamQueue, 128)
	go func() {
		defer close(ch)
		url := feed.DataFeedURL
		client := http.Client{}

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			ch <- &StreamQueue{Error: errors.Wrap(err, "fail to create a graylog http request")}
			return
		}

		req.Header.Add("Authorization", "Token "+feed.SessionToken.Token)
		req.Header.Add("Accept", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			ch <- &StreamQueue{Error: errors.Wrap(err, "fail to send request to server")}
			return
		}

		Logger.WithFields(logrus.Fields{
			"url":  feed.DataFeedURL,
			"code": resp.StatusCode,
		}).Info("Opened DataFeedURL")

		defer resp.Body.Close()
		decoder := json.NewDecoder(resp.Body)

		for {
			ev := new(streamEvent)
			if err := decoder.Decode(&ev); err == io.EOF {
				break
			} else if err != nil {
				ch <- &StreamQueue{Error: errors.Wrap(err, "fail to unmarshal event stream")}
				return
			}

			q := new(StreamQueue)
			q.Meta = &ev.Meta
			q.Event = ev.Event
			ch <- q
		}
	}()

	return ch
}

// EventStreamInput is arguments of EventStream
type EventStreamInput struct {
	AppID *string

	// Timeout is waiting seconds to retrieve DataFeedURL. Because another DataFeedURL
	// can not be retrieved if same AppID process is running,
	Timeout *int
}

// EventStream generates channel of event stream
func (x *SensorAPI) EventStream(input *EventStreamInput) chan *StreamQueue {
	ch := make(chan *StreamQueue, StreamEventQueueSize)
	if input == nil {
		input = &EventStreamInput{}
	}

	go func() {
		defer close(ch)
		wait := 0.0

		appID := strings.Replace(uuid.New().String()[:23], "-", "", -1)
		if input.AppID != nil {
			appID = *input.AppID
		}
		timeout := 120
		if input.Timeout != nil {
			timeout = *input.Timeout
		}
		startTime := time.Now()

		var output *EntitiesDatafeedOutput
		var err error

		for {
			output, err = x.EntitiesDatafeed(&EntitiesDatafeedInput{
				AppID: &appID,
			})
			if err != nil {
				ch <- &StreamQueue{Error: err}
				return
			}

			if len(output.Resources) > 0 {
				break
			}

			sec := time.Duration(math.Pow(2, wait))
			Logger.Warnf("No event stream info. Retry after %d sec...", sec)
			time.Sleep(time.Second * sec)
			if wait < 6 {
				wait++
			}

			now := time.Now()
			if now.Sub(startTime) > time.Second*time.Duration(timeout) {
				ch <- &StreamQueue{Error: errors.Wrap(err, "Fail to retrieve DataFeedURL, timeout")}
				return
			}
		}

		var wg sync.WaitGroup
		for _, feed := range output.Resources {
			wg.Add(1)

			go func(f DataFeedResource) {
				defer wg.Done()
				defer Logger.WithFields(logrus.Fields{
					"appId": appID,
					"url":   f.DataFeedURL,
				}).Info("Exit DataFeedURL")

				partition, err := f.Partition()
				if err != nil {
					ch <- &StreamQueue{Error: err}
					return
				}

				readCh := readEventStreamFeed(f)
				ticker := time.NewTicker(time.Minute * 25)

				for {
					select {
					case q := <-readCh:
						if q == nil {
							return
						}
						ch <- q
						if q.Error != nil {
							return
						}

					case <-ticker.C:
						_, err = x.EntitiesDatafeedAction(&EntitiesDatafeedActionInput{
							AppID:      &appID,
							ActionName: String("refresh_active_stream_session"),
							Partition:  &partition,
						})
						if err != nil {
							ch <- &StreamQueue{Error: errors.Wrap(err, "fail to unmarshal event stream")}
							return
						}

						Logger.WithFields(logrus.Fields{
							"appId":     appID,
							"partition": partition,
							"url":       f.DataFeedURL,
						}).Info("Refresh DataFeedURL")
					}
				}
			}(feed)
		}
		wg.Wait()

	}()

	return ch
}
