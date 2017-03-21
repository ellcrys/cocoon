package golang

// CocoonCode defines the interface of a cocoon code.
type CocoonCode interface {
	OnInit(link *Link) error
	OnInvoke(link *Link, txID, function string, params []string) (interface{}, error)
}
