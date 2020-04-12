package types

import (
	"fmt"
)

const (
	nullString = ""
	// OutPathPostfix : Output path post fix for notebook jobs
	OutPathPostfix = "_out"
)

func uint64toString(i uint64) string {
	return fmt.Sprintf("%d", i)
}
