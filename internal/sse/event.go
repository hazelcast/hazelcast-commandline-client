package sse

import (
	"bufio"
	"bytes"
	"io"
)

var (
	eventSeparator = []byte("\n\n")
	headerData     = "data"
)

type Event struct {
	Data []byte
	// Omitted fields: ID, Event, Retry
}

func (e Event) Empty() bool {
	return len(e.Data) == 0
}

func (e *Event) Unmarshal(msg []byte) {
	var data []byte
	// Normalize the new lines in the event
	normalizedMsg := bytes.Replace(msg, []byte("\n\r"), []byte("\n"), -1)
	lines := bytes.Split(normalizedMsg, []byte("\n"))
	// Implementation spec is taken from:
	// https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events#fields
	for _, line := range lines {
		switch headerName(line) {
		case headerData:
			data = append(data, dataValue(line)...)
			// append new line after each data line per spec
			data = append(data, '\n')
		default:
			// Ignore all other headers. Note we omitted the "id", "event", and "retry" headers.
		}
	}
	// Trim trailing new lines as specified in the spec
	data = bytes.TrimRight(data, "\n")
	e.Data = data
}

func headerName(data []byte) string {
	cIndex := bytes.Index(data, []byte(":"))
	// If there is no colon, whole line is the header name
	if cIndex == -1 {
		return string(data)
	}
	return string(data[:cIndex])
}

func dataValue(data []byte) []byte {
	cIndex := bytes.Index(data, []byte(":"))
	if cIndex == -1 || len(data) == cIndex+1 {
		return nil
	}
	data = data[cIndex+1:]
	// remove the space after the colon
	if len(data) > 0 && data[0] == ' ' {
		data = data[1:]
	}
	return data
}

type EventScanner struct {
	scanner *bufio.Scanner
}

func NewEventScanner(eventStream io.Reader) *EventScanner {
	scanner := bufio.NewScanner(eventStream)
	split := func(data []byte, isEOF bool) (int, []byte, error) {
		if len(data) == 0 && isEOF {
			return 0, nil, nil
		}
		// Send the event, we have found the event separator
		if i := bytes.Index(data, eventSeparator); i >= 0 {
			return i + len(eventSeparator), data[0:i], nil
		}
		// If we're at EOF, we have all of the data.
		if isEOF {
			return len(data), data, nil
		}
		// Request more data.
		return 0, nil, nil
	}
	scanner.Split(split)
	return &EventScanner{
		scanner: scanner,
	}
}

func (e *EventScanner) ReadEvent() (*Event, error) {
	if !e.scanner.Scan() {
		if err := e.scanner.Err(); err != nil {
			return nil, err
		}
		// EOF is expected when the connection is closed by the server
		return nil, io.EOF
	}
	event := e.scanner.Bytes()
	var ev Event
	ev.Unmarshal(event)
	if ev.Empty() {
		return nil, nil
	}
	return &ev, nil
}
