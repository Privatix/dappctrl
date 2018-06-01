package report

// Reporter interface.
// Enable if it is true Reporter running.
// Notify takes three arguments:
// err - standard error;
// sync - if true then the function waits for the end of sending;
// skip - how many errors to remove from stacktrace.
type Reporter interface {
	Enable() bool
	Notify(error, bool, int)
}
