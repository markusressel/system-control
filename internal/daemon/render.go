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
	r := statusRenderer{painter: painter{opts.Color}, width: opts.Width, now: opts.Now}
	if r.width <= 0 {
		r.width = 100
	}
	if r.now.IsZero() {
		r.now = time.Now()
	}

	r.line("%s %s  pid %d  up %s  %d ops",
		r.paint(styleBold, "system-control daemon"), r.paint(styleGreen, "● running"),
		info.PID, formatDuration(r.now.Sub(info.StartedAt)), info.Ops)
	r.line("%s %s", r.paint(styleDim, "socket"), info.Socket)
	r.watchers(info.Watches)
	r.executions(info.Executions)
	if len(info.Details) > 0 {
		r.heading("Sources", len(info.Details))
		for _, d := range info.Details {
			r.line("  %s", d)
		}
	}
	return r.sb.String()
}

// statusRenderer collects the lines of RenderStatus.
type statusRenderer struct {
	painter
	width int
	now   time.Time
	sb    strings.Builder
}

func (r *statusRenderer) line(format string, args ...any) {
	fmt.Fprintf(&r.sb, format+"\n", args...)
}

func (r *statusRenderer) heading(title string, count int) {
	r.line("")
	r.line("%s %s", r.paint(styleBold+styleCyan, title), r.paint(styleDim, fmt.Sprintf("(%d)", count)))
}

func (r *statusRenderer) watchers(watches []WatchStatus) {
	r.heading("Watchers", len(watches))
	if len(watches) == 0 {
		r.line("  %s", r.paint(styleDim, "nothing is watched"))
		return
	}

	opCol, argCol := len("OP"), len("ARGS")
	for _, w := range watches {
		opCol = max(opCol, utf8.RuneCountInString(w.Op))
		argCol = max(argCol, utf8.RuneCountInString(w.Args))
	}
	argCol = min(argCol, max(r.width/3, 12))
	valueCol := max(r.width-2-opCol-2-argCol-2-4-2, 8)

	r.line("  %s  %s  %s  %s", r.cell(styleDim, "OP", opCol, false), r.cell(styleDim, "ARGS", argCol, false),
		r.cell(styleDim, "SUBS", 4, true), r.paint(styleDim, "VALUE"))
	for _, w := range watches {
		value, style := strings.Join(w.Value, " ⏎ "), ""
		if w.Error != "" {
			value, style = "error: "+w.Error, styleRed
		}
		r.line("  %s  %s  %s  %s", r.cell(styleBold, w.Op, opCol, false), r.cell("", w.Args, argCol, false),
			r.cell("", fmt.Sprint(w.Subscribers), 4, true), r.paint(style, truncate(value, valueCol)))
	}
}

func (r *statusRenderer) executions(executions []ExecStatus) {
	r.heading("Executions", len(executions))
	if len(executions) == 0 {
		r.line("  %s", r.paint(styleDim, "no command was run yet"))
		return
	}

	opCol := len("OP")
	for _, x := range executions {
		opCol = max(opCol, utf8.RuneCountInString(x.Op))
	}
	r.line("  %s  %s  %s  %s  %s  %s  %s", r.cell(styleDim, "OP", opCol, false),
		r.cell(styleDim, "CALLS", 6, true), r.cell(styleDim, "AVG", 8, true), r.cell(styleDim, "MAX", 8, true),
		r.cell(styleDim, "DROPPED", 7, true), r.cell(styleDim, "FAILED", 6, true), r.cell(styleDim, "FALLBACK", 8, true))
	for _, x := range executions {
		r.line("  %s  %s  %s  %s  %s  %s  %s", r.cell(styleBold, x.Op, opCol, false),
			r.cell("", fmt.Sprint(x.Calls), 6, true),
			r.cell("", formatDuration(x.Avg), 8, true), r.cell("", formatDuration(x.Max), 8, true),
			r.count(x.Dropped, styleYellow, 7), r.count(x.Failed, styleRed, 6), r.count(x.Unhandled, "", 8))
		if x.LastError != "" {
			ago := formatDuration(r.now.Sub(x.LastErrorAt))
			r.line("  %s  %s", strings.Repeat(" ", opCol),
				r.paint(styleRed, truncate(fmt.Sprintf("last error %s ago: %s", ago, x.LastError), max(r.width-opCol-4, 20))))
		}
	}
}

// count is a right aligned number cell, a dim dot for 0.
func (r *statusRenderer) count(n int, style string, width int) string {
	if n == 0 {
		return r.cell(styleDim, "·", width, true)
	}
	return r.cell(style, fmt.Sprint(n), width, true)
}
