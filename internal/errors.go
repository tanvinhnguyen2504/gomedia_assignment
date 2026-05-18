package internal

import "errors"

var (
	ErrNotFound      = errors.New("viewing not found")
	ErrConflict      = errors.New("agent already has a viewing at that time")
	ErrInvalidStatus = errors.New("one or more viewings have an invalid status for this action")
	ErrMissingField  = errors.New("missing required field")
	ErrPastDate      = errors.New("scheduled_at must be in the future")
	ErrInvalidAction = errors.New("invalid action")
)
