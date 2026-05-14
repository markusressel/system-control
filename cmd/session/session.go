package session

import "github.com/spf13/cobra"

var Command = &cobra.Command{
	Use:              "session",
	Short:            "Control desktop session state",
	Long:             ``,
	TraverseChildren: true,
}
