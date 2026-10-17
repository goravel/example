package commands

import (
	"errors"
	"fmt"
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/support/color"
	supportconsole "github.com/goravel/framework/support/console"
)

type ConsoleColorCommand struct{}

func (r *ConsoleColorCommand) Signature() string {
	return "test:console-color"
}

func (r *ConsoleColorCommand) Description() string {
	return "Verify color tag rendering through gookit/color (framework#1571)"
}

func (r *ConsoleColorCommand) Extend() command.Extend {
	return command.Extend{Category: "console"}
}

const (
	// redTestText is the plain text the red test tag must render to.
	redTestText = "test"
	// redTestTag is the red tag used by the probe and the sample rows.
	redTestTag = "<red>" + redTestText + "</>"
)

// tagEscaper swaps tag delimiters for lookalikes in display copies only, so
// ctx.Line prints them literally instead of rendering or clearing them.
var tagEscaper = strings.NewReplacer("<", "‹", ">", "›")

// colorEnabled reports the process-global color state. The framework exposes no
// getter, so probe a tag: gookit clears tags when disabled and emits ANSI when
// enabled.
func colorEnabled() bool {
	return color.Sprint(redTestTag) != redTestText
}

func escapeTags(s string) string {
	return tagEscaper.Replace(s)
}

// Handle prints every color-tag case in labelled sections so one run verifies
// framework#1571 by hand: prefix printers, tag forms, StripColor,
// TwoColumnDetail width measurement, and the Disable/Enable toggle. Each case
// shows a visual render, and most add the exact rendered bytes with %q. Where
// the unrendered form matters, tag delimiters are shown as ‹ › because ctx.Line
// would otherwise render or clear them.
//
// Whether color is on is detected with a probe, since the framework exposes no
// getter. When the probe says color is off at entry, Handle forces color on,
// re-probes, and restores the entry state before printing anything: a tag that
// is still cleared means the renderer is broken, so it reports framework#1571
// as an error, while a tag that renders once forced on means color was disabled
// by policy or by a latched --no-ansi, which is only reported and returns nil.
// The process-global color state is snapshotted first and restored on return,
// because the toggle section changes it.
func (r *ConsoleColorCommand) Handle(ctx console.Context) error {
	wasEnabled := colorEnabled()
	defer func() {
		if wasEnabled {
			color.Enable()
		} else {
			color.Disable()
		}
	}()

	if !wasEnabled {
		// Force color on to tell a policy/latch disable (for example --no-ansi)
		// apart from a renderer that is actually broken, then restore the entry
		// state so this branch never prints ANSI.
		color.Enable()
		broken := color.Sprint(redTestTag) == redTestText
		color.Disable()

		if broken {
			ctx.Error("FAIL: tags are still cleared after color.Enable(); rendering is broken (framework#1571).")
			ctx.Line(fmt.Sprintf("   input %q -> %q (delimiters shown as ‹ ›)",
				escapeTags(redTestTag), color.Sprint(redTestTag)))
			return errors.New("tag rendering is off even after color.Enable() (framework#1571)")
		}

		ctx.Line("Color rendering is disabled by --no-ansi or a prior --no-ansi run; " +
			"re-run in a fresh process to verify framework#1571.")
		ctx.Line(fmt.Sprintf("   input %q -> %q (delimiters shown as ‹ ›; tag cleared, no ANSI)",
			escapeTags(redTestTag), color.Sprint(redTestTag)))
		return nil
	}

	ctx.Line("<fg=yellow;op=bold>Color tag rendering check (framework#1571)</>")
	ctx.NewLine()

	// 1. Prefix printers render tags through gookit as well; the rendered output
	// is the evidence, so this section stays visual.
	ctx.Line("<fg=cyan>1. Prefix printers</>")
	ctx.Info("<fg=green>Application is healthy</>")
	ctx.Warning("<fg=yellow>Queue worker is not running</>")
	ctx.Success("<green>Migration finished</>")
	ctx.Error("<fg=red>Cache connection failed</>")
	ctx.Comment("<fg=gray>Checking cache store</>")
	ctx.NewLine()

	// 2. Tag forms, each rendered visually and then as exact bytes.
	ctx.Line("<fg=cyan>2. Tag forms (visual, then exact rendered bytes)</>")
	ctx.Line(fmt.Sprintf("   style tag (red): %s", redTestTag))
	ctx.Line(fmt.Sprintf("   style tag bytes: %q", color.Sprint(redTestTag)))
	ctx.Line("   fg attribute (green): <fg=green>ok</>")
	ctx.Line(fmt.Sprintf("   fg attribute bytes: %q", color.Sprint("<fg=green>ok</>")))
	ctx.Line("   bg attribute (blue): <bg=blue>MSG</>")
	ctx.Line(fmt.Sprintf("   bg attribute bytes (Sprintf): %q", color.Sprintf("<bg=blue>%s</>", "MSG")))
	ctx.Line(fmt.Sprintf("   nested (outer red, inner green): %s", color.Red().Sprint("outer <fg=green>inner</> end")))
	ctx.Line(fmt.Sprintf("   nested bytes: %q", color.Red().Sprint("outer <fg=green>inner</> end")))
	ctx.NewLine()

	// 3. StripColor consumes tags and ANSI alike, including from mixed input.
	// Display copies replace the tag delimiters with lookalikes so the raw syntax
	// survives ctx.Line; StripColor still receives the real values.
	ctx.Line("<fg=cyan>3. color.StripColor (tag delimiters shown as ‹ › -> stripped result)</>")
	ctx.Line(fmt.Sprintf("   tagged input:      %q -> %q", escapeTags(redTestTag), color.StripColor(redTestTag)))
	raw := "\x1b[1mplain\x1b[0m"
	ctx.Line(fmt.Sprintf("   raw ANSI input:    %q -> %q", raw, color.StripColor(raw)))
	mixed := color.Green().Sprint("a ") + "<fg=yellow>b</>"
	ctx.Line(fmt.Sprintf("   mixed ANSI + tag:  %q -> %q", escapeTags(mixed), color.StripColor(mixed)))
	ctx.NewLine()

	// 4. TwoColumnDetail measures rendered width, so the tagged second column
	// lines up with the pre-colored one and with plain text. Color is guaranteed
	// enabled here, so both the rendered bytes and the stripped layout must agree.
	ctx.Line("<fg=cyan>4. TwoColumnDetail width measurement</>")
	taggedDetail := supportconsole.TwoColumnDetail("<fg=cyan>cache</>", "<fg=green>OK</>")
	coloredDetail := supportconsole.TwoColumnDetail(color.Cyan().Sprint("cache"), color.Green().Sprint("OK"))
	plainDetail := supportconsole.TwoColumnDetail("cache", "OK")
	ctx.Line(fmt.Sprintf("   tagged:   %s", taggedDetail))
	ctx.Line(fmt.Sprintf("   colored:  %s", coloredDetail))
	ctx.Line(fmt.Sprintf("   plain:    %s", plainDetail))
	match := color.Sprint(taggedDetail) == coloredDetail &&
		color.StripColor(taggedDetail) == color.StripColor(plainDetail)
	ctx.Line(fmt.Sprintf("   layout match: %t", match))
	ctx.NewLine()

	// 5. Disable/Enable toggle. Kept last so the explicit Enable() cannot change
	// how the earlier sections render; the deferred restore fixes the global
	// state again on return. This only runs when color was enabled at entry.
	ctx.Line("<fg=cyan>5. color.Disable() and color.Enable() toggle</>")
	color.Disable()
	ctx.Line(fmt.Sprintf("   disabled bytes:  %q", color.Sprint(redTestTag)))
	ctx.Line(fmt.Sprintf("   disabled visual: %s", redTestTag))
	color.Enable()
	ctx.Line(fmt.Sprintf("   enabled bytes:   %q", color.Sprint(redTestTag)))
	ctx.Line(fmt.Sprintf("   enabled visual:  %s", redTestTag))

	return nil
}
