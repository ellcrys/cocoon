package golang

// CocoonCode defines the interface of a cocoon code.
type CocoonCode interface {
	Init(link *Link) error
	Invoke(link *Link, txID, function string, params []string) (interface{}, error)
}
