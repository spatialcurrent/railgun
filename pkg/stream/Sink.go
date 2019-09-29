package stream

import (
	"github.com/pkg/errors"

	"github.com/spatialcurrent/go-pipe/pkg/pipe"
)

type Sink struct {
	init func() (interface{}, error)
	it   pipe.Iterator
}

func NewSink(fn func() (interface{}, error)) *Sink {
	return &Sink{init: fn, it: nil}
}

func (s *Sink) Next() (interface{}, error) {
	if s.it == nil {
		objects, err := s.init()
		if err != nil {
			return nil, errors.Wrap(err, "error initializing sink")
		}
		it, err := pipe.NewSliceIterator(objects)
		if err != nil {
			return nil, errors.Wrap(err, "error creating iterator for objects")
		}
		s.it = it
	}
	return s.it.Next()
}
