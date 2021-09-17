package cliu

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"os"
)

type MarkdownCmd struct {
	*cli.Command
	file *cli.StringFlag
}

func NewMarkdownCmd(app *cli.App) (ret *MarkdownCmd) {
	targetFile := newTargetFileFlag(fmt.Sprintf("%v.md", app.Name))
	ret = &MarkdownCmd{
		Command: &cli.Command{
			Name:  "markdown",
			Usage: "Generate markdown help file",
			Flags: []cli.Flag{
				targetFile,
			},
			Action: func(c *cli.Context) (err error) {
				var markdown string
				if markdown, err = app.ToMarkdown(); err == nil {
					err = os.WriteFile(targetFile.CurrentValue, []byte(markdown), 0777)
				}
				return
			},
		},
	}

	return
}

func newTargetFileFlag(name string) (ret *StringFlag) {
	ret = NewStringFlag(&cli.StringFlag{
		Name:  "targetFile",
		Usage: "The path file to generate.",
		Value: name,
	})
	return
}
