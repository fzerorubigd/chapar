package initializer

import "context"

// Interface is the initializer object
type Interface interface {
	// Initialize is called with a context, finalization signal is the context done channel
	Initialize(ctx context.Context) error
}

// Do is a simple helper function to call initialize the object if its initializer
func Do(ctx context.Context, in interface{}) error {
	if i, ok := in.(Interface); ok {
		return i.Initialize(ctx)
	}
	return nil
}
