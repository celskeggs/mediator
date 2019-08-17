package path

import (
	"github.com/pkg/errors"
	"strings"
)

type TypePath struct {
	IsAbsolute bool
	Segments   []string
}

func Root() TypePath {
	return TypePath{
		IsAbsolute: true,
		Segments:   nil,
	}
}

func Empty() TypePath {
	return TypePath{
		IsAbsolute: false,
		Segments:   nil,
	}
}

func (t TypePath) IsEmpty() bool {
	return !t.IsAbsolute && len(t.Segments) == 0
}

func RelativeFromSegments(segments ...string) TypePath {
	replica := make([]string, len(segments))
	copy(replica, segments)
	return TypePath{
		IsAbsolute: false,
		Segments:   replica,
	}
}

func (t TypePath) Add(segments ...string) TypePath {
	return t.Join(RelativeFromSegments(segments...))
}

func (t TypePath) Join(o TypePath) TypePath {
	if o.IsAbsolute {
		return o
	}
	concat := make([]string, len(t.Segments)+len(o.Segments))
	copy(concat, t.Segments)
	copy(concat[len(t.Segments):], o.Segments)
	return TypePath{
		IsAbsolute: t.IsAbsolute,
		Segments:   concat,
	}
}

func (t TypePath) SplitLast() (TypePath, string, error) {
	if len(t.Segments) < 2 {
		return Empty(), "", errors.New("type path not long enough")
	}
	return TypePath{
		IsAbsolute: t.IsAbsolute,
		Segments:   t.Segments[:len(t.Segments)-1],
	}, t.Segments[len(t.Segments)-1], nil
}

func (t TypePath) String() string {
	if t.IsAbsolute {
		return "/" + strings.Join(t.Segments, "/")
	} else {
		return strings.Join(t.Segments, "/")
	}
}
