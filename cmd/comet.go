package cmd

import (
	"context"
	"github.com/spf13/cobra"
	"mercury/lib"
)

func NewCometCommand(f Factory) *cobra.Command {
	o := CometOptions{}
	cmd := &cobra.Command{
		Use:   "comet",
		Short: "",
		Long:  ``,
		Annotations: map[string]string{
			"group": "server",
		},
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

// CometOptions encapsulates state for the comet command
type CometOptions struct {
	Setup       bool
	CometServer *lib.CometServer
}

// Complete adds any missing configuration that can only be added just before calling Run
func (o *CometOptions) Complete(f Factory, args []string) (err error) {
	o.CometServer, err = f.CometServer()
	if err != nil {
		return
	}
	return
}

// Run executes the infra command with currently configured state
func (o *CometOptions) Run() error {
	ctx := context.Background()
	return o.CometServer.Serve(ctx)
}
