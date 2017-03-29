package web

import (
	"bufio"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEventIfEventIsEmptyReturnsFalse(t *testing.T) {
	t.Parallel()
	// given
	event := &Event{timestamp: time.Now(),
		eventType: "status_update_event",
		body:      []byte(`{"id": "simpleId"}`),
		id:        "id",
	}
	// when
	expected := false
	actual := event.isEmpty()
	// then
	assert.Equal(t, expected, actual)
}

func TestEventIfEventIsEmptyReturnsTrue(t *testing.T) {
	t.Parallel()
	// given
	event := &Event{}
	// when
	expected := true
	actual := event.isEmpty()
	// then
	assert.Equal(t, expected, actual)
}

func TestParseLineWhenStautsUpdateEventPassed(t *testing.T) {
	t.Parallel()
	// given
	event := &Event{}
	line0 := []byte("id: 0\n")
	line1 := []byte("event: status_update_event\n")
	line2 := []byte("data: testData\n")
	expected0 := "0"
	expected1 := "status_update_event"
	expected2 := []byte("testData\n")
	// when
	event.parseLine(line0)
	event.parseLine(line1)
	event.parseLine(line2)
	// then
	assert.Equal(t, expected0, event.id)
	assert.Equal(t, expected1, event.eventType)
	assert.Equal(t, expected2, event.body)
}

func TestParseLineWhenGarbageIsProvidedBodyShouldBeNil(t *testing.T) {
	t.Parallel()
	// given
	event := &Event{}
	line := []byte("garbage data\n")
	expectedBody := []byte(nil)
	// when
	_ = event.parseLine(line)
	// then
	assert.Equal(t, expectedBody, event.body)
}

func TestParseEventWhenOneStatusUpdateEventIsProvided(t *testing.T) {
	t.Parallel()
	// given
	sreader := strings.NewReader("event: status_update_event\ndata: testData\n")
	reader := bufio.NewReader(sreader)
	expectedEvent := "status_update_event"
	// when
	event, _ := parseEvent(reader)
	// then
	assert.Equal(t, expectedEvent, event.eventType)
}

func TestParseEventWhenSimpleDataIsProvidedShouldReturnEOFError(t *testing.T) {
	t.Parallel()
	// given
	sreader := strings.NewReader("event: status_update_event\ndata: testData\n")
	reader := bufio.NewReader(sreader)
	expectedEvent := "status_update_event"
	expectedError := errors.New("EOF")
	// when
	event, err := parseEvent(reader)
	// then
	assert.Equal(t, expectedError, err)
	assert.Equal(t, expectedEvent, event.eventType)
}

func TestParseEventWhenSimpleDataIsProvidedAndNotCompleteEventIsProvidedShouldReturnUnexpectedEOFError(t *testing.T) {
	t.Parallel()
	// given
	sreader := strings.NewReader("event: status_update_event\ndata: testData\nlkajsd")
	reader := bufio.NewReader(sreader)
	expectedEvent := "status_update_event"
	expectedError := errors.New("Unexpected EOF")
	// when
	event, err := parseEvent(reader)
	// then
	assert.Equal(t, expectedError, err)
	assert.Equal(t, expectedEvent, event.eventType)
}

func TestParseEventWhenNotCompleteDataIsProvidedEventShouldContainOnlyEventType(t *testing.T) {
	t.Parallel()
	// given
	sreader := strings.NewReader("event: status_update_event\ndata")
	reader := bufio.NewReader(sreader)
	expectedEvent := "status_update_event"
	// when
	event, _ := parseEvent(reader)
	// then
	assert.Equal(t, expectedEvent, event.eventType)
	assert.Nil(t, event.body)
}

func TestParseEventWhenDataIsEmptyProvidedEventShouldContainOnlyEventType(t *testing.T) {
	t.Parallel()
	// given
	sreader := strings.NewReader("event: status_update_event\ndata:\n")
	reader := bufio.NewReader(sreader)
	expectedEvent := "status_update_event"
	// when
	event, _ := parseEvent(reader)
	// then
	assert.Equal(t, expectedEvent, event.eventType)
	assert.Nil(t, event.body)
}

func TestParseEventWhenDataIsProvidedButNoLFShouldContainDataEvenIfNotComplete(t *testing.T) {
	t.Parallel()
	// given
	sreader := strings.NewReader("event: status_update_event\ndata: testEventData")
	reader := bufio.NewReader(sreader)
	expectedEvent := "status_update_event"
	// when
	event, _ := parseEvent(reader)
	// then
	assert.Equal(t, expectedEvent, event.eventType)
	assert.Equal(t, []byte("testEventData\n"), event.body)
}
