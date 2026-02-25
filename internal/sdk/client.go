package sdk

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/joshp123/pi-golang/internal/stream"
	"github.com/joshp123/pi-golang/internal/transport"
)

var defaultShutdownTimeout = 2 * time.Second

var debugEnabledProvider = func() bool {
	return os.Getenv("PI_DEBUG") == "1"
}

func SetDebugEnabledProvider(provider func() bool) {
	if provider == nil {
		debugEnabledProvider = func() bool { return false }
		return
	}
	debugEnabledProvider = provider
}

func debugf(format string, args ...any) {
	if debugEnabledProvider() {
		log.Printf("[pi-golang] "+format, args...)
	}
}

type Client struct {
	process *exec.Cmd
	stdin   io.WriteCloser

	stderr   bytes.Buffer
	stderrMu sync.Mutex

	requests *transport.RequestManager
	events   *stream.Hub[Event]

	requestCounter uint64

	writeLock sync.Mutex
	closeOnce sync.Once
	closed    chan struct{}

	waitDone chan struct{}

	processErrOnce sync.Once

	eventQueue       *transport.Queue[Event]
	eventDispatchEnd chan struct{}
	stopDispatchOnce sync.Once

	runInProgress atomic.Bool

	managedCompactionHook *managedCompactionHook
}

type SessionClient struct {
	*Client
}

type OneShotClient struct {
	*Client
}

type startConfig struct {
	appName            string
	workDir            string
	systemPrompt       string
	mode               Mode
	dragons            DragonsOptions
	sessionName        string
	auth               ProviderAuth
	environment        map[string]string
	inheritEnvironment bool
	seedAuthFromHome   bool
	compactionPrompt   string
	useSession         bool
}

func StartSession(options SessionOptions) (*SessionClient, error) {
	normalized, err := normalizeSessionOptions(options)
	if err != nil {
		return nil, err
	}
	client, err := startClient(startConfig{
		appName:            normalized.AppName,
		workDir:            normalized.WorkDir,
		systemPrompt:       normalized.SystemPrompt,
		mode:               normalized.Mode,
		dragons:            normalized.Dragons,
		sessionName:        normalized.SessionName,
		auth:               normalized.Auth,
		environment:        normalized.Environment,
		inheritEnvironment: normalized.InheritEnvironment,
		seedAuthFromHome:   normalized.SeedAuthFromHome,
		compactionPrompt:   normalized.CompactionPrompt,
		useSession:         true,
	})
	if err != nil {
		return nil, err
	}
	return &SessionClient{Client: client}, nil
}

func StartOneShot(options OneShotOptions) (*OneShotClient, error) {
	normalized, err := normalizeOneShotOptions(options)
	if err != nil {
		return nil, err
	}
	client, err := startClient(startConfig{
		appName:            normalized.AppName,
		workDir:            normalized.WorkDir,
		systemPrompt:       normalized.SystemPrompt,
		mode:               normalized.Mode,
		dragons:            normalized.Dragons,
		auth:               normalized.Auth,
		environment:        normalized.Environment,
		inheritEnvironment: normalized.InheritEnvironment,
		seedAuthFromHome:   normalized.SeedAuthFromHome,
		compactionPrompt:   normalized.CompactionPrompt,
		useSession:         false,
	})
	if err != nil {
		return nil, err
	}
	return &OneShotClient{Client: client}, nil
}

func startClient(config startConfig) (client *Client, err error) {
	var hook *managedCompactionHook
	if strings.TrimSpace(config.compactionPrompt) != "" {
		hook, err = createManagedCompactionHook(config.compactionPrompt)
		if err != nil {
			return nil, err
		}
		defer func() {
			if err != nil && hook != nil {
				_ = hook.cleanup()
			}
		}()
	}

	modelConfig, err := resolveModelConfig(config.mode, config.dragons)
	if err != nil {
		return nil, err
	}
	command, err := ResolveCommand()
	if err != nil {
		return nil, err
	}
	if err := validateProviderAuth(modelConfig.provider, config.auth); err != nil {
		return nil, err
	}

	environment := cloneStringMap(config.environment)
	if hook != nil {
		hook.injectEnvironment(environment)
	}
	env, err := buildEnv(config.appName, config.inheritEnvironment, config.seedAuthFromHome, config.auth, environment)
	if err != nil {
		return nil, err
	}

	args := []string{
		"--mode", "rpc",
		"--provider", modelConfig.provider,
		"--model", modelConfig.model,
		"--thinking", modelConfig.thinking,
	}
	if hook != nil {
		args = append(args, hook.arguments()...)
	}
	if config.useSession {
		if strings.TrimSpace(config.sessionName) != "" {
			args = append(args, "--session", config.sessionName)
		}
	} else {
		args = append(args, "--no-session")
	}
	if strings.TrimSpace(config.systemPrompt) != "" {
		args = append(args, "--system-prompt", config.systemPrompt)
	}

	cmd := exec.Command(command.Executable, command.WithArgs(args)...)
	cmd.Env = env
	if config.workDir != "" {
		cmd.Dir = config.workDir
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

	client = &Client{
		process:               cmd,
		stdin:                 stdin,
		requests:              transport.NewRequestManager(ErrClientClosed),
		events:                newEventHub(),
		closed:                make(chan struct{}),
		waitDone:              make(chan struct{}),
		eventQueue:            transport.NewQueue[Event](),
		eventDispatchEnd:      make(chan struct{}),
		managedCompactionHook: hook,
	}

	if err = cmd.Start(); err != nil {
		return nil, err
	}

	go client.captureStderr(stderr)
	go client.dispatchEvents()
	go client.readStdout(stdout)
	go client.waitForProcess()

	return client, nil
}

func (client *Client) Close() error {
	client.closeOnce.Do(func() {
		close(client.closed)
		if client.stdin != nil {
			_ = client.stdin.Close()
		}
		if client.process != nil && client.process.Process != nil {
			_ = client.process.Process.Signal(syscall.SIGTERM)
		}

		select {
		case <-client.waitDone:
		case <-time.After(defaultShutdownTimeout):
			if client.process != nil && client.process.Process != nil {
				_ = client.process.Process.Kill()
			}
			<-client.waitDone
		}

		client.closeAll(nil)
		select {
		case <-client.eventDispatchEnd:
		case <-time.After(250 * time.Millisecond):
		}
		if client.managedCompactionHook != nil {
			_ = client.managedCompactionHook.cleanup()
			client.managedCompactionHook = nil
		}
	})
	return nil
}

func (client *Client) Stderr() string {
	client.stderrMu.Lock()
	defer client.stderrMu.Unlock()
	return client.stderr.String()
}

func (client *Client) appendStderr(chunk []byte) {
	if len(chunk) == 0 {
		return
	}
	client.stderrMu.Lock()
	_, _ = client.stderr.Write(chunk)
	client.stderrMu.Unlock()
}

func (client *Client) stopEventDispatch() {
	client.stopDispatchOnce.Do(func() {
		if client.eventQueue != nil {
			client.eventQueue.Close()
		}
	})
}

func (client *Client) nextRequestID() string {
	value := atomic.AddUint64(&client.requestCounter, 1)
	return fmt.Sprintf("req-%d", value)
}

func (client *Client) Subscribe(policy SubscriptionPolicy) (<-chan Event, func(), error) {
	if err := client.currentProcessError(); err != nil {
		return nil, nil, err
	}
	if client.isClosed() {
		return nil, nil, ErrClientClosed
	}
	if err := validateSubscriptionPolicy(policy); err != nil {
		return nil, nil, err
	}
	return client.events.Subscribe(toStreamPolicy(policy))
}
