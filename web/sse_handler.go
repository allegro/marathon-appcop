package web

import (
	"bufio"
	"context"
	"io"
	"net/http"

	"github.com/sethgrid/pester"

	"net/url"

	log "github.com/Sirupsen/logrus"
)

// SSEHandler defines handler for marathon event stream, opening and closing
// subscription
type SSEHandler struct {
	eventQueue chan Event
	loc        string
	client     *pester.Client
	close      context.CancelFunc
	req        *http.Request
}

func close(r *http.Response) {
	err := r.Body.Close()
	if err != nil {
		log.WithError(err).Error("Can't close response")
	}
}

func newSSEHandler(eventQueue chan Event, auth *url.Userinfo, loc string) *SSEHandler {

	subURL := subscribeURL(auth, loc)
	req, err := http.NewRequest("GET", subURL, nil)
	if err != nil {
		log.WithError(err).Fatalf("Unable to generate sse request %v", err)
		return nil
	}

	req.Header.Set("Accept", "text/event-stream")
	ctx, cancel := context.WithCancel(context.Background())
	req = req.WithContext(ctx)

	client := pester.New()
	client.Concurrency = 1
	client.MaxRetries = 3
	client.Backoff = pester.ExponentialBackoff
	client.KeepLog = true

	return &SSEHandler{
		eventQueue: eventQueue,
		loc:        loc,
		client:     client,
		close:      cancel,
		req:        req,
	}
}

// Open connection to marathon v2/events
func (h *SSEHandler) start() chan<- stopEvent {
	res, err := h.client.Do(h.req)
	if err != nil {
		log.WithFields(log.Fields{
			"Location": h.loc,
			"Method":   "GET",
		}).Fatalf("error performing request : %v", err)
	}
	if res.StatusCode != http.StatusOK {
		log.WithFields(log.Fields{
			"Location": h.loc,
			"Method":   "GET",
		}).Errorf("Got status code : %d", res.StatusCode)
	}
	log.WithFields(log.Fields{
		"Location": h.loc,
		"Method":   "GET",
	}).Debug("Subsciption success")
	stopChan := make(chan stopEvent)
	go func() {
		<-stopChan
		h.stop()
	}()

	go func() {
		defer close(res)

		reader := bufio.NewReader(res.Body)
		for {
			e, err := parseEvent(reader)
			if err != nil {
				if err == io.EOF {
					h.eventQueue <- e
				}
				log.Fatalf("Error processing parsing event %s", err)
			}
			h.eventQueue <- e
		}
	}()
	return stopChan
}

// Close connections managed by context
func (h *SSEHandler) stop() {
	h.close()
}
