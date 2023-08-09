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
	var tmp Event
	// Normalize the new lines in the event
	normalizedMsg := bytes.Replace(msg, []byte("\n\r"), []byte("\n"), -1)
	msgHeaders := bytes.Split(normalizedMsg, []byte("\n"))
	for _, line := range msgHeaders {
		// Implementation spec is taken from:
		// https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events#fields
		switch headerName(line) {
		case headerData:
			tmp.Data = append(tmp.Data, dataValue(line)...)
			// append new line after each data line per spec
			tmp.Data = append(tmp.Data, []byte("\n")...)
		default:
			// Ignore all other headers. Note we omitted the "id", "event", and "retry" headers.
		}
	}
	// Trim trailing new lines as specified in the spec
	tmp.Data = bytes.TrimRight(tmp.Data, "\n")
	*e = tmp
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
	// If there is no colon, value is empty
	if cIndex == -1 {
		return []byte{}
	}
	// If there is a colon, return the value after the colon
	return append(data[cIndex+1:])
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
