package supervisor

import (
	log "github.com/sirupsen/logrus"
)

// Logger defines supervisor log, needs to force Printf
// for zookeeper client library
type Logger interface {
	Printf(format string, p ...interface{})
	Infof(format string, p ...interface{})
	Debugf(format string, p ...interface{})
	Errorf(format string, p ...interface{})
}

// DefaultLogger default logger using logrus
type DefaultLogger struct {
}

// Printf required for zookeeper client library, it will print
// as Debugf.
func (DefaultLogger) Printf(format string, p ...interface{}) {
	log.Debugf(format, p)
}

// Infof logs a message at level Info on the standard logger.
func (DefaultLogger) Infof(format string, p ...interface{}) {
	log.Infof(format, p)
}

// Debugf logs a message at level Debug on the standard logger.
func (DefaultLogger) Debugf(format string, p ...interface{}) {
	log.Debugf(format, p)
}

// Warnf logs a message at level Warn on the standard logger.
func (DefaultLogger) Warnf(format string, p ...interface{}) {
	log.Warnf(format, p)
}

// Errorf logs a message at level Error on the standard logger.
func (DefaultLogger) Errorf(format string, p ...interface{}) {
	log.Errorf(format, p)
}
