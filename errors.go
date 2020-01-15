package duckduckgo

import "fmt"

type BadCode int

func (c BadCode) Error() string { return fmt.Sprintf("bad code: %d", c) }
