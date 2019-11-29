package path

import (
	"fmt"
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
	if len(t.Segments) < 1 {
		return Empty(), "", errors.New("type path not long enough")
	}
	return TypePath{
		IsAbsolute: t.IsAbsolute,
		Segments:   t.Segments[:len(t.Segments)-1],
	}, t.Segments[len(t.Segments)-1], nil
}

func (t TypePath) IndexOf(segment string) int {
	for i, s := range t.Segments {
		if s == segment {
			return i
		}
	}
	return -1
}

func (t TypePath) EndsWith(segment ...string) bool {
	if len(t.Segments) < len(segment) {
		return false
	}
	for i, seg := range segment {
		if t.Segments[len(t.Segments)-len(segment)+i] != seg {
			return false
		}
	}
	return true
}

func (t TypePath) IsVarDef() bool {
	return len(t.Segments) >= 3 && t.Segments[len(t.Segments)-2] == "var"
}

func (t TypePath) SplitVarDef() (TypePath, string) {
	if !t.IsVarDef() {
		panic("not a variable definition")
	}
	return TypePath{
		IsAbsolute: t.IsAbsolute,
		Segments:   t.Segments[:len(t.Segments)-2],
	}, t.Segments[len(t.Segments)-1]
}

func (t TypePath) CheckKeywords() error {
	varIndex := t.IndexOf("var")
	if varIndex >= 0 && varIndex < len(t.Segments)-2 || t.EndsWith("var", "var") {
		return fmt.Errorf("invalid path %v: var not expected", t)
	}
	return nil
}

func (t TypePath) String() string {
	if t.IsAbsolute {
		return "/" + strings.Join(t.Segments, "/")
	} else {
		return strings.Join(t.Segments, "/")
	}
}

func (t TypePath) Equals(other TypePath) bool {
	if t.IsAbsolute != other.IsAbsolute || len(t.Segments) != len(other.Segments) {
		return false
	}
	for i, path := range t.Segments {
		if other.Segments[i] != path {
			return false
		}
	}
	return true
}

func ParseTypePath(path string) (TypePath, error) {
	output := Empty()
	origPath := path
	if strings.HasPrefix(path, "/") {
		output = Root()
		path = path[1:]
	}
	if strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}
	if len(path) > 0 {
		for _, segment := range strings.Split(path, "/") {
			if segment == "" {
				return Empty(), fmt.Errorf("invalid path %q: empty segment", origPath)
			}
			output = output.Add(segment)
		}
	}
	return output, nil
}

func ConstTypePath(path string) TypePath {
	tp, err := ParseTypePath(path)
	if err != nil {
		panic(fmt.Sprintf("constant type paths should always be valid, but %q wasn't", path))
	}
	return tp
}
