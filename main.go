package main

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"github.com/rivo/uniseg"
)

func init() {
	if len(os.Args) > 1 && os.Args[1] == "tcell" {
		return // continue to tcell demo in main()
	}
	runGocuiDemo()
	os.Exit(0)
}

var icons = []struct {
	Name string
	Icon string
}{
	{"Gear", "⚙️"},
	{"Frame", "🖼️"},
	{"Trash", "🗑️"},
	{"Left", "⬅️"},
	{"Right", "➡️"},
	{"Info", "ℹ️"},
	{"TextGear", "⛭"},
	{"Package", "📦"},
}

// Characters affected by EastAsianWidth setting
var ambiguousChars = []struct {
	Name string
	Char string
}{
	{"Greek α", "α"},
	{"Greek Σ", "Σ"},
	{"Cyrillic Д", "Д"},
	{"Roman Ⅳ", "Ⅳ"},
	{"Arrow →", "→"},
	{"Music ♪", "♪"},
	{"Star ★", "★"},
	{"Circle ●", "●"},
}

var screen tcell.Screen
var eastAsianWidth = false

func main() {
	var err error
	screen, err = tcell.NewScreen()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating screen: %v\n", err)
		os.Exit(1)
	}
	if err := screen.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing screen: %v\n", err)
		os.Exit(1)
	}
	defer screen.Fini()

	runewidth.DefaultCondition.EastAsianWidth = eastAsianWidth

	draw()
	screen.Show()

	for {
		ev := screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyCtrlC || ev.Key() == tcell.KeyEscape || ev.Rune() == 'q' {
				return
			}
			if ev.Rune() == 't' {
				eastAsianWidth = !eastAsianWidth
				screen.Sync()
				draw()
				screen.Show()
			}
		case *tcell.EventResize:
			screen.Sync()
			draw()
			screen.Show()
		}
	}
}

func draw() {
	screen.Clear()
	w, _ := screen.Size()

	style := tcell.StyleDefault
	boldStyle := tcell.StyleDefault.Bold(true)
	greenStyle := tcell.StyleDefault.Foreground(tcell.ColorGreen)
	redStyle := tcell.StyleDefault.Foreground(tcell.ColorRed)

	y := 0

	// Title
	title := fmt.Sprintf("Unicode Width Test (EastAsianWidth: %v) - Press 't' to toggle, 'q' to quit", eastAsianWidth)
	drawString(0, y, title, boldStyle)
	y += 2

	// Data table header
	drawString(0, y, "Icon | Bytes | Runes | runewidth | uniseg | Drift | FE0F", style)
	y++
	drawString(0, y, strings.Repeat("-", 60), style)
	y++

	for _, icon := range icons {
		bytes := len(icon.Icon)
		runes := utf8.RuneCountInString(icon.Icon)
		rwWidth := runewidth.StringWidth(icon.Icon)
		unisegWidth := uniseg.StringWidth(icon.Icon)
		drift := unisegWidth - runes
		hasFE0F := strings.ContainsRune(icon.Icon, 0xFE0F)

		// Draw icon using proper grapheme handling
		x := 1
		x = drawGrapheme(x, y, icon.Icon, style)

		// Draw rest of row
		rest := fmt.Sprintf("   |   %d   |   %d   |     %d     |    %d   |  %2d   | %v",
			bytes, runes, rwWidth, unisegWidth, drift, hasFE0F)
		drawString(4, y, rest, style)
		y++
	}
	y++

	// Box tests
	const boxWidth = 16
	boxHeight := len(icons) + 2

	// BROKEN - on its own row (corrupts anything on same row)
	drawString(0, y, "BROKEN: FE0F as separate cell (causes ││ artifact)", redStyle)
	y++
	drawBox(0, y, boxWidth, boxHeight, style)
	boxStartY := y
	y++
	for _, icon := range icons {
		// Clear the content area first
		for x := 1; x < boxWidth-1; x++ {
			screen.SetContent(x, y, ' ', nil, style)
		}
		// Broken: each rune in separate cell, including FE0F
		// FE0F as standalone cell causes visual artifact (double │)
		x := 2
		for _, r := range icon.Icon {
			screen.SetContent(x, y, r, nil, style)
			x += runewidth.RuneWidth(r)
		}
		y++
	}
	y = boxStartY + boxHeight + 1

	// FIXED | WORKAROUND - side by side (both use correct rendering)
	col2X := 50
	drawString(0, y, "FIXED: combc[] for FE0F", greenStyle)
	drawString(col2X, y, "WORKAROUND: strip FE0F", style)
	y++
	drawBox(0, y, boxWidth, boxHeight, style)
	drawBox(col2X, y, boxWidth, boxHeight, style)
	boxStartY = y
	y++
	for _, icon := range icons {
		// FIXED - grapheme clusters
		drawGrapheme(2, y, icon.Icon, style)
		// WORKAROUND - strip FE0F
		stripped := strings.ReplaceAll(icon.Icon, "\uFE0F", "")
		drawGrapheme(col2X+2, y, stripped, style)
		y++
	}
	y = boxStartY + boxHeight + 1

	// Icon+Text side by side
	textBoxWidth := 20
	textCol2 := textBoxWidth + 2
	drawString(0, y, "Icon+Text (no space)", style)
	drawString(textCol2, y, "Icon + Text (with space)", style)
	y++
	drawBox(0, y, textBoxWidth, boxHeight, style)
	drawBox(textCol2, y, textBoxWidth+2, boxHeight, style)
	boxStartY = y
	y++
	for _, icon := range icons {
		// Left box: no space
		x := 2
		x = drawGrapheme(x, y, icon.Icon, style)
		drawString(x, y, icon.Name, style)
		// Right box: with space
		x = textCol2 + 2
		x = drawGrapheme(x, y, icon.Icon, style)
		x = drawString(x, y, " ", style)
		drawString(x, y, icon.Name, style)
		y++
	}
	y = boxStartY + boxHeight + 1

	// Embedded titles - BROKEN on own row
	titleBoxW := 22
	drawString(0, y, "Title BROKEN: per-rune", redStyle)
	y++
	drawBoxWithTitle(0, y, titleBoxW, 3, "⚙️ Settings", false, style)
	y += 4

	// FIXED | Multi-icon side by side
	drawString(0, y, "Title FIXED: combc[]", greenStyle)
	drawString(titleBoxW+2, y, "Multi-icon: combc[]", greenStyle)
	y++
	drawBoxWithTitle(0, y, titleBoxW, 3, "⚙️ Settings", true, style)
	drawBoxWithTitle(titleBoxW+2, y, titleBoxW+8, 3, "📦 Pkg ⚙️ Cfg", true, style)
	y += 4

	// More embedded title examples - all FIXED
	bw := 16
	drawBoxWithTitle(0, y, bw, 3, "🖼️ Images", true, style)
	drawBoxWithTitle(bw+1, y, bw, 3, "🗑️ Trash", true, style)
	drawBoxWithTitle((bw+1)*2, y, bw, 3, "ℹ️ Info", true, style)
	drawBoxWithTitle((bw+1)*3, y, bw, 3, "📦 Pkg", true, style)
	y += 4

	// EastAsianWidth test - these chars change width with 't' toggle
	yellowStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow)
	drawString(0, y, fmt.Sprintf("EastAsianWidth: %v (press 't' to toggle)", eastAsianWidth), yellowStyle)
	y++
	drawString(0, y, "Char | RW | Name (RW=1 when false, RW=2 when true)", style)
	y++
	cond := &runewidth.Condition{EastAsianWidth: eastAsianWidth}
	for _, c := range ambiguousChars {
		rw := cond.StringWidth(c.Char)
		drawGrapheme(0, y, c.Char, style)
		drawString(4, y, fmt.Sprintf("| %d  | %s", rw, c.Name), style)
		y++
	}

	// Draw width indicator at bottom
	_, h := screen.Size()
	drawString(0, h-1, fmt.Sprintf("Screen: %dx%d | 't'=toggle EastAsian, 'q'=quit", w, h), style)
}

