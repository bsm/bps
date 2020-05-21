package bps

import "github.com/google/uuid"

// GenClientID generates random client ID.
func GenClientID() string {
	// uuid.New may panic only if random generator failures,
	// which should be very rare (if happen at all):
	return "bps-" + uuid.New().String()
}
