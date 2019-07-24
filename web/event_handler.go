package web

import (
	"bytes"
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/allegro/marathon-appcop/marathon"
	"github.com/allegro/marathon-appcop/metrics"
	"github.com/allegro/marathon-appcop/score"
)

type eventHandler struct {
	id          int
	marathon    marathon.Marathoner
	eventQueue  <-chan Event
	scoreUpdate chan score.Update
}

type stopEvent struct{}

const (
	taskFinished = "TASK_FINISHED"
	taskFailed   = "TASK_FAILED"
	taskKilled   = "TASK_KILLED"
	taskRunning  = "TASK_RUNNING"
)

func newEventHandler(id int, marathon marathon.Marathoner, eventQueue <-chan Event,
	scoreUpdate chan score.Update) *eventHandler {
	return &eventHandler{
		id:          id,
		marathon:    marathon,
		eventQueue:  eventQueue,
		scoreUpdate: scoreUpdate,
	}
}

// Start event handler
func (fh *eventHandler) Start() chan<- stopEvent {
	var event Event
	process := func() {
		err := fh.handleEvent(event.eventType, event.body)
		if err != nil {
			metrics.Mark("events.processing.error")
		} else {
			metrics.Mark("events.processing.succes")
		}
	}

	quitChan := make(chan stopEvent)
	log.WithField("Id", fh.id).Println("Starting worker")
	go func() {
		for {
			select {
			case event = <-fh.eventQueue:
				metrics.Mark(fmt.Sprintf("events.handler.%d", fh.id))
				metrics.UpdateGauge("events.queue.len", int64(len(fh.eventQueue)))
				metrics.UpdateGauge("events.queue.delay_ns", time.Since(event.timestamp).Nanoseconds())
				metrics.Time("events.processing."+event.eventType, process)
			case <-quitChan:
				log.WithField("Id", fh.id).Info("Stopping worker")
			}
		}
	}()
	return quitChan
}

func (fh *eventHandler) handleEvent(eventType string, body []byte) error {

	body = replaceTaskIDWithID(body)

	switch eventType {
	case "status_update_event":
		return fh.handleStatusEvent(body)
	case "unhealthy_task_kill_event":
		return fh.handleUnhealthyTaskKillEvent(body)
	default:
		log.WithField("EventType", eventType).Debug("Not handled event type")
		return nil
	}
}

func (fh *eventHandler) handleStatusEvent(body []byte) error {
	task, err := marathon.ParseTask(body)

	if err != nil {
		log.WithField("Body", body).Error("Could not parse event body")
		return err
	}

	log.WithFields(log.Fields{
		"Id":         task.ID,
		"TaskStatus": task.TaskStatus,
	}).Debug("Got StatusEvent")

	appMetric := task.GetMetric(fh.marathon.GetAppIDPrefix())
	metrics.MarkApp(appMetric)

	switch task.TaskStatus {
	case taskFinished, taskFailed, taskKilled:
		appID := task.AppID
		app, err := fh.marathon.AppGet(appID)
		if err != nil {
			return err
		}
		fh.scoreUpdate <- score.Update{App: app, Update: 1}
		return nil
	case taskRunning:
		log.WithFields(log.Fields{
			"Id":    task.AppID,
			"Host":  task.Host,
			"Ports": task.Ports,
		}).Info("Got task running status")
		return nil
	default:
		log.WithFields(log.Fields{
			"Id":         task.ID,
			"taskStatus": task.TaskStatus,
		}).Debug("Not handled task status")
		return nil
	}
}

func (fh *eventHandler) handleUnhealthyTaskKillEvent(body []byte) error {
	task, err := marathon.ParseTask(body)

	if err != nil {
		log.WithField("Body", body).Error("Could not parse event body")
		return err
	}

	log.WithFields(log.Fields{
		"Id": task.ID,
	}).Debug("Got Unhealthy TaskKilled Event")

	// update score killed app
	appID := task.AppID
	app, err := fh.marathon.AppGet(appID)
	if err != nil {
		log.WithField("appID", appID).Error("Could not get app by id")
		return err
	}
	fh.scoreUpdate <- score.Update{App: app, Update: 1}
	return nil
}

// for every other use of Tasks, Marathon uses the "id" field for the task ID.
// Here, it uses "taskId", with most of the other fields being equal. We'll
// just swap "taskId" for "id" in the body so that we can successfully parse
// incoming events.
func replaceTaskIDWithID(body []byte) []byte {
	return bytes.Replace(body, []byte("taskId"), []byte("id"), -1)
}
