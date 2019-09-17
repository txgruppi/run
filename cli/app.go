package cli

import (
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	"github.com/txgruppi/run/build"
	"github.com/txgruppi/run/text"
	"github.com/txgruppi/run/valuesloader"
	"github.com/urfave/cli"
)

const (
	description = "Compile config templates based on environment variables" +
		" and run a command after the template is successfully compiled." +
		"\n   The compiled file will have the permissions set to 0777." +
		"\n   Check the projects page for the documentation and more info at" +
		" https://github.com/txgruppi/run"
)

// NewApp returns a configured cli application.
func NewApp() *cli.App {
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
		cli.IntFlag{
			Name:   "delay, d",
			Usage:  "Number of seconds to wait before running the command",
			EnvVar: "RUN_DELAY",
		},
		cli.StringFlag{
			Name:   "json, j",
			Usage:  "JSON data to be used by JSONLoader",
			EnvVar: "RUN_JSON",
		},
		cli.StringFlag{
			Name:   "remote-json, r",
			Usage:  "URL to a JSON file to be used by RemoteJSONLoader",
			EnvVar: "RUN_REMOTE_JSON",
		},
		cli.StringFlag{
			Name:   "json-file, f",
			Usage:  "Path to a JSON file to be used by JSONFileLoader",
			EnvVar: "RUN_JSON_FILE",
		},
	}
	app.Action = func(c *cli.Context) error {
		input := c.String("input")
		output := c.String("output")
		delay := c.Int("delay")

		if delay > 0 {
			time.Sleep(time.Duration(delay) * time.Second)
		}

		if input != "" && output != "" {
			data, err := ioutil.ReadFile(input)
			if err != nil {
				return newExitError(err, 1)
			}

			envLoader, err := valuesloader.EnvironmentLoader()
			if err != nil {
				return newExitError(err, 4)
			}
			loaderFuncs := []valuesloader.ValueLoaderFunc{envLoader}

			if value := c.String("json"); value != "" {
				loader, err := valuesloader.JSONLoader([]byte(value))
				if err != nil {
					return newExitError(err, 5)
				}
				loaderFuncs = append(loaderFuncs, loader)
			}

			if value := c.String("remote-json"); value != "" {
				loader, err := valuesloader.RemoteJSONLoader(value)
				if err != nil {
					return newExitError(err, 6)
				}
				loaderFuncs = append(loaderFuncs, loader)
			}

			if value := c.String("json-file"); value != "" {
				loader, err := valuesloader.JSONFileLoader(value)
				if err != nil {
					return newExitError(err, 7)
				}
				loaderFuncs = append(loaderFuncs, loader)
			}

			vl, err := valuesloader.New(loaderFuncs...)
			if err != nil {
				return newExitError(err, 2)
			}

			tokens := text.Tokens(data)

		TokensLoop:
			for _, token := range tokens {
				for _, key := range token.Keys {
					if value, ok := vl.Lookup(key); ok {
						data = text.Replace(data, token.Raw, value)
						continue TokensLoop
					}
				}
				data = text.Replace(data, token.Raw, "")
			}

			err = ioutil.WriteFile(output, data, 0777)
			if err != nil {
				return newExitError(err, 3)
			}
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
