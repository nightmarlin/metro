package metro

import (
	"fmt"
)

var (
	ErrNotFound     = fmt.Errorf("not found")
	ErrWrongLine    = fmt.Errorf("wrong line")
	ErrInvalidRoute = fmt.Errorf("invalid route")
)
