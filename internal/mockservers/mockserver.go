// Package mockservers defines the available mockserver for the telemetry data
package mockservers

import "context"

// Mockserver defines what a mockserver should do
type Mockserver interface {
	Replay(context.Context, bool)
	Record(output string)
}
