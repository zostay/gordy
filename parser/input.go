package parser

import (
	"fmt"
	"io"
	"reflect"
	"runtime"
	"strings"
)

// Tracer is a function that is use to log or report parser traces. This
// function signature was chosen because it is commonly available, such as
// fmt.Print or log.Println, etc.
type Tracer func(v ...any)

type Stage int

const (
	StageTry Stage = iota
	StageGot
	StageFail
)

// Input provides the tool for keeping track of how the parser input is being
// read during the parsing process.
type Input struct {
	TraceFunc Tracer

	parent *Input
	buf    *Buffer
	r      *Reader
}

// New creates a new parser for recursive descent parsing using the
// default buffer size (inherited from bufio.Reader).
func New(r io.Reader) *Input {
	buf := NewBuffer(r)
	return &Input{
		buf: buf,
		r:   buf.Reader(),
	}
}

// NewSize creates a new parser helper for recursive descent parsing, but with a
// custom internal Buffer size.
func NewSize(r io.Reader, size int) *Input {
	buf := NewBufferSize(r, size)
	return &Input{
		buf: NewBufferSize(r, size),
		r:   buf.Reader(),
	}
}

// Trace may be called to help track the progress through a parse for help in
// debugging.
func (p *Input) Trace(stage Stage, name string, args ...any) {
	if p.TraceFunc != nil {
		out := &strings.Builder{}
		switch stage {
		case StageFail:
			fmt.Fprint(out, "ERR ")
		case StageGot:
			fmt.Fprint(out, "GOT ")
		case StageTry:
			fmt.Fprint(out, "TRY ")
		}

		fmt.Fprint(out, name)
		fmt.Fprint(out, "(")

		var bs [10]byte
		n, _ := p.buf.peek(p.r.n, bs[:])
		fmt.Fprint(out, string(bs[:n]))
		fmt.Fprint(out, "…")

		for i, arg := range args {
			fmt.Fprint(out, ", ")

			if reflect.TypeOf(arg).Kind() == reflect.Func {
				fmt.Fprint(out, runtime.FuncForPC(reflect.ValueOf(arg).Pointer()).Name())
				continue
			}

			if i == len(args)-1 {
				if err, isErr := arg.(error); isErr {
					fmt.Fprintf(out, "): %v", err)
					return
				}

				if m, isMatch := arg.(*Match); isMatch {
					fmt.Fprintf(out, ") = %v", m)
					return
				}
			}

			fmt.Fprint(out, arg)
		}

		fmt.Fprint(out, ")")

		p.TraceFunc(out.String())
	}
}

// Read reads the next bytes from input.
func (p *Input) Read(bs []byte) (int, error) {
	return p.r.Read(bs)
}

// ReadRunes reads teh next runes from input.
func (p *Input) ReadRunes(rs []rune) (int, error) {
	return p.r.ReadRunes(rs)
}

// MayFail returns a new Input that can be used to read input starting at the
// offset of the current Input. Reads on the returned Input will not impact
// the parent. When finished, you may call Keep on the child parser if you are
// ready to keep the reads made.
func (p *Input) MayFail() *Input {
	return &Input{
		parent: p,
		buf:    p.buf,
		r:      p.r.Clone(),
	}
}

// Keep returns the parent Input after updating it to have the same state as
// the child.
//
// When Keep is called on the root Input object or its direct descendants, it
// will also free up memory by discarding data that won't be read again at the
// start of the buffer.
func (p *Input) Keep() *Input {
	// detect root or child of root cases
	var root *Input
	if p.parent == nil {
		root = p
	} else if p.parent.parent == nil {
		root = p.parent
	}

	// when we are at or child of root, we can discard the read bytes
	if root != nil {
		root.buf.Collect(p.r)
		root.r.Reset()
		return root
	}

	// otherwise, we just want to make sure the parent moves forward to the
	// cursor position in the input so far
	p.parent.r = p.r
	return p.parent
}

// Discard returns the parent Input without updating the state of the parent ot
// match the child.
func (p *Input) Discard() *Input {
	if p.parent != nil {
		return p.parent
	}
	return p
}
