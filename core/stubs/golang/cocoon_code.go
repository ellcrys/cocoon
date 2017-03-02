package golang

// CocoonCode defines the interface of a cocoon code.
type CocoonCode interface {
	Init() error
	Invoke(txID, function string, params []string) (interface{}, error)
}
