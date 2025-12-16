package query

type errorRecorder interface {
	setError(error) errorRecorder
	getError() error
}

type errorRecord struct {
	Error error
}

func (e *errorRecord) setError(err error) errorRecorder {
	e.Error = err
	return e
}

func (e *errorRecord) getError() error {
	return e.Error
}
