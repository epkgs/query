package query

type errorRecorder interface {
	setError(error) errorRecorder
	getError() error
}

type errorRecord struct {
	// Error 记录查询构建过程中发生的错误。
	// 如果构建过程正常，该字段为 nil。
	Error error
}

func (e *errorRecord) setError(err error) errorRecorder {
	e.Error = err
	return e
}

func (e *errorRecord) getError() error {
	return e.Error
}
