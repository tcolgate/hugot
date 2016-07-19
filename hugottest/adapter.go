package hugottest

type Adapter struct {
	*ResponseRecorder
	*MessagePlayer
}

func NewAdapter() Adapter {
	return Adapter{
		&ResponseRecorder{},
		&MessagePlayer{},
	}
}
