package indent

import (
	"io"
	"bytes"
)

type Writer struct {
	base        io.Writer
	indentDepth int
	needsIndent bool
}

var _ io.Writer = &Writer{}

func NewWriter(base io.Writer) *Writer {
	return &Writer{
		base:        base,
		needsIndent: true,
	}
}

func (iw *Writer) Indent() {
	iw.indentDepth += 1
}

func (iw *Writer) Unindent() {
	if iw.indentDepth <= 0 {
		panic("indent going negative")
	}
	iw.indentDepth -= 1
}

func (iw *Writer) Write(data []byte) (i int, err error) {
	parts := bytes.SplitAfter(data, []byte("\n"))
	written := 0
	for _, segment := range parts {
		if len(segment) == 0 {
			continue
		}
		if iw.needsIndent {
			iw.needsIndent = false
			if iw.indentDepth > 0 {
				_, err := iw.base.Write(bytes.Repeat([]byte("\t"), iw.indentDepth))
				if err != nil {
					return written, err
				}
			}
		}
		i, err := iw.base.Write(segment)
		written += i
		if err != nil {
			return written, err
		}
		if segment[len(segment)-1] == '\n' {
			iw.needsIndent = true
		}
	}
	if written < len(data) {
		panic("inconsistency: should have written everything")
	}
	return written, nil
}
