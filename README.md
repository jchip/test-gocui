# test-gocui

A tool to verify and demonstrate Unicode emoji rendering issues in terminal TUI libraries, specifically with `gocui` and `go-runewidth`.

## The Problem

Emoji with variation selectors (like `⚙️` = U+2699 + U+FE0F) cause border misalignment in `gocui` because:

1. **gocui** calls `SetContent(x, y, rune, nil, style)` for EACH rune separately
2. **FE0F** (Variation Selector-16) is a zero-width combining character that should be passed in the `combc[]` parameter, not as a separate cell
3. **go-runewidth** returns incorrect widths for these sequences (`runewidth.StringWidth("⚙️") = 1`, but terminal renders as 2 columns)

## The Fix

Use `uniseg` to iterate grapheme clusters and pass combining characters properly:

```go
import "github.com/rivo/uniseg"

func drawGrapheme(x, y int, s string, style tcell.Style) int {
    gr := uniseg.NewGraphemes(s)
    for gr.Next() {
        runes := gr.Runes()
        // First rune is main, rest are combining (FE0F goes here!)
        screen.SetContent(x, y, runes[0], runes[1:], style)
        x += gr.Width()
    }
    return x
}
```

## Width Comparison

| Icon | Bytes | Runes | runewidth | uniseg | FE0F |
|------|-------|-------|-----------|--------|------|
| ⚙️   | 6     | 2     | 1         | 2      | true |
| 🖼️   | 7     | 2     | 1         | 2      | true |
| 📦   | 4     | 1     | 2         | 2      | false |
| ⛭    | 3     | 1     | 1         | 1      | false |

**Key insight**: `uniseg.StringWidth()` returns correct terminal column widths, `runewidth.StringWidth()` does not for emoji+FE0F sequences.

## Usage

```bash
go run main.go
# or
go build && ./test-gocui
```

### Controls

- `t` - Toggle `runewidth.DefaultCondition.EastAsianWidth`
- `q` or `Ctrl+C` - Quit

## What This Tool Tests

1. **BROKEN box** - Simulates gocui's per-rune rendering (borders misaligned)
2. **FIXED box** - Uses grapheme clusters with `combc[]` (borders aligned)
3. **WORKAROUND box** - Strips FE0F for text-style emoji
4. **Icon+Text** - Side-by-side comparison with/without space after icon
5. **Embedded titles** - Icons in box title edge (`┌─ ⚙️ Settings ─┐`)

## Workarounds for gocui Users

Until gocui is patched, you can:

1. **Strip FE0F** from emoji to use text-style variants
2. **Avoid emoji with variation selectors** in alignment-critical areas
3. **Use Nerd Font PUA icons** instead of Unicode emoji

## Dependencies

- `github.com/gdamore/tcell/v2` - Terminal handling
- `github.com/mattn/go-runewidth` - (Demonstrates the bug)
- `github.com/rivo/uniseg` - Correct grapheme cluster handling

## Related Issues

This affects any TUI using gocui with emoji, including lazydocker, lazygit, etc.
