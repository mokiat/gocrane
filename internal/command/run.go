package command

import (
	"github.com/urfave/cli/v2"

	"github.com/mokiat/gocrane/internal/crane"
)

func Run() *cli.Command {
	return &cli.Command{
		Name: "run",
		Action: func(c *cli.Context) error {
			settings, err := newRunSettings(c)
			if err != nil {
				return err
			}
			return crane.Run(c.Context, settings)
		},
	}
}
