package cli

import (
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/nproc/run/build"
	"github.com/nproc/run/text"
	"github.com/nproc/run/valuesloader"
	"github.com/urfave/cli"
)

const (
	description = "Compile config templates based on environment variables" +
		" and run a command after the template is successfully compiled." +
		"\n   The compiled file will have the permissions set to 0777." +
		"\n   Check the projects page for the documentation and more info at" +
		" https://github.com/nproc/run"
)

// NewApp returns a configured cli application.
func NewApp(loaderFunc valuesloader.ValueLoaderFunc) *cli.App {
	app := cli.NewApp()
	app.Version = build.Version
	app.Compiled = build.CompiledTime().UTC()
	app.Metadata = map[string]interface{}{"commit": build.Commit}
	app.Name = "run"
	app.Usage = "Docker container command runner"
	app.Description = description
	app.UsageText = app.Name + ` -i <file> -o <file> [command[ args]]`
	app.Authors = []cli.Author{
		{
			Name:  "Tarcisio Gruppi",
			Email: "txgruppi@gmail.com",
		},
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "input, i",
			Usage:  "The config template with the tokens to be replaced",
			EnvVar: "RUN_INPUT",
		},
		cli.StringFlag{
			Name:   "output, o",
			Usage:  "The output path for the compiled config file",
			EnvVar: "RUN_OUTPUT",
		},
	}
	app.Action = func(c *cli.Context) error {
		input := c.String("input")
		output := c.String("output")

		data, err := ioutil.ReadFile(input)
		if err != nil {
			return newExitError(err, 1)
		}

		vl, err := valuesloader.New(loaderFunc)
		if err != nil {
			return newExitError(err, 2)
		}

		tokens := text.Tokens(data)

		for _, token := range tokens {
			data = text.Replace(data, token, vl.Get(token))
		}

		err = ioutil.WriteFile(output, data, 0777)
		if err != nil {
			return newExitError(err, 3)
		}

		if len(c.Args()) == 0 {
			return nil
		}

		name := c.Args()[0]
		args := c.Args()[1:]

		cmd := exec.Command(name, args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = app.Writer
		cmd.Stderr = cli.ErrWriter

		return cmd.Run()
	}

	return app
}

func newExitError(err error, code int) error {
	return cli.NewExitError(err.Error(), code)
}
