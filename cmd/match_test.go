package cmd

import (
	"fmt"
	"io"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/markusressel/system-control/internal/daemon"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// The daemon parses command lines itself. Whatever it accepts must be understood by cobra
// in exactly the same way, or a command would do something else than without a daemon.
func TestDaemonParsesLikeCobra(t *testing.T) {
	if err := daemon.Default.Bind(RootCmd); err != nil {
		t.Fatal(err)
	}

	rng := rand.New(rand.NewSource(1))
	accepted := 0

	for _, op := range daemon.Default.Ops() {
		path := strings.Split(op.Name, ".")
		command, _, err := RootCmd.Find(path)
		if err != nil || command == RootCmd {
			t.Fatalf("op %s has no command", op.Name)
		}

		for i := 0; i < 300; i++ {
			rest := randomArguments(rng, command)
			argv := append(append([]string{}, path...), rest...)

			matched, got, ok := daemon.Default.MatchArgv(argv)
			if !ok {
				continue
			}
			if matched.Name != op.Name {
				// the first positional argument selected another command
				continue
			}
			accepted++

			want, err := parseWithCobra(t, command, rest)
			if err != nil {
				t.Errorf("%v: accepted by the daemon, rejected by cobra: %v", argv, err)
				continue
			}

			if !reflect.DeepEqual(got.Pos, want.Pos) && !(len(got.Pos) == 0 && len(want.Pos) == 0) {
				t.Errorf("%v: positional arguments differ: daemon %q, cobra %q", argv, got.Pos, want.Pos)
			}
			if got.Watch != want.Watch {
				t.Errorf("%v: watch differs: daemon %v, cobra %v", argv, got.Watch, want.Watch)
			}
			for name, wantValue := range want.Flags {
				gotValue := got.Flags[name]
				if f := command.Flags().Lookup(name); f != nil && f.Value.Type() == "float64" {
					// "1.0" and "1" are the same number, ops parse the text
					gotNumber, gotErr := strconv.ParseFloat(gotValue, 64)
					wantNumber, wantErr := strconv.ParseFloat(wantValue, 64)
					if gotErr == nil && wantErr == nil && gotNumber == wantNumber {
						continue
					}
				}
				if gotValue != wantValue {
					t.Errorf("%v: flag %s differs: daemon %q, cobra %q", argv, name, gotValue, wantValue)
				}
			}
		}
	}

	if accepted < 100 {
		t.Errorf("only %d generated command lines were accepted, the generator is too narrow", accepted)
	}
}

// randomArguments builds a command line out of the flags of a command in all spellings.
func randomArguments(rng *rand.Rand, command *cobra.Command) []string {
	var flags []*pflag.Flag
	command.Flags().VisitAll(func(f *pflag.Flag) {
		if f.Name != "help" {
			flags = append(flags, f)
		}
	})

	var args []string
	for n := rng.Intn(4); n > 0 && len(flags) > 0; n-- {
		f := flags[rng.Intn(len(flags))]
		value := sampleValue(rng, f)
		spelling := rng.Intn(4)
		if f.Shorthand == "" && spelling >= 2 {
			spelling = rng.Intn(2)
		}

		switch {
		case f.Value.Type() == "bool" && spelling%2 == 0:
			args = append(args, "--"+f.Name)
		case f.Value.Type() == "bool" && f.Shorthand != "":
			args = append(args, "-"+f.Shorthand)
		case spelling == 0:
			args = append(args, "--"+f.Name, value)
		case spelling == 1:
			args = append(args, "--"+f.Name+"="+value)
		case spelling == 2:
			args = append(args, "-"+f.Shorthand, value)
		default:
			args = append(args, "-"+f.Shorthand+value)
		}
	}

	for n := rng.Intn(3); n > 0; n-- {
		position := rng.Intn(len(args) + 1)
		args = append(args[:position], append([]string{fmt.Sprintf("p%d", n)}, args[position:]...)...)
	}
	if rng.Intn(10) == 0 {
		args = append(args, "--", "-x")
	}
	return args
}

func sampleValue(rng *rand.Rand, f *pflag.Flag) string {
	switch f.Value.Type() {
	case "int", "int64":
		return fmt.Sprint(rng.Intn(100))
	case "float64":
		return fmt.Sprintf("%.1f", rng.Float64())
	case "stringSlice":
		return []string{"a", "a,b", "c"}[rng.Intn(3)]
	case "bool":
		return []string{"true", "false"}[rng.Intn(2)]
	default:
		return []string{"x", "two words", "5"}[rng.Intn(3)]
	}
}

// parseWithCobra parses the arguments with the flags of a command like cobra does, on a
// fresh flag set so that one parse can't influence the next, and returns them as op arguments.
func parseWithCobra(t *testing.T, command *cobra.Command, rest []string) (daemon.Args, error) {
	t.Helper()
	fs := pflag.NewFlagSet(command.Name(), pflag.ContinueOnError)
	fs.SetOutput(io.Discard)

	command.Flags().VisitAll(func(f *pflag.Flag) {
		switch f.Value.Type() {
		case "string":
			fs.StringP(f.Name, f.Shorthand, f.DefValue, "")
		case "int":
			fs.IntP(f.Name, f.Shorthand, 0, "")
			_ = fs.Set(f.Name, f.DefValue)
		case "int64":
			fs.Int64P(f.Name, f.Shorthand, 0, "")
			_ = fs.Set(f.Name, f.DefValue)
		case "float64":
			fs.Float64P(f.Name, f.Shorthand, 0, "")
			_ = fs.Set(f.Name, f.DefValue)
		case "bool":
			fs.BoolP(f.Name, f.Shorthand, f.DefValue == "true", "")
		case "stringSlice":
			var defaults []string
			if def := strings.Trim(f.DefValue, "[]"); def != "" {
				defaults = strings.Split(def, ",")
			}
			fs.StringSliceP(f.Name, f.Shorthand, defaults, "")
		default:
			t.Fatalf("flag --%s has type %s, extend the test", f.Name, f.Value.Type())
		}
	})

	if err := fs.Parse(rest); err != nil {
		return daemon.Args{}, err
	}
	positional := fs.Args()
	if command.Args != nil {
		if err := command.Args(command, positional); err != nil {
			return daemon.Args{}, err
		}
	}

	a := daemon.Args{Flags: map[string]string{}, Pos: positional}
	a.Watch, _ = fs.GetBool("watch")
	fs.VisitAll(func(f *pflag.Flag) {
		if f.Name == "watch" {
			return
		}
		if f.Value.Type() == "stringSlice" {
			values, _ := fs.GetStringSlice(f.Name)
			a.Flags[f.Name] = strings.Join(values, ",")
			return
		}
		a.Flags[f.Name] = f.Value.String()
	})
	return a, nil
}
