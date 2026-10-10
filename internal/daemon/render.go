package daemon

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

// RenderOptions control how a StatusInfo is drawn.
type RenderOptions struct {
	// Color enables ANSI styles.
	Color bool
	// Width is the number of columns available, 0 for a default.
	Width int
	// Now is used for ages and uptime.
	Now time.Time
}

const (
	styleReset  = "\x1b[0m"
	styleBold   = "\x1b[1m"
	styleDim    = "\x1b[2m"
	styleRed    = "\x1b[31m"
	styleGreen  = "\x1b[32m"
	styleYellow = "\x1b[33m"
	styleCyan   = "\x1b[36m"
)

type painter struct{ color bool }

func (p painter) paint(style, s string) string {
	if !p.color || s == "" || style == "" {
		return s
	}
	return style + s + styleReset
}

// pad fills s with spaces to the width (left or right aligned) and styles the result.
func (p painter) cell(style, s string, width int, right bool) string {
	s = truncate(s, width)
	fill := strings.Repeat(" ", max(width-utf8.RuneCountInString(s), 0))
	if right {
		return fill + p.paint(style, s)
	}
	return p.paint(style, s) + fill
}

func truncate(s string, width int) string {
	if width <= 0 || utf8.RuneCountInString(s) <= width {
		return s
	}
	r := []rune(s)
	if width == 1 {
		return string(r[:1])
	}
	return string(r[:width-1]) + "…"
}

func widest(min int, values ...string) int {
	for _, v := range values {
		min = max(min, utf8.RuneCountInString(v))
	}
	return min
}

func formatDuration(d time.Duration) string {
	switch {
	case d <= 0:
		return "-"
	case d < time.Millisecond:
		return d.Round(time.Microsecond).String()
	case d < time.Second:
		return d.Round(100 * time.Microsecond).String()
	case d < time.Minute:
		return d.Round(100 * time.Millisecond).String()
	default:
		return d.Round(time.Second).String()
	}
}

// RenderStatus draws the status of a daemon as text.
func RenderStatus(info StatusInfo, opts RenderOptions) string {
	p := painter{opts.Color}
	width := opts.Width
	if width <= 0 {
		width = 100
	}
	now := opts.Now
	if now.IsZero() {
		now = time.Now()
	}

	var sb strings.Builder
	line := func(format string, args ...any) { fmt.Fprintf(&sb, format+"\n", args...) }
	heading := func(title string, count int) {
		line("")
		line("%s %s", p.paint(styleBold+styleCyan, title), p.paint(styleDim, fmt.Sprintf("(%d)", count)))
	}

	line("%s %s  pid %d  up %s  %d ops",
		p.paint(styleBold, "system-control daemon"), p.paint(styleGreen, "● running"),
		info.PID, formatDuration(now.Sub(info.StartedAt)), info.Ops)
	line("%s %s", p.paint(styleDim, "socket"), info.Socket)

	// watchers
	heading("Watchers", len(info.Watches))
	if len(info.Watches) == 0 {
		line("  %s", p.paint(styleDim, "nothing is watched"))
	} else {
		opCol, argCol := len("OP"), len("ARGS")
		for _, w := range info.Watches {
			opCol = max(opCol, utf8.RuneCountInString(w.Op))
			argCol = max(argCol, utf8.RuneCountInString(w.Args))
		}
		argCol = min(argCol, max(width/3, 12))
		valueCol := max(width-2-opCol-2-argCol-2-4-2, 8)

		line("  %s  %s  %s  %s", p.cell(styleDim, "OP", opCol, false), p.cell(styleDim, "ARGS", argCol, false),
			p.cell(styleDim, "SUBS", 4, true), p.paint(styleDim, "VALUE"))
		for _, w := range info.Watches {
			value, style := strings.Join(w.Value, " ⏎ "), ""
			if w.Error != "" {
				value, style = "error: "+w.Error, styleRed
			}
			line("  %s  %s  %s  %s", p.cell(styleBold, w.Op, opCol, false), p.cell("", w.Args, argCol, false),
				p.cell("", fmt.Sprint(w.Subscribers), 4, true), p.paint(style, truncate(value, valueCol)))
		}
	}

	// executions
	heading("Executions", len(info.Executions))
	if len(info.Executions) == 0 {
		line("  %s", p.paint(styleDim, "no command was run yet"))
	} else {
		opCol := len("OP")
		for _, x := range info.Executions {
			opCol = max(opCol, utf8.RuneCountInString(x.Op))
		}
		line("  %s  %s  %s  %s  %s  %s  %s", p.cell(styleDim, "OP", opCol, false),
			p.cell(styleDim, "CALLS", 6, true), p.cell(styleDim, "AVG", 8, true), p.cell(styleDim, "MAX", 8, true),
			p.cell(styleDim, "DROPPED", 7, true), p.cell(styleDim, "FAILED", 6, true), p.cell(styleDim, "FALLBACK", 8, true))

		count := func(n int, style string, w int) string {
			if n == 0 {
				return p.cell(styleDim, "·", w, true)
			}
			return p.cell(style, fmt.Sprint(n), w, true)
		}
		for _, x := range info.Executions {
			line("  %s  %s  %s  %s  %s  %s  %s", p.cell(styleBold, x.Op, opCol, false),
				p.cell("", fmt.Sprint(x.Calls), 6, true),
				p.cell("", formatDuration(x.Avg), 8, true), p.cell("", formatDuration(x.Max), 8, true),
				count(x.Dropped, styleYellow, 7), count(x.Failed, styleRed, 6), count(x.Unhandled, "", 8))
			if x.LastError != "" {
				ago := formatDuration(now.Sub(x.LastErrorAt))
				line("  %s  %s", strings.Repeat(" ", opCol),
					p.paint(styleRed, truncate(fmt.Sprintf("last error %s ago: %s", ago, x.LastError), max(width-opCol-4, 20))))
			}
		}
	}

	if len(info.Details) > 0 {
		heading("Sources", len(info.Details))
		for _, d := range info.Details {
			line("  %s", d)
		}
	}
	return sb.String()
}
