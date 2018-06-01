package report

// Reporter interface.
// Notify takes three arguments:
// err - standard error;
// sync - if true then the function waits for the end of sending;
// skip - how many errors to remove from stacktrace.
// Enable if it is true Reporter running.
type Reporter interface {
	Notify(error, bool, int)
	Enable() bool
}
