package server

import (
	"io"
	"strings"

	"rsc.io/qr"
)

// The spec's quiet zone; a scanner reads nothing without it.
const quietZone = 4

// The upper half block puts two module rows in one cell, keeping it square.
func writeQR(w io.Writer, text string) error {
	code, err := qr.Encode(text, qr.M)
	if err != nil {
		return err
	}
	side := code.Size + 2*quietZone
	dark := func(x, y int) bool {
		x, y = x-quietZone, y-quietZone
		if x < 0 || y < 0 || x >= code.Size || y >= code.Size {
			return false
		}
		return code.Black(x, y)
	}
	var b strings.Builder
	for y := 0; y < side; y += 2 {
		for x := 0; x < side; x++ {
			// Explicit colours: a camera reads dark-on-light whatever the terminal theme.
			fg, bg := "97", "107"
			if dark(x, y) {
				fg = "30"
			}
			if dark(x, y+1) {
				bg = "40"
			}
			b.WriteString("\x1b[" + fg + ";" + bg + "m▀")
		}
		b.WriteString("\x1b[0m\n")
	}
	_, err = io.WriteString(w, b.String())
	return err
}
