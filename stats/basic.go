package stats

type Basic struct {
	base
}

func NewBasicStats() *Basic {
	b := &Basic{}
	b.init()

	return b
}
