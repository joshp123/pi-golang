package sdk

import (
	"context"
	"errors"
	"strings"
)

// Batteries layer classifiers.
//
// These are pure functions over typed SDK outputs:
// - ClassifyManaged derives a stable completion class + recovery facts from RunDetailedResult.
// - ClassifyRunError classifies runtime/process failures separately from terminal outcome classes.

func ClassifyManaged(result RunDetailedResult) ManagedSummary {
	facts := RecoveryFacts{
		CompactionObserved: result.AutoCompactionStart != nil || result.AutoCompactionEnd != nil,
		OverflowDetected:   isOverflowCompaction(result.AutoCompactionStart),
	}

	recovered := facts.OverflowDetected && compactionSucceeded(result.AutoCompactionEnd) && result.Outcome.Status == TerminalStatusCompleted
	facts.Recovered = recovered

	return ManagedSummary{
		Class: classifyCompletionClass(result.Outcome.Status, recovered),
		Facts: facts,
	}
}

func ClassifyRunError(err error) (BrokenCause, bool) {
	if err == nil {
		return "", false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return "", false
	}
	if errors.Is(err, ErrProcessDied) {
		return BrokenCauseProcessDied, true
	}
	if errors.Is(err, ErrProtocolViolation) {
		return BrokenCauseProtocol, true
	}
	if errors.Is(err, ErrClientClosed) {
		return BrokenCauseClient, true
	}
	return "", false
}

func isOverflowCompaction(event *AutoCompactionStartEvent) bool {
	if event == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(event.Reason), "overflow")
}

func compactionSucceeded(event *AutoCompactionEndEvent) bool {
	if event == nil {
		return false
	}
	if event.Result == nil {
		return false
	}
	if event.Aborted {
		return false
	}
	return strings.TrimSpace(event.ErrorMessage) == ""
}

func classifyCompletionClass(status TerminalStatus, recovered bool) CompletionClass {
	switch status {
	case TerminalStatusAborted:
		return CompletionClassAborted
	case TerminalStatusFailed:
		return CompletionClassFailed
	case TerminalStatusCompleted:
		if recovered {
			return CompletionClassOKAfterRecovery
		}
		return CompletionClassOK
	default:
		return CompletionClassFailed
	}
}
