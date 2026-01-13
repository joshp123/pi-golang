package pi

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

var defaultShutdownTimeout = 2 * time.Second

type Client struct {
	options        Options
	command        Command
	process        *exec.Cmd
	stdin          io.WriteCloser
	stderr         bytes.Buffer
	pending        map[string]chan Response
	subscribers    map[chan Event]struct{}
	requestCounter uint64
	lock           sync.Mutex
	writeLock      sync.Mutex
	closeOnce      sync.Once
	closed         chan struct{}
}

func Start(options Options) (*Client, error) {
	options = options.withDefaults()
	config, err := resolveModelConfig(options)
	if err != nil {
		return nil, err
	}
	command, err := ResolveCommand()
	if err != nil {
		return nil, err
	}

	env, err := buildEnv(options)
	if err != nil {
		return nil, err
	}

	args := []string{
		"--mode", "rpc",
		"--provider", config.provider,
		"--model", config.model,
		"--thinking", config.thinking,
		"--no-session",
	}
	if strings.TrimSpace(options.SystemPrompt) != "" {
		args = append(args, "--system-prompt", options.SystemPrompt)
	}

	cmd := exec.Command(command.Executable, command.WithArgs(args)...)
	cmd.Env = env
	if options.WorkDir != "" {
		cmd.Dir = options.WorkDir
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	client := &Client{
		options:     options,
		command:     command,
		process:     cmd,
		stdin:       stdin,
		pending:     map[string]chan Response{},
		subscribers: map[chan Event]struct{}{},
		closed:      make(chan struct{}),
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	go client.captureStderr(stderr)
	go client.readStdout(stdout)

	return client, nil
}

func (client *Client) Close() error {
	var closeErr error
	client.closeOnce.Do(func() {
		close(client.closed)
		if client.stdin != nil {
			_ = client.stdin.Close()
		}
		if client.process != nil && client.process.Process != nil {
			_ = client.process.Process.Signal(syscall.SIGTERM)
		}

		done := make(chan struct{})
		go func() {
			if client.process != nil {
				_ = client.process.Wait()
			}
			close(done)
		}()

		select {
		case <-done:
		case <-time.After(defaultShutdownTimeout):
			if client.process != nil && client.process.Process != nil {
				_ = client.process.Process.Kill()
			}
		}

		client.lock.Lock()
		for ch := range client.subscribers {
			close(ch)
		}
		client.subscribers = map[chan Event]struct{}{}
		for _, responseChan := range client.pending {
			close(responseChan)
		}
		client.pending = map[string]chan Response{}
		client.lock.Unlock()
	})

	return closeErr
}

func (client *Client) Stderr() string {
	return client.stderr.String()
}

func (client *Client) Subscribe(buffer int) (<-chan Event, func()) {
	if buffer <= 0 {
		buffer = 16
	}
	ch := make(chan Event, buffer)
	client.lock.Lock()
	client.subscribers[ch] = struct{}{}
	client.lock.Unlock()

	return ch, func() {
		client.lock.Lock()
		delete(client.subscribers, ch)
		close(ch)
		client.lock.Unlock()
	}
}

func (client *Client) Send(ctx context.Context, command RpcCommand) (Response, error) {
	if command == nil {
		return Response{}, errors.New("command is required")
	}
	if _, ok := command["type"]; !ok {
		return Response{}, errors.New("command type is required")
	}

	requestID := client.nextRequestID()
	command["id"] = requestID

	payload, err := json.Marshal(command)
	if err != nil {
		return Response{}, err
	}

	responseChan := make(chan Response, 1)
	client.lock.Lock()
	client.pending[requestID] = responseChan
	client.lock.Unlock()

	client.writeLock.Lock()
	_, writeErr := client.stdin.Write(append(payload, '\n'))
	client.writeLock.Unlock()
	if writeErr != nil {
		client.dropPending(requestID)
		return Response{}, writeErr
	}

	select {
	case <-client.closed:
		client.dropPending(requestID)
		return Response{}, errors.New("client closed")
	case <-ctx.Done():
		client.dropPending(requestID)
		return Response{}, ctx.Err()
	case response, ok := <-responseChan:
		if !ok {
			return Response{}, errors.New("response channel closed")
		}
		if !response.Success {
			if response.Error == "" {
				return response, errors.New("rpc command failed")
			}
			return response, errors.New(response.Error)
		}
		return response, nil
	}
}

func (client *Client) Prompt(ctx context.Context, message string, images ...ImageContent) error {
	if strings.TrimSpace(message) == "" {
		return errors.New("message is required")
	}
	command := RpcCommand{
		"type":    "prompt",
		"message": message,
	}
	if len(images) > 0 {
		command["images"] = images
	}
	_, err := client.Send(ctx, command)
	return err
}

func (client *Client) Run(ctx context.Context, message string) (RunResult, error) {
	if err := client.Prompt(ctx, message); err != nil {
		return RunResult{}, err
	}

	events, cancel := client.Subscribe(8)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return RunResult{}, ctx.Err()
		case event, ok := <-events:
			if !ok {
				return RunResult{}, errors.New("event stream closed")
			}
			if event.Type == "agent_end" {
				return extractRunResult(event)
			}
		}
	}
}

func (client *Client) nextRequestID() string {
	value := atomic.AddUint64(&client.requestCounter, 1)
	return fmt.Sprintf("req-%d", value)
}

func (client *Client) dropPending(requestID string) {
	client.lock.Lock()
	responseChan, ok := client.pending[requestID]
	if ok {
		delete(client.pending, requestID)
		close(responseChan)
	}
	client.lock.Unlock()
}

func (client *Client) captureStderr(stderr io.Reader) {
	_, _ = io.Copy(&client.stderr, stderr)
}

func (client *Client) readStdout(stdout io.Reader) {
	reader := bufio.NewReader(stdout)
	for {
		line, err := reader.ReadBytes('\n')
		line = bytes.TrimSpace(line)
		if len(line) > 0 {
			client.handleLine(line)
		}
		if err != nil {
			return
		}
	}
}

func (client *Client) handleLine(line []byte) {
	var envelope struct {
		Type string `json:"type"`
		ID   string `json:"id,omitempty"`
	}
	if err := json.Unmarshal(line, &envelope); err != nil {
		client.broadcastEvent(Event{Type: "parse_error", Raw: append([]byte{}, line...)})
		return
	}

	if envelope.Type == "response" {
		var response Response
		if err := json.Unmarshal(line, &response); err != nil {
			client.broadcastEvent(Event{Type: "response_parse_error", Raw: append([]byte{}, line...)})
			return
		}
		client.lock.Lock()
		responseChan := client.pending[response.ID]
		delete(client.pending, response.ID)
		client.lock.Unlock()
		if responseChan != nil {
			responseChan <- response
			close(responseChan)
		}
		return
	}

	client.broadcastEvent(Event{Type: envelope.Type, Raw: append([]byte{}, line...)})
}

func (client *Client) broadcastEvent(event Event) {
	client.lock.Lock()
	for ch := range client.subscribers {
		select {
		case ch <- event:
		default:
		}
	}
	client.lock.Unlock()
}
