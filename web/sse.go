package web

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/allegro/marathon-appcop/marathon"
	"github.com/allegro/marathon-appcop/mgc"
	"github.com/allegro/marathon-appcop/score"
)

// Stop all channels
type Stop func()

// NewHandler is main initialization function
func NewHandler(config Config, marathon marathon.Marathoner, gc *mgc.MarathonGC,
	scoreUpdate chan score.Update) Stop {

	// TODO implement proper leader election
	// Right now this part of code highly rely on marathon v2/leader endpoint
	leaderPoll(marathon, config.MyLeader)

	stopChannels := make([]chan<- stopEvent, config.WorkersCount)
	eventQueue := make(chan Event, config.QueueSize)

	for i := 0; i < config.WorkersCount; i++ {
		handler := newEventHandler(i, marathon, eventQueue, scoreUpdate)
		stopChannels[i] = handler.Start()
	}

	// start dispatcher
	sse := newSSEHandler(eventQueue, marathon.AuthGet(), marathon.LocationGet())
	dispatcherStop := sse.start()
	stopChannels = append(stopChannels, dispatcherStop)

	// schedule marathon GC job
	go gc.StartMarathonGCJob()

	return stop(stopChannels)
}

func leaderPoll(service marathon.Marathoner, myLeader string) {
	pollTicker := time.NewTicker(5 * time.Second)
	for {

		leader, err := service.LeaderGet()
		if err != nil {
			log.WithError(err).Error("Error while getting leader")
			continue
		}
		if leader == myLeader {
			break
		}
		log.Debug("I am not leader")
		<-pollTicker.C
	}

}

func stop(channels []chan<- stopEvent) Stop {
	return func() {
		for _, channel := range channels {
			channel <- stopEvent{}
		}
	}
}
