package operation

// A cursor identifies unambiguous a position in the JSON document by describing
// the path from the root until the leaf/node selected and the element ID
type Cursor struct {
	Path []CursorElement
}

func NewCursor(path ...CursorElement) Cursor {
	c := Cursor{}
	for _, e := range path {
		c.Path = append(c.Path, e)
	}
	return c
}

type CursorElement interface {
	Get() interface{}
}

type MapKey struct {
	Key string
}

func (k MapKey) Get() interface{} {
	return k.Key
}

type ListKey struct {
	Key int
}

func (k ListKey) Get() interface{} {
	return k.Key
}
