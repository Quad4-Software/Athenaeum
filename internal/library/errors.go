package library

import "errors"

var (
	errNoContainer = errors.New("epub: missing META-INF/container.xml")
	errNoRootfile  = errors.New("epub: no rootfile declared in container")
)