// drawString draws a simple string (no special handling)
func drawString(x, y int, s string, style tcell.Style) int {
	for _, r := range s {
		screen.SetContent(x, y, r, nil, style)
		x += runewidth.RuneWidth(r)
	}
	return x
}

// drawGrapheme draws a string using proper grapheme cluster handling
// This is the CORRECT way to handle emoji with combining characters
func drawGrapheme(x, y int, s string, style tcell.Style) int {
	gr := uniseg.NewGraphemes(s)
	for gr.Next() {
		runes := gr.Runes()
		width := gr.Width()
		if len(runes) > 0 {
			// First rune is main character, rest are combining characters
			screen.SetContent(x, y, runes[0], runes[1:], style)
			x += width
		}
	}
	return x
}

// drawBox draws a box at the given position
func drawBox(x, y, w, h int, style tcell.Style) {
	// Corners
	screen.SetContent(x, y, '┌', nil, style)
	screen.SetContent(x+w-1, y, '┐', nil, style)
	screen.SetContent(x, y+h-1, '└', nil, style)
	screen.SetContent(x+w-1, y+h-1, '┘', nil, style)

	// Horizontal lines
	for i := x + 1; i < x+w-1; i++ {
		screen.SetContent(i, y, '─', nil, style)
		screen.SetContent(i, y+h-1, '─', nil, style)
	}

	// Vertical lines
	for i := y + 1; i < y+h-1; i++ {
		screen.SetContent(x, i, '│', nil, style)
		screen.SetContent(x+w-1, i, '│', nil, style)
	}
}

// drawBoxWithTitle draws a box with title embedded in top edge: ┌─ Title ─┐
func drawBoxWithTitle(x, y, w, h int, title string, useGrapheme bool, style tcell.Style) {
	// Draw basic box first
	drawBox(x, y, w, h, style)

	// Draw title embedded in top edge
	// Format: ┌─ Title ─────┐
	titleX := x + 2
	screen.SetContent(titleX, y, ' ', nil, style)
	titleX++

	if useGrapheme {
		// FIXED: use grapheme clusters
		titleX = drawGrapheme(titleX, y, title, style)
	} else {
		// BROKEN: per-rune like gocui
		for _, r := range title {
			screen.SetContent(titleX, y, r, nil, style)
			titleX += runewidth.RuneWidth(r)
		}
	}

	screen.SetContent(titleX, y, ' ', nil, style)
}
