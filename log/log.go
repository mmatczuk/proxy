package log

// Logger is proxy logger.
type Logger interface {
	Log(keyvals ...interface{}) error
}
