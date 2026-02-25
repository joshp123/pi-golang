package sdk

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/joshp123/pi-golang/internal/rpc"
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
		event, ok := client.eventQueue.Pop()
		if !ok {
			return
		}
		client.events.Publish(event)
	}
}

func (client *Client) waitForProcess() {
	waitErr := client.process.Wait()
	close(client.waitDone)

	select {
	case <-client.closed:
		return
	default:
	}
	if waitErr != nil {
		client.markProcessDied(waitErr)
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
		client.enqueueEvent(Event{Type: rpc.EventParseError, Raw: append([]byte(nil), line...)})
		return
	}
	if strings.TrimSpace(envelope.Type) == "" {
		client.enqueueEvent(Event{Type: rpc.EventParseError, Raw: append([]byte(nil), line...)})
		return
	}

	if envelope.Type == rpc.EventResponse {
		response, err := decodeRPCResponse(line)
		if err != nil {
			client.enqueueEvent(Event{Type: rpc.EventResponseParseError, Raw: append([]byte(nil), line...)})
			return
		}
		if client.requests.Resolve(response) {
			return
		}
		client.enqueueEvent(Event{Type: rpc.EventResponse, Raw: append([]byte(nil), line...)})
		return
	}

	client.enqueueEvent(Event{Type: envelope.Type, Raw: append([]byte(nil), line...)})
}

func (client *Client) enqueueEvent(event Event) {
	if client.eventQueue == nil {
		return
	}
	_ = client.eventQueue.Push(event)
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

		client.requests.MarkProcessDied(processErr)
		client.events.ProcessDied(newProcessDiedEvent(cause))
		client.stopEventDispatch()
	})
}

func (client *Client) closeAll(processErr error) {
	client.requests.Close(processErr)
	client.events.Close()
	client.stopEventDispatch()
}

func (client *Client) currentProcessError() error {
	return client.requests.CurrentError()
}

func newProcessDiedEvent(cause error) Event {
	payload := map[string]any{"type": EventTypeProcessDied}
	if cause != nil && !errors.Is(cause, io.EOF) {
		payload["error"] = cause.Error()
	}
	raw, _ := json.Marshal(payload)
	return Event{Type: EventTypeProcessDied, Raw: raw}
}
