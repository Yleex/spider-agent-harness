package scope

import "errors"

var (
	ErrOutsideScope = errors.New("path fuera del alcance del proyecto")
)
