package server

import (
	"bytes"
	"regexp"
	"strings"
	"testing"

	"rsc.io/qr"
)

var cellPattern = regexp.MustCompile(`\x1b\[(\d+);(\d+)m▀`)

func decodeQR(t *testing.T, painted string) [][]bool {
	t.Helper()
	var grid [][]bool
	for _, line := range strings.Split(strings.TrimSuffix(painted, "\n"), "\n") {
		cells := cellPattern.FindAllStringSubmatch(line, -1)
		if len(cells) == 0 {
			t.Fatalf("line carries no cells: %q", line)
		}
		top := make([]bool, len(cells))
		bottom := make([]bool, len(cells))
		for x, c := range cells {
			top[x] = c[1] == "30"
			bottom[x] = c[2] == "40"
		}
		grid = append(grid, top, bottom)
	}
	return grid
}

func TestWriteQRPaintsTheCode(t *testing.T) {
	const text = "http://192.168.1.57:7777/#code=H8F6-7TNU"
	code, err := qr.Encode(text, qr.M)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := writeQR(&buf, text); err != nil {
		t.Fatal(err)
	}
	grid := decodeQR(t, buf.String())

	side := code.Size + 2*quietZone
	if len(grid) < side {
		t.Fatalf("painted %d module rows, want at least %d", len(grid), side)
	}
	for y := range side {
		for x := range side {
			inside := x >= quietZone && x < quietZone+code.Size && y >= quietZone && y < quietZone+code.Size
			want := inside && code.Black(x-quietZone, y-quietZone)
			if grid[y][x] != want {
				t.Fatalf("module (%d,%d) is %v, want %v", x, y, grid[y][x], want)
			}
		}
	}
}
