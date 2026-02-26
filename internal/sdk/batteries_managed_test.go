package sdk

import (
	"context"
	"fmt"
	"testing"
)

func TestClassifyManaged(t *testing.T) {
	compactionResult := &CompactResult{Summary: "ok", FirstKeptEntryID: "entry-1", TokensBefore: 1000}

	tests := []struct {
		name      string
		result    RunDetailedResult
		wantClass CompletionClass
		wantFacts RecoveryFacts
	}{
		{
			name: "completed_no_compaction",
			result: RunDetailedResult{
				Outcome: TerminalOutcome{Status: TerminalStatusCompleted},
			},
			wantClass: CompletionClassOK,
			wantFacts: RecoveryFacts{},
		},
		{
			name: "completed_overflow_compaction_success",
			result: RunDetailedResult{
				Outcome:             TerminalOutcome{Status: TerminalStatusCompleted},
				AutoCompactionStart: &AutoCompactionStartEvent{Reason: "overflow"},
				AutoCompactionEnd:   &AutoCompactionEndEvent{Result: compactionResult},
			},
			wantClass: CompletionClassOKAfterRecovery,
			wantFacts: RecoveryFacts{
				CompactionObserved: true,
				OverflowDetected:   true,
				Recovered:          true,
			},
		},
		{
			name: "completed_non_overflow_compaction",
			result: RunDetailedResult{
				Outcome:             TerminalOutcome{Status: TerminalStatusCompleted},
				AutoCompactionStart: &AutoCompactionStartEvent{Reason: "manual"},
				AutoCompactionEnd:   &AutoCompactionEndEvent{Result: compactionResult},
			},
			wantClass: CompletionClassOK,
			wantFacts: RecoveryFacts{CompactionObserved: true},
		},
		{
			name: "completed_overflow_compaction_failed",
			result: RunDetailedResult{
				Outcome:             TerminalOutcome{Status: TerminalStatusCompleted},
				AutoCompactionStart: &AutoCompactionStartEvent{Reason: "overflow"},
				AutoCompactionEnd:   &AutoCompactionEndEvent{Result: nil, ErrorMessage: "boom"},
			},
			wantClass: CompletionClassOK,
			wantFacts: RecoveryFacts{
				CompactionObserved: true,
				OverflowDetected:   true,
				Recovered:          false,
			},
		},
		{
			name: "aborted_precedence_over_recovery",
			result: RunDetailedResult{
				Outcome:             TerminalOutcome{Status: TerminalStatusAborted},
				AutoCompactionStart: &AutoCompactionStartEvent{Reason: "overflow"},
				AutoCompactionEnd:   &AutoCompactionEndEvent{Result: compactionResult},
			},
			wantClass: CompletionClassAborted,
			wantFacts: RecoveryFacts{
				CompactionObserved: true,
				OverflowDetected:   true,
				Recovered:          false,
			},
		},
		{
			name: "failed_terminal",
			result: RunDetailedResult{
				Outcome: TerminalOutcome{Status: TerminalStatusFailed},
			},
			wantClass: CompletionClassFailed,
			wantFacts: RecoveryFacts{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := ClassifyManaged(test.result)
			if got.Class != test.wantClass {
				t.Fatalf("unexpected class: got=%q want=%q", got.Class, test.wantClass)
			}
			if got.Facts != test.wantFacts {
				t.Fatalf("unexpected facts: got=%+v want=%+v", got.Facts, test.wantFacts)
			}
		})
	}
}

func TestClassifyRunError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantCause  BrokenCause
		wantBroken bool
	}{
		{name: "nil", err: nil, wantBroken: false},
		{name: "context_canceled", err: context.Canceled, wantBroken: false},
		{name: "context_deadline", err: context.DeadlineExceeded, wantBroken: false},
		{name: "process_died", err: ErrProcessDied, wantCause: BrokenCauseProcessDied, wantBroken: true},
		{name: "process_died_wrapped", err: fmt.Errorf("wrapped: %w", ErrProcessDied), wantCause: BrokenCauseProcessDied, wantBroken: true},
		{name: "protocol", err: ErrProtocolViolation, wantCause: BrokenCauseProtocol, wantBroken: true},
		{name: "client", err: ErrClientClosed, wantCause: BrokenCauseClient, wantBroken: true},
		{name: "rpc_error", err: &RPCError{Command: "prompt", Message: "bad"}, wantBroken: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cause, broken := ClassifyRunError(test.err)
			if broken != test.wantBroken {
				t.Fatalf("unexpected broken flag: got=%v want=%v", broken, test.wantBroken)
			}
			if cause != test.wantCause {
				t.Fatalf("unexpected cause: got=%q want=%q", cause, test.wantCause)
			}
		})
	}
}
