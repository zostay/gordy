package parser

import (
	"bufio"
	"errors"
	"io"
	"sync"
	"unicode/utf8"
)

type batch struct {
	closed bool
}

type Buffer struct {
	r       *bufio.Reader
	lock    sync.Mutex
	offsets []int
}

func NewBuffer(r io.Reader) *Buffer {
	return &Buffer{r: bufio.NewReader(r)}
}

func NewBufferSize(r io.Reader, size int) *Buffer {
	return &Buffer{r: bufio.NewReaderSize(r, size)}
}

func (b *Buffer) peek(
	off int,
	p []byte,
) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	pbs, err := b.r.Peek(off + len(p))
	if err != nil {
		return 0, err
	}

	copy(p, pbs[off:])

	return len(pbs[off:]), nil
}

func (b *Buffer) discard(n int) {
	_, _ = b.r.Discard(n)
}

func (b *Buffer) peekRunes(off int, p []rune) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	pbs, err := b.r.Peek(off + len(p))
	if err != nil && !(len(pbs) > 0 && errors.Is(err, io.EOF)) {
		return 0, err
	}

	atEof := false
	if errors.Is(err, io.EOF) {
		atEof = true
	}

	pbs = pbs[off:]
	glued := make([]byte, 0, len(p)*utf8.UTFMax)
	total := 0
	var readErr error
	for i := 0; i < len(p); i++ {
		var n int
		readErr = nil
		for {
			switch {
			case rune(pbs[0]) < utf8.RuneSelf:
				// single byte rune, add it to the output and move on
				p[i] = rune(pbs[0])
				pbs = pbs[1:]
				total += 1
				break

			case utf8.FullRune(pbs):
				// complete multi-byte rune, add it to the output and move on
				p[i], n = utf8.DecodeRune(pbs)
				pbs = pbs[n:]
				total += n
				break

			case atEof:
				// EOF reached, decode the partial and quit
				p[i], n = utf8.DecodeRune(pbs)
				total += n
				return total, nil

			default:
				// if we get here and we already had a read error last time,
				// let's just croak here and now. This ensures that we quit.
				if readErr != nil {
					return 0, readErr
				}

				// we don't have a full rune, but there's more input: read more
				glued = glued[:0]
				glued = append(glued, pbs...)
				pbs, err = b.r.Peek(off + total + len(p) - i)
				if err != nil && !(len(pbs) > 0 && errors.Is(err, io.EOF)) {
					return 0, err
				}

				atEof = errors.Is(err, io.EOF)

				glued = append(glued, pbs[off+total:]...)
				pbs = glued
			}
		}
	}

	return total, nil
}

type Reader struct {
	buf *Buffer
	n   int
}

func (b *Buffer) Reader() *Reader {
	return &Reader{b, 0}
}

func (b *Buffer) Collect(r *Reader) {
	b.discard(r.n)
}

func (r *Reader) Clone() *Reader {
	return &Reader{r.buf, r.n}
}

func (r *Reader) Read(p []byte) (n int, err error) {
	r.buf.lock.Lock()
	defer r.buf.lock.Unlock()

	n, err = r.buf.peek(r.n, p)
	r.n += n
	if err != nil {
		return n, err
	}

	return n, nil
}

func (r *Reader) ReadRunes(p []rune) (n int, err error) {
	r.buf.lock.Lock()
	defer r.buf.lock.Unlock()

	n, err = r.buf.peekRunes(r.n, p)
	r.n += n
	if err != nil {
		return n, err
	}

	return n, nil
}

func (r *Reader) Reset() {
	r.n = 0
}
