package pi

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestWithDefaultRequestTimeoutRejectsNilContext(t *testing.T) {
	_, _, err := withDefaultRequestTimeout(nil)
	if !errors.Is(err, ErrNilContext) {
		t.Fatalf("expected ErrNilContext, got %v", err)
	}
}

func TestWithDefaultRequestTimeoutAddsDeadlineWhenMissing(t *testing.T) {
	ctx, cancel, err := withDefaultRequestTimeout(context.Background())
	if err != nil {
		t.Fatalf("withDefaultRequestTimeout returned error: %v", err)
	}
	defer cancel()

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("expected deadline to be set")
	}
	if time.Until(deadline) <= 0 {
		t.Fatal("expected deadline in the future")
	}
}

func TestValidatePromptRequestRejectsInvalidStreamingBehavior(t *testing.T) {
	err := validatePromptRequest(PromptRequest{Message: "hello", StreamingBehavior: StreamingBehavior("nope")}, true)
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestValidatePromptRequestRejectsStreamingBehaviorWhenNotAllowed(t *testing.T) {
	err := validatePromptRequest(PromptRequest{Message: "hello", StreamingBehavior: StreamingBehaviorSteer}, false)
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestEncodeImagesUsesRPCCanonicalShape(t *testing.T) {
	images := encodeImages([]ImageContent{{Data: "abc", MIMEType: "image/png"}})
	if len(images) != 1 {
		t.Fatalf("expected one image, got %d", len(images))
	}
	if images[0]["type"] != "image" {
		t.Fatalf("expected type=image, got %q", images[0]["type"])
	}
	if images[0]["mimeType"] != "image/png" {
		t.Fatalf("expected mimeType=image/png, got %q", images[0]["mimeType"])
	}
}

func TestAbortCommandUsesUpstreamType(t *testing.T) {
	command := abortCommand()
	commandType, err := commandTypeOf(command)
	if err != nil {
		t.Fatalf("commandTypeOf returned error: %v", err)
	}
	if commandType != rpcCommandAbort {
		t.Fatalf("expected abort command type, got %q", commandType)
	}
	if len(command) != 1 {
		t.Fatalf("expected abort command to contain only type, got %d fields", len(command))
	}
}
