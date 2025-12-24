package main

import (
	"fmt"
	"log"

	"github.com/jesseduffield/gocui"
)

// Demo to expose gocui's FE0F emoji rendering bugs
// Run: ./test-gocui gocui

func runGocuiDemo() {
	g, err := gocui.NewGui(gocui.NewGuiOpts{
		OutputMode: gocui.OutputTrue,
	})
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.SetManagerFunc(gocuiLayout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, gocuiQuit); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("", gocui.KeyEsc, gocui.ModNone, gocuiQuit); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("", 'q', gocui.ModNone, gocuiQuit); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err.Error() != "quit" {
		log.Panicln(err)
	}
}

func gocuiLayout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	// Title test views - these use drawTitle() which has the FE0F bug
	titleTests := []string{
		"⚙️ Settings",
		"🖼️ Images",
		"🗑️ Trash",
		"📦 Packages",
	}

	viewWidth := 18
	y := 1

	// Header
	if v, _ := g.SetView("header", 0, 0, maxX-1, 2, 0); v != nil {
		if v.Buffer() == "" {
			v.Frame = false
			fmt.Fprintln(v, "gocui FE0F Bug Demo - Press 'q' to quit")
		}
		g.SetCurrentView("header")
	}

	// Title bug demo
	if v, _ := g.SetView("title-label", 0, y, 30, y+2, 0); v != nil && v.Buffer() == "" {
		v.Frame = false
		fmt.Fprintln(v, "TITLE BUG (drawTitle):")
	}
	y += 2

	for i, title := range titleTests {
		name := fmt.Sprintf("title-%d", i)
		x0 := (i % 4) * (viewWidth + 1)
		x1 := x0 + viewWidth
		if x1 >= maxX {
			x1 = maxX - 1
		}

		if v, _ := g.SetView(name, x0, y, x1, y+3, 0); v != nil {
			v.Title = title
			if v.Buffer() == "" {
				fmt.Fprintln(v, "Look at title ^")
			}
		}
	}
	y += 5

	// Content bug demo
	if v, _ := g.SetView("content-label", 0, y, 35, y+2, 0); v != nil && v.Buffer() == "" {
		v.Frame = false
		fmt.Fprintln(v, "CONTENT BUG (cell stores rune):")
	}
	y += 2

	// Content view - write emoji directly to view content
	contentWidth := 40
	contentHeight := len(titleTests) + 4
	if v, _ := g.SetView("content", 0, y, contentWidth, y+contentHeight, 0); v != nil {
		v.Title = "Content with FE0F emoji"
		if v.Buffer() == "" {
			fmt.Fprintln(v, "Each line has emoji with FE0F:")
			fmt.Fprintln(v, "")
			for _, icon := range titleTests {
				fmt.Fprintf(v, "  %s\n", icon)
			}
			fmt.Fprintln(v, "")
			fmt.Fprintln(v, "If broken: extra space or artifacts")
		}
	}
	y += contentHeight + 2

	// Info view
	if y+6 < maxY {
		if v, _ := g.SetView("info", 0, y, maxX-1, y+5, 0); v != nil && v.Buffer() == "" {
			v.Title = "Info"
			fmt.Fprintln(v, "The FE0F bug causes:")
			fmt.Fprintln(v, "1. Title/Content: FE0F treated as separate char with width, causing extra spaces")
			fmt.Fprintln(v, "2. Misaligned lines and rendering artifacts on right edge")
		}
	}

	return nil
}

func gocuiQuit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
