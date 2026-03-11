package logging

const (
	Method   = "method"
	Args     = "args"
	Cost     = "cost_ms"
	Response = "response"
	Error    = "error"
)

type ArgFormatter interface {
	FormatArg() (string, error)
}
