package cmd

import (
	"context"
	"github.com/spf13/cobra"
	"mercury/lib"
)

func NewLogicCommand(f Factory) *cobra.Command {
	o := LogicOptions{}
	cmd := &cobra.Command{
		Use:   "logic",
		Short: "",
		Annotations: map[string]string{
			"group": "server",
		},
		Long: ``,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := o.Complete(f, args); err != nil {
				return err
			}
			return o.Run()
		},
	}

	return cmd
}

// LogicOptions encapsulates state for the logic command
type LogicOptions struct {
	Setup       bool
	LogicServer *lib.LogicServer
}

// Complete adds any missing configuration that can only be added just before calling Run
func (o *LogicOptions) Complete(f Factory, args []string) (err error) {
	o.LogicServer, err = f.LogicServer()
	if err != nil {
		return
	}
	return
}

// Run executes the infra command with currently configured state
func (o *LogicOptions) Run() error {
	ctx := context.Background()
	return o.LogicServer.Serve(ctx)
}
