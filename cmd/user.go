package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"mercury/app/logic/service"
	"mercury/config"
	"mercury/x/ecode"
)

func NewUserCommand(f Factory) *cobra.Command {
	o := UserOptions{}
	cmd := &cobra.Command{
		Use:   "user",
		Short: "",
		Long:  ``,
		Annotations: map[string]string{
			"group": "user",
		},
	}

	token := &cobra.Command{
		Use:   "token",
		Short: "",
		Long:  ``,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := o.Complete(f, args); err != nil {
				return err
			}
			return o.Token()
		},
	}

	token.Flags().StringVar(&o.ClientToken, "client_token", "", "the token of client")
	_ = token.MarkFlagRequired("client_token")

	cmd.AddCommand(token)
	return cmd
}

// UserOptions encapsulates state for the user command & subcommands
type UserOptions struct {
	ClientToken string
	Service     service.Servicer

	args []string
}

// Complete adds any missing configuration that can only be added just before calling Run
func (o *UserOptions) Complete(f Factory, args []string) (err error) {
	if err = f.Init(); err != nil {
		return
	}

	var cfg *config.Config
	cfg, err = f.Config()
	if err != nil {
		return
	}

	if o.Service, err = service.NewService(config.NewProviderConfig(cfg), f.Logger().New("service", "mercury.logic")); err != nil {
		return err
	}
	o.args = args
	return
}

func (o *UserOptions) Token() error {
	var clientID string
	_, err := o.Service.Authenticate(o.ClientToken, &clientID)
	if err != nil {
		return ecode.ErrInvalidToken
	}

	ctx := service.ContextWithClientID(context.Background(), clientID)
	token, lifetime, err := o.Service.GenerateUserToken(ctx, o.args[0])
	if err != nil {
		return err
	}

	fmt.Printf("token: %s, lifetime: %s \n", token, lifetime)
	return nil
}
