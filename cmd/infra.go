package cmd

import (
	"context"
	"errors"
	"github.com/spf13/cobra"
	"mercury/lib"
)

func NewInfraCommand(f Factory) *cobra.Command {
	o := InfraOptions{}
	cmd := &cobra.Command{
		Use:   "infra",
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

	cmd.Flags().BoolVarP(&o.Setup, "setup", "", false, "run setup if necessary, reading options from environment variables")

	return cmd
}

// InfraOptions encapsulates state for the infra command
type InfraOptions struct {
	Setup       bool
	InfraServer *lib.InfraServer
}

// Complete adds any missing configuration that can only be added just before calling Run
func (o *InfraOptions) Complete(f Factory, args []string) (err error) {
	repoErr := lib.RepoExists(f.RepoPath())
	if errors.Is(repoErr, lib.ErrNoRepo) {
		p := lib.SetupParams{
			RepoPath: f.RepoPath(),
		}
		if err = lib.Setup(p); err != nil {
			return
		}
	} else if repoErr != nil {
		return repoErr
	}

	o.InfraServer, err = f.InfraServer()
	if err != nil {
		return
	}
	return
}

// Run executes the infra command with currently configured state
func (o *InfraOptions) Run() error {
	ctx := context.Background()
	return o.InfraServer.Serve(ctx)
}
