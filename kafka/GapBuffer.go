package main

import "fmt"

const defaultCapacity = 16

type GapBuffer struct {
	buf      []rune
	gapStart int // index of first rune IN the gap (free space)
	gapEnd   int // index of first rune AFTER the gap (real text resumes)
	offset   int // cumulative rune shift from inserts/deletes since buffer creation
}

func NewGapBuffer(text string, cursor int) (*GapBuffer, error) {
	runes := []rune(text)
	textLen := len(runes)

	if cursor < 0 || cursor > textLen {
		return nil, fmt.Errorf("cursor position out of range")
	}

	capacity := (textLen + defaultCapacity) * 2
	buf := make([]rune, capacity)

	copy(buf, runes[:cursor])

	gapEnd := capacity - (textLen - cursor)
	copy(buf[gapEnd:], runes[cursor:])

	return &GapBuffer{
		buf:      buf,
		gapStart: cursor,
		gapEnd:   gapEnd,
	}, nil
}

func (g *GapBuffer) Len() int {
	return len(g.buf) - (g.gapEnd - g.gapStart)
}

func (g *GapBuffer) gapSize() int {
	return g.gapEnd - g.gapStart
}

func (g *GapBuffer) grow() {
	newCap := len(g.buf) * 2
	newBuf := make([]rune, newCap)

	copy(newBuf, g.buf[:g.gapStart])

	afterLen := len(g.buf) - g.gapEnd
	newGapEnd := newCap - afterLen
	copy(newBuf[newGapEnd:], g.buf[g.gapEnd:])

	g.buf = newBuf
	g.gapEnd = newGapEnd
}

func (g *GapBuffer) MoveCursor(pos int) error {
	if pos < 0 || pos > g.Len() {
		return fmt.Errorf("cursor position out of range")
	}

	if pos < g.gapStart {
		delta := g.gapStart - pos
		g.gapEnd -= delta
		copy(g.buf[g.gapEnd:], g.buf[pos:g.gapStart])
		g.gapStart = pos
	} else if pos > g.gapStart {
		delta := pos - g.gapStart
		copy(g.buf[g.gapStart:], g.buf[g.gapEnd:g.gapEnd+delta])
		g.gapStart += delta
		g.gapEnd += delta
	}
	return nil
}

func (g *GapBuffer) Insert(r rune) {
	if g.gapSize() == 0 {
		g.grow()
	}
	g.buf[g.gapStart] = r
	g.gapStart++
	g.offset++
}

func (g *GapBuffer) InsertString(s string) {
	for _, r := range s {
		g.Insert(r)
	}
}

func (g *GapBuffer) Delete() {
	if g.gapStart == 0 {
		return
	}
	g.gapStart--
	g.offset--
}

func (g *GapBuffer) DeleteForward() {
	if g.gapEnd == len(g.buf) {
		return
	}
	g.gapEnd++
}

func (g *GapBuffer) String() string {
	result := make([]rune, 0, g.Len())
	result = append(result, g.buf[:g.gapStart]...)
	result = append(result, g.buf[g.gapEnd:]...)
	return string(result)
}

func (g *GapBuffer) CursorPos() int {
	return g.gapStart
}

// OriginalCursorPos returns the cursor position in the original (pre-edit) coordinate space.
func (g *GapBuffer) OriginalCursorPos() int {
	return g.gapStart - g.offset
}

// MoveCursorOriginal moves the cursor to an original-coordinate position,
// adjusting for accumulated inserts/deletes.
func (g *GapBuffer) MoveCursorOriginal(pos int) error {
	return g.MoveCursor(max(pos+g.offset, 0))
}

func (g *GapBuffer) debugView() {
	fmt.Print("buf: [")
	for i, r := range g.buf {
		if i == g.gapStart {
			fmt.Print("|GAP|")
		}
		if i >= g.gapStart && i < g.gapEnd {
			fmt.Print("_")
		} else {
			fmt.Printf("%c", r)
		}
	}
	if g.gapStart == len(g.buf) {
		fmt.Print("|GAP|")
	}
	fmt.Printf("] gapStart=%d gapEnd=%d len=%d\n", g.gapStart, g.gapEnd, g.Len())
}
