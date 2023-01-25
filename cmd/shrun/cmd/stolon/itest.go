package stolon

import (
	"fmt"

	"github.com/docker/docker/client"
	"github.com/spf13/cobra"

	"github.com/wmentor/shrun/cmd"
	"github.com/wmentor/shrun/internal/cases/stolon/itest"
	"github.com/wmentor/shrun/internal/container"
	"github.com/wmentor/shrun/internal/image"
)

var (
	_ cmd.CobraCommand = (*CommandITest)(nil)
)

type CommandITest struct {
	command *cobra.Command
	cli     *client.Client
	tests   []string
}

func NewCommandITest(cli *client.Client) *CommandITest {
	ci := &CommandITest{
		cli: cli,
	}

	cc := &cobra.Command{
		Use:   "itest",
		Short: "Run stolon integration tests",
		RunE:  ci.exec,
	}

	cc.Flags().StringSliceVarP(&ci.tests, "test", "t", []string{}, "test names to run")

	ci.command = cc

	return ci
}

func (ci *CommandITest) Command() *cobra.Command {
	return ci.command
}

func (ci *CommandITest) exec(cc *cobra.Command, _ []string) error {
	ctx := cc.Context()

	mng, err := container.NewManager(ci.cli)
	if err != nil {
		return fmt.Errorf("create container manager error: %w", err)
	}

	builder, err := image.NewManager(ci.cli)
	if err != nil {
		return fmt.Errorf("create image manager error: %w", err)
	}

	myCase, err := itest.NewCase(builder, mng)
	if err != nil {
		return fmt.Errorf("create case object error: %w", err)
	}

	return myCase.WithTests(ci.tests).Exec(ctx)
}
