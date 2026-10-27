package ui

import (
	"strings"
)

// BrailleCanvas provides a 2x4 dot resolution canvas for drawing graphs.
// Each cell in the terminal corresponds to 2 columns and 4 rows of dots.
type BrailleCanvas struct {
	widthCells  int
	heightCells int
	grid        [][]byte // [cellY][cellX] bitmask for braille dots
}

// NewBrailleCanvas creates a new canvas with cell dimensions.
func NewBrailleCanvas(widthCells, heightCells int) *BrailleCanvas {
	if widthCells < 1 {
		widthCells = 1
	}
	if heightCells < 1 {
		heightCells = 1
	}
	grid := make([][]byte, heightCells)
	for y := range grid {
		grid[y] = make([]byte, widthCells)
	}
	return &BrailleCanvas{
		widthCells:  widthCells,
		heightCells: heightCells,
		grid:        grid,
	}
}

// Dot positions in Braille Unicode (U+2800 + mask):
// Row 0: left = 0x01, right = 0x08
// Row 1: left = 0x02, right = 0x10
// Row 2: left = 0x04, right = 0x20
// Row 3: left = 0x40, right = 0x80
var dotMasks = [4][2]byte{
	{0x01, 0x08},
	{0x02, 0x10},
	{0x04, 0x20},
	{0x40, 0x80},
}

// SetDot sets a single dot in dot-coordinates:
// dotX: 0 .. (widthCells*2 - 1)
// dotY: 0 .. (heightCells*4 - 1), where 0 is bottom and (heightCells*4 - 1) is top.
func (b *BrailleCanvas) SetDot(dotX, dotY int) {
	totalDotW := b.widthCells * 2
	totalDotH := b.heightCells * 4

	if dotX < 0 || dotX >= totalDotW || dotY < 0 || dotY >= totalDotH {
		return
	}

	// Invert Y so that dotY=0 is bottom and dotY=max is top
	screenY := (totalDotH - 1) - dotY

	cellX := dotX / 2
	subX := dotX % 2

	cellY := screenY / 4
	subY := screenY % 4

	b.grid[cellY][cellX] |= dotMasks[subY][subX]
}

// RenderAreaGraph plots a filled area graph given a series of data points (0.0 to maxVal).
// The values slice represents the history from oldest (left) to newest (right).
func (b *BrailleCanvas) RenderAreaGraph(values []float64, maxVal float64) {
	if maxVal <= 0 {
		maxVal = 100.0
	}
	totalDotW := b.widthCells * 2
	totalDotH := b.heightCells * 4

	// Clear grid
	for y := range b.grid {
		for x := range b.grid[y] {
			b.grid[y][x] = 0
		}
	}

	// Take the last totalDotW values (or pad if fewer)
	offset := len(values) - totalDotW
	for dotX := 0; dotX < totalDotW; dotX++ {
		idx := offset + dotX
		var val float64
		if idx >= 0 && idx < len(values) {
			val = values[idx]
		}
		if val < 0 {
			val = 0
		}
		if val > maxVal {
			val = maxVal
		}

		// Calculate height in dots
		dotH := int((val / maxVal) * float64(totalDotH))
		if dotH > totalDotH {
			dotH = totalDotH
		}

		// Fill from bottom (0) up to dotH
		for dotY := 0; dotY < dotH; dotY++ {
			b.SetDot(dotX, dotY)
		}
	}
}

// String returns the rendered canvas as lines separated by newlines.
func (b *BrailleCanvas) String() string {
	var sb strings.Builder
	for y := 0; y < b.heightCells; y++ {
		for x := 0; x < b.widthCells; x++ {
			mask := b.grid[y][x]
			if mask == 0 {
				sb.WriteRune(' ')
			} else {
				sb.WriteRune(rune(0x2800 + int(mask)))
			}
		}
		if y < b.heightCells-1 {
			sb.WriteRune('\n')
		}
	}
	return sb.String()
}
