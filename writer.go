package lol

import (
	"fmt"
	"io"
	"os"
	"strings"

	isatty "github.com/mattn/go-isatty"
)

const (
	DEFAULT_SPREAD = float64(3.0)
	DEFAULT_FREQ   = float64(0.1)
	DEFAULT_ORIGIN = 0
)

const (
	ColorModeTrueColor = iota
	ColorMode256
	ColorMode0
)

// LolWriter writes a little lol-er.
type Writer struct {
	Output    io.Writer
	ColorMode int
	Freq      float64
	Spread    float64
	lineIdx   int
	Origin    int
}

var noColor = os.Getenv("TERM") == "dumb" || (!isatty.IsTerminal(os.Stdout.Fd()) && !isatty.IsCygwinTerminal(os.Stdout.Fd()))

// writeRaw will write a lol'd s to the underlying writer.  It does no line
// detection.
func (w *Writer) writeRaw(s string) (int, error) {
	c, err := w.getColorer()
	if err != nil {
		return -1, err
	}
	nWritten := 0
	for _, r := range s {
		c.rainbow(w.Freq, float64(w.Origin)+float64(w.lineIdx)/w.Spread)
		_, err := w.Output.Write(c.format())
		if err != nil {
			return nWritten, err
		}
		n, err := w.Output.Write([]byte(string(r)))
		if err != nil {
			return nWritten, err
		}
		_, err = w.Output.Write(c.reset())
		if err != nil {
			return nWritten, err
		}
		nWritten += n
		w.lineIdx++
	}
	return nWritten, nil
}

// getColorer will attempt to map the defined color mode, to a colorer{}
func (w *Writer) getColorer() (colorer, error) {
	switch w.ColorMode {
	case ColorModeTrueColor:
		return newTruecolorColorer(), nil
	case ColorMode256:
		return New256Colorer(), nil
	case ColorMode0:
		return New0Colorer(), nil
	default:
		return nil, fmt.Errorf("Invalid colorer: [%d]", w.ColorMode)
	}
}

// Write will write a byte slice to the Writer
func (w *Writer) Write(p []byte) (int, error) {
	nWritten := 0
	ss := strings.Split(string(p), "\n")
	for i, s := range ss {
		// TODO: strip out pre-existing ANSI codes and expand tabs. Would be
		// great to expand tabs in a context aware way (line linux expand
		// command).

		n, err := w.writeRaw(s)
		if err != nil {
			return nWritten, err
		}
		nWritten += n

		// Increment the Origin (line count) for each newline.  There is a
		// newline for every item in this array except the last one.
		if i != len(ss)-1 {
			n, err := w.Output.Write([]byte("\n"))
			if err != nil {
				return nWritten, err
			}
			nWritten += n
			w.Origin++
			w.lineIdx = 0
		}
	}
	return nWritten, nil
}

// NewLolWriter will return a new io.Writer with a default ColorMode of 256
func NewLolWriter() io.Writer {
	colorMode := ColorMode256
	if noColor {
		colorMode = ColorMode0
	}

	return &Writer{
		Output:    stdout,
		ColorMode: colorMode,
		Freq:      DEFAULT_FREQ,
		Spread:    DEFAULT_SPREAD,
		Origin:    DEFAULT_ORIGIN,
	}
}

// NewTruecolorLolWriter will return a new io.Writer with a default ColorMode of truecolor
func NewTruecolorLolWriter() io.Writer {
	colorMode := ColorModeTrueColor
	if noColor {
		colorMode = ColorMode0
	}
	return &Writer{
		Output:    stdout,
		ColorMode: colorMode,
	}
}
