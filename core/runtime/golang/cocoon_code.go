package golang

// CocoonCode defines the interface of a cocoon code.
type CocoonCode interface {
	OnInit() error
	OnInvoke(header Metadata, function string, params []string) (interface{}, error)
}
