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
	//ctx := context.Background()
	image := args[0]

	namedRef, err := reference.Parse(image)
	if err != nil {
		return err
	}
	namedRef = reference.TrimTagForDigest(reference.WithDefaultTagIfMissing(namedRef))

	responseBody, err := apiClient.ImagePush(context.TODO(), namedRef.String(), fetchRegistryAuth(namedRef.Name()))
	if err != nil {
		return fmt.Errorf("failed to push image: %v", err)
	}
	defer responseBody.Close()

	return jsonstream.DisplayJSONMessagesToStream(responseBody)
}

func pushExample() string {
	return `$ pouch push docker.io/library/busybox:1.25
manifest-sha256:29f5d56d12684887bdfa50dcd29fc31eea4aaf4ad3bec43daf19026a7ce69912: done           |++++++++++++++++++++++++++++++++++++++|
layer-sha256:56bec22e355981d8ba0878c6c2f23b21f422f30ab0aba188b54f1ffeff59c190:    done           |++++++++++++++++++++++++++++++++++++++|
config-sha256:e02e811dd08fd49e7f6032625495118e63f597eb150403d02e3238af1df240ba:   done           |++++++++++++++++++++++++++++++++++++++|
elapsed: 10.5s                                                                    total:   0.0 B (0.0 B/s)
`
}

