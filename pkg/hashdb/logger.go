package hashdb

type nullLogger struct {
}

func (l nullLogger) Errorf(string, ...interface{})   {}
func (l nullLogger) Warningf(string, ...interface{}) {}
func (l nullLogger) Infof(string, ...interface{})    {}
func (l nullLogger) Debugf(string, ...interface{})   {}
