package service

// Service is a generalized interface for a long running asynchronous service
type Service interface {
	// Start returns an error channel that is only written to if err == nil.  It is implied that
	// the service is ready if the returned err == nil.
	Start() (errChan chan error, err error)
	// Stop will shutdown the service.  The errChan returned in Start() is expected to be closed.
	// Packages that implement the Service interface should implement a timeout to ensure Stop()
	// does not block indefinitely.
	Stop() (err error)
}
