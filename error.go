package externalip

import "errors"

type InvalidIPError string

func (err InvalidIPError) Error() string {
	return "Invalid IP: " + string(err)
}

var (
	NoIPError               = errors.New("no IP could be found")
	InsufficientWeightError = errors.New("a voter's weight has to be at least 1")
	NoSourceError           = errors.New("no voter's source given")
)
