package lint

import "time"

// subscriptionWaitDelay holds a delay to wait for subscriber to actually start consuming messages.
const subscriptionWaitDelay = 500 * time.Millisecond
