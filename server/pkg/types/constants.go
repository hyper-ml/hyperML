package types

import "fmt"

const (
	nullString     = ""
	OutPathPostfix = "_out"
)

func uint64toString(i uint64) string {
	return fmt.Sprintf("%d", i)
}
