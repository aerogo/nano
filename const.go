package nano

import "time"

const (
	// diskWriteDelay specifies the delay in milliseconds until a new write will happen.
	diskWriteDelay = 250 * time.Millisecond
)
