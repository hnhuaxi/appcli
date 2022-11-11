package appcli

import "errors"

var (
	ErrInvalidMapNode     = errors.New("yaml: invalid map node, need a map")
	ErrMapMemberNotEnough = errors.New("yaml: map contents not enough")
)
