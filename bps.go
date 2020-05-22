package bps

import "github.com/google/uuid"

// GenClientID generates random client ID.
// It is guaranteed to:
// start with "bps-",
// only contain [0-9A-Za-z-]
func GenClientID() string {
	// uuid.New may panic only if random generator failures,
	// which should be very rare (if happen at all):
	return "bps-" + uuid.New().String()
}
