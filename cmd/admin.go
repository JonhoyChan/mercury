package cmd

import (
	"context"
	"errors"
	"github.com/spf13/cobra"
	"mercury/lib"
)

func NewAdminCommand(f Factory) *cobra.Command {
	o := AdminOptions{}
	cmd := &cobra.Command{
		Use:   "admin",
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

// AdminOptions encapsulates state for the admin command
type AdminOptions struct {
	Setup       bool
	AdminServer *lib.AdminServer
}

// Complete adds any missing configuration that can only be added just before calling Run
func (o *AdminOptions) Complete(f Factory, args []string) (err error) {
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

	o.AdminServer, err = f.AdminServer()
	if err != nil {
		return
	}
	return
}

// Run executes the infra command with currently configured state
func (o *AdminOptions) Run() error {
	ctx := context.Background()
	return o.AdminServer.Serve(ctx)
}
