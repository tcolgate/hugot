package hugot

type Storer interface {
	Get(key []byte) ([]byte, error)
	Set(key []byte, value []byte) error
	Unset(key []byte) error
	List(prefix []byte)
}
