package web

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"net/url"
	"time"
)

// Event holds state of parsed fields from marathon EventStream
type Event struct {
	timestamp time.Time
	eventType string
	body      []byte
	id        string
}

func (e *Event) parseLine(line []byte) bool {

	// https://www.w3.org/TR/2011/WD-eventsource-20110208/
	// Quote: Lines must be separated by either a U+000D CARRIAGE RETURN U+000A
	// LINE FEED (CRLF) character pair, a single U+000A LINE FEED (LF) character,
	// or a single U+000D CARRIAGE RETURN (CR) character.
	line = bytes.TrimSuffix(line, []byte{'\n'})
	line = bytes.TrimSuffix(line, []byte{'\r'})

	//If the line is empty (a blank line)
	if len(line) == 0 {
		//Dispatch the event, as defined below.
		return !e.isEmpty()
	}

	//If the line starts with a U+003A COLON character (:)
	if bytes.HasPrefix(line, []byte{':'}) {
		//Ignore the line.
		return false
	}

	var field string
	var value []byte
	//If the line contains a U+003A COLON character (:)
	//Collect the characters on the line before the first U+003A COLON character (:), and let field be that string.
	split := bytes.SplitN(line, []byte{':'}, 2)
	if len(split) == 2 {
		field = string(split[0])
		//Collect the characters on the line after the first U+003A COLON character (:), and let value be that string.
		//If value starts with a U+0020 SPACE character, remove it from value.
		value = bytes.TrimPrefix(split[1], []byte{' '})
	} else {
		//Otherwise, the string is not empty but does not contain a U+003A COLON character (:)
		//Process the field using the steps described below, using the whole line as the field name,
		//and the empty string as the field value.
		field = string(line)
		value = []byte{}

	}
	stringValue := string(value)
	//If the field name is
	switch field {
	case "event":
		//Set the event name buffer to field value.
		e.eventType = stringValue
	case "data":
		//If the data buffer is not the empty string,
		if len(value) != 0 {
			//Append the field value to the data buffer,
			//then append a single U+000A LINE FEED (LF) character to the data buffer.
			e.body = append(e.body, value...)
			e.body = append(e.body, '\n')
		}
	case "id":
		//Set the last event ID buffer to the field value.
		e.id = stringValue
	case "retry":
		// TODO consider reconnection delay
	}

	return false
}

func (e *Event) isEmpty() bool {
	return e.eventType == "" && e.body == nil && e.id == ""
}

func parseEvent(reader *bufio.Reader) (Event, error) {
	e := Event{}
	for dispatch := false; !dispatch; {
		//TODO: Use scanner use ReadLine
		line, err := reader.ReadBytes('\n')
		if err == io.EOF {
			dispatch = e.parseLine(line)
			if !dispatch {
				return e, errors.New("Unexpected EOF")
			}
			return e, io.EOF
		}
		if err != nil {
			return e, err
		}
		dispatch = e.parseLine(line)
	}
	return e, nil
}

func subscribeURL(auth *url.Userinfo, location string) string {

	marathonurl := url.URL{
		Scheme: "http",
		User:   auth,
		Host:   location,
		Path:   "/v2/events",
	}
	query := marathonurl.Query()

	marathonurl.RawQuery = query.Encode()
	return marathonurl.String()
}
