package pi

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

func (client *Client) captureStderr(stderr io.Reader) {
	buffer := make([]byte, 4096)
	for {
		read, err := stderr.Read(buffer)
		if read > 0 {
			client.appendStderr(buffer[:read])
		}
		if err != nil {
			return
		}
	}
}

func (client *Client) readStdout(stdout io.Reader) {
	reader := bufio.NewReader(stdout)
	for {
		line, err := reader.ReadBytes('\n')
		line = bytes.TrimSpace(line)
		if len(line) > 0 {
			client.handleLine(line)
		}
		if err == nil {
			continue
		}
		client.markProcessDied(err)
		return
	}
}

func (client *Client) dispatchEvents() {
	defer close(client.eventDispatchEnd)
	for {
		event, ok := client.eventQueue.pop()
		if !ok {
			return
		}
		client.events.publish(event)
	}
}

func (client *Client) waitForProcess() {
	client.waitErr = client.process.Wait()
	close(client.waitDone)

	select {
	case <-client.closed:
		return
	default:
	}
	if client.waitErr != nil {
		client.markProcessDied(client.waitErr)
		return
	}

	client.markProcessDied(io.EOF)
}

func (client *Client) handleLine(line []byte) {
	var envelope struct {
		Type string `json:"type"`
		ID   string `json:"id,omitempty"`
	}
	if err := json.Unmarshal(line, &envelope); err != nil {
		client.enqueueEvent(Event{Type: eventTypeParseError, Raw: append([]byte(nil), line...)})
		return
	}
	if strings.TrimSpace(envelope.Type) == "" {
		client.enqueueEvent(Event{Type: eventTypeParseError, Raw: append([]byte(nil), line...)})
		return
	}

	if envelope.Type == eventTypeResponse {
		response, err := decodeRPCResponse(line)
		if err != nil {
			client.enqueueEvent(Event{Type: eventTypeResponseParseError, Raw: append([]byte(nil), line...)})
			return
		}
		if client.requests.resolve(response) {
			return
		}
		client.enqueueEvent(Event{Type: eventTypeResponse, Raw: append([]byte(nil), line...)})
		return
	}

	client.enqueueEvent(Event{Type: envelope.Type, Raw: append([]byte(nil), line...)})
}

func (client *Client) enqueueEvent(event Event) {
	if client.eventQueue == nil {
		return
	}
	_ = client.eventQueue.push(event)
}

func (client *Client) markProcessDied(cause error) {
	client.processErrOnce.Do(func() {
		select {
		case <-client.closed:
			return
		default:
		}

		processErr := ErrProcessDied
		if cause != nil && !errors.Is(cause, io.EOF) {
			processErr = fmt.Errorf("%w: %v", ErrProcessDied, cause)
		}

		client.requests.markProcessDied(processErr)
		client.events.processDied(newProcessDiedEvent(cause))
		client.stopEventDispatch()
	})
}

func (client *Client) closeAll(processErr error) {
	client.requests.close(processErr)
	client.events.close()
	client.stopEventDispatch()
}

func (client *Client) currentProcessError() error {
	return client.requests.currentError()
}

func newProcessDiedEvent(cause error) Event {
	payload := map[string]any{"type": EventTypeProcessDied}
	if cause != nil && !errors.Is(cause, io.EOF) {
		payload["error"] = cause.Error()
	}
	raw, _ := json.Marshal(payload)
	return Event{Type: EventTypeProcessDied, Raw: raw}
}
