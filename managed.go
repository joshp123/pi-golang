package pi

import "github.com/joshp123/pi-golang/internal/sdk"

func ClassifyManaged(result RunDetailedResult) ManagedSummary {
	return sdk.ClassifyManaged(result)
}

func ClassifyRunError(err error) (BrokenCause, bool) {
	return sdk.ClassifyRunError(err)
}
