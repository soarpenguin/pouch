package main

import (
	"context"
	"fmt"

	"github.com/alibaba/pouch/pkg/jsonstream"
	"github.com/alibaba/pouch/pkg/reference"
	"github.com/spf13/cobra"
)

// pushDescription is used to describe push command in detail and auto generate command doc.
var pushDescription = "Push an image or a repository to a registry."

// PushCommand use to implement 'push' command, it push a image
type PushCommand struct {
	baseCommand
}

// Init initialize push command.
func (p *PushCommand) Init(c *Cli) {
	p.cli = c
	p.cmd = &cobra.Command{
		Use:   "push [OPTIONS] NAME[:TAG]",
		Short: "Push an image or a repository to a registry",
		Long:  pushDescription,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.runPush(args)
		},
		Example: pushExample(),
	}
	p.addFlags()
}

// addFlags adds flags for specific command.
func (p *PushCommand) addFlags() {
	// TODO: add flags here
}

// runPush is the entry of push command.
func (p *PushCommand) runPush(args []string) error {
	apiClient := p.cli.Client()
	ctx := context.Background()
	image := args[0]

	namedRef, err := reference.Parse(image)
	if err != nil {
		return err
	}
	namedRef = reference.WithDefaultTagIfMissing(namedRef)

	responseBody, err := apiClient.ImagePush(ctx, image, fetchRegistryAuth(namedRef.Name()))
	if err != nil {
		return fmt.Errorf("failed to push image: %v", err)
	}
	defer responseBody.Close()

	return jsonstream.DisplayJSONMessagesToStream(responseBody)
}

func pushExample() string {
	return `$ pouch push docker.io/library/redis:alpine
manifest-sha256:300a28bded4df40cd59c7f3711b9f0fd4fc94100d07c3c1eeabdd449773cfdcf: done |++++++++++++++++++++++++++++++++++++++|
layer-sha256:4d017670a09e6e7d537a5be1d34e54aaef884ef5b9996b7c48f68d310e431b01:    done |++++++++++++++++++++++++++++++++++++++|
config-sha256:11d89718732da30fa851e37f30d164e4154ad10d4024c00aec0fd7ce4fd613cf:   done |++++++++++++++++++++++++++++++++++++++|
layer-sha256:11116c137da88158b07bdcecd6629b3f7ff98f80d82bf512cdcabcdcb4370e59:    done |++++++++++++++++++++++++++++++++++++++|
layer-sha256:0653bff3c5cf23727e0ebceae7a28f7534ab64ed13966e080e4c9b035176c401:    done |++++++++++++++++++++++++++++++++++++++|
elapsed: 0.9 s                                                                    total:  3.0 Ki (3.3 KiB/s)
`
}

