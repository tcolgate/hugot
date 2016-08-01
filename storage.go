package hugot

// Storer is an interface to external key/value storage
type Storer interface {
	Get(key []byte) ([]byte, bool, error)
	Set(key []byte, value []byte) error
	Unset(key []byte) error
}

type prefixStore struct {
	pfx  []byte
	base Storer
}

func (p prefixStore) Get(key []byte) ([]byte, bool, error) {
	return p.base.Get(append(p.pfx, key...))
}

func (p prefixStore) Set(key []byte, value []byte) error {
	return p.base.Set(append(p.pfx, key...), value)
}

func (p prefixStore) Unset(key []byte) error {
	return p.base.Unset(append(p.pfx, key...))
}

func newPrefixedStore(pfx []byte, s Storer) prefixStore {
	return prefixStore{
		pfx:  append(pfx, []byte("#")...),
		base: s,
	}
}
