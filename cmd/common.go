package cmd

import (
	"github.com/spf13/cobra"
)

type CobraCommand interface {
	Command() *cobra.Command
}
