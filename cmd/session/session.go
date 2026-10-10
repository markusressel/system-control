package session

import "github.com/spf13/cobra"

var (
	sessionUser    string
	sessionDisplay string
)

var Command = &cobra.Command{
	Use:              "session",
	Short:            "Control desktop session state",
	Long:             ``,
	TraverseChildren: true,
}

func init() {
	Command.PersistentFlags().StringVar(&sessionUser, "user", "", "User that owns the session (default: $SUDO_USER, else current user)")
	Command.PersistentFlags().StringVar(&sessionDisplay, "display", "", "X display of the session (default: $DISPLAY, else :0)")
}
