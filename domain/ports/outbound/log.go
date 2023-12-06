// logging outbound port that supports various levels and adding fields to the log
package outbound

type LogField struct {
	Key   string
	Value interface{}
}

type Log interface {
	Debug(msg string, fields ...LogField)
	Info(msg string, fields ...LogField)
	Error(msg string, fields ...LogField)
}
