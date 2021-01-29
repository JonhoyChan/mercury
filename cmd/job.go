package cmd

import (
	"context"
	"github.com/spf13/cobra"
	"mercury/lib"
)

func NewJobCommand(f Factory) *cobra.Command {
	o := JobOptions{}
	cmd := &cobra.Command{
		Use:   "job",
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

// JobOptions encapsulates state for the job command
type JobOptions struct {
	Setup     bool
	JobServer *lib.JobServer
}

// Complete adds any missing configuration that can only be added just before calling Run
func (o *JobOptions) Complete(f Factory, args []string) (err error) {
	o.JobServer, err = f.JobServer()
	if err != nil {
		return
	}
	return
}

// Run executes the infra command with currently configured state
func (o *JobOptions) Run() error {
	ctx := context.Background()
	return o.JobServer.Serve(ctx)
}
