package ssdp

type ST uint

const (
	ALL ST = iota
)

func (st ST) String() string {
	return []string{"ssdp:all"}[st]
}
