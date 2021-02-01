package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"mercury/app/logic/api"
	"mercury/app/logic/service"
	"mercury/config"
)

func NewClientCommand(f Factory) *cobra.Command {
	o := ClientOptions{}
	cmd := &cobra.Command{
		Use:   "client",
		Short: "",
		Long:  ``,
		Annotations: map[string]string{
			"group": "client",
		},
	}

	token := &cobra.Command{
		Use:   "token",
		Short: "",
		Long:  ``,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := o.Complete(f, args); err != nil {
				return err
			}
			return o.Token()
		},
	}

	token.Flags().StringVar(&o.ClientID, "client_id", "", "the id of client")
	token.Flags().StringVar(&o.ClientSecret, "client_secret", "", "the secret of client")
	_ = token.MarkFlagRequired("client_id")
	_ = token.MarkFlagRequired("client_secret")

	pullMessage := &cobra.Command{
		Use:   "pull_message",
		Short: "",
		Long:  ``,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := o.Complete(f, args); err != nil {
				return err
			}
			return o.PullMessage()
		},
	}

	cmd.AddCommand(token, pullMessage)
	return cmd
}

// UserOptions encapsulates state for the user command & subcommands
type ClientOptions struct {
	ClientID     string
	ClientSecret string
	Service      service.Servicer

	args []string
}

// Complete adds any missing configuration that can only be added just before calling Run
func (o *ClientOptions) Complete(f Factory, args []string) (err error) {
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

func (o *ClientOptions) Token() error {
	token, lifetime, err := o.Service.GenerateToken(context.Background(), &api.GenerateTokenReq{
		ClientID:     o.ClientID,
		ClientSecret: o.ClientSecret,
	})
	if err != nil {
		return err
	}

	fmt.Printf("token: %s, lifetime: %s \n", token, lifetime)
	return nil
}

func (o *ClientOptions) PullMessage() error {
	messages, err := o.Service.PullMessage(context.Background(), &api.PullMessageReq{
		UID: o.args[0],
	})
	if err != nil {
		return err
	}

	for _, message := range messages {
		fmt.Printf("message topic: %s, count: %d \n", message.Topic, message.Count)
		for _, m := range message.Messages {
			fmt.Printf("message: %+v \n", *m)
		}
	}
	return nil
}
