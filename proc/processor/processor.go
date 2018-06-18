package processor

// Processor encapsulates a set of top-level business logic routines.
type Processor interface {
	SuspendChannel(id, jobCreator string, agent bool) (string, error)
	ActivateChannel(id, jobCreator string, agent bool) (string, error)
	TerminateChannel(id, jobCreator string, agent bool) (string, error)
}
