package log

// Logger is the fundamental interface for all log operations, it's compatible
// wit go-kit logger.
type Logger interface {
	Log(keyvals ...interface{}) error
}
