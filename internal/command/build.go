package command

import (
	"github.com/urfave/cli/v2"

	"github.com/mokiat/gocrane/internal/crane"
)

func Build() *cli.Command {
	return &cli.Command{
		Name: "build",
		Action: func(c *cli.Context) error {
			settings, err := newBuildSettings(c)
			if err != nil {
				return err
			}
			return crane.Build(c.Context, settings)
		},
	}
}
