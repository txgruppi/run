package cli

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	"github.com/joho/godotenv"
	"github.com/txgruppi/run/build"
	"github.com/txgruppi/run/logger"
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
		cli.BoolFlag{
			Name:   "debug",
			Usage:  "Enable debug output",
			EnvVar: "RUN_DEBUG",
		},
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
		cli.StringFlag{
			Name:   "aws-secret",
			Usage:  "The ARN or name of a secret with a JSON encoded value",
			EnvVar: "RUN_AWS_SECRET_ARN",
		},
		cli.StringFlag{
			Name:   "env-file",
			Usage:  "A dotenv file template to be rendered and added to the environment",
			EnvVar: "RUN_ENV_FILE",
		},
		cli.StringFlag{
			Name:   "env-output-var",
			Usage:  "Create a environment variable with the contents of the output file",
			EnvVar: "RUN_ENV_OUTPUT_VAR",
		},
	}
	app.Action = func(c *cli.Context) (err error) {
		var envData []byte
		var envTokens []*text.Token
		var inputRender, envRender []byte
		var envSlice []string
		var vl *valuesloader.ValuesLoader

		logger.Debug = c.Bool("debug")
		input := c.String("input")
		output := c.String("output")
		delay := c.Int("delay")

		if delay > 0 {
			logger.Printf("Starting delay of %s", time.Duration(delay)*time.Second)
			time.Sleep(time.Duration(delay) * time.Second)
		}

		logger.Printf("Registering environment loader")
		envLoader, err := valuesloader.EnvironmentLoader()
		if err != nil {
			return newExitError(err, 4)
		}
		loaderFuncs := []valuesloader.ValueLoaderFunc{envLoader}

		if value := c.String("json"); value != "" {
			logger.Printf("Registering JSON loader with value %s", c.String("json"))
			loader, err := valuesloader.JSONLoader([]byte(value))
			if err != nil {
				return newExitError(err, 5)
			}
			loaderFuncs = append(loaderFuncs, loader)
		}

		if value := c.String("remote-json"); value != "" {
			logger.Printf("Registering remote JSON loader with URL %s", c.String("remote-json"))
			loader, err := valuesloader.RemoteJSONLoader(value)
			if err != nil {
				return newExitError(err, 6)
			}
			loaderFuncs = append(loaderFuncs, loader)
		}

		if value := c.String("json-file"); value != "" {
			logger.Printf("Registering JSON file loader with file %s", c.String("json-file"))
			loader, err := valuesloader.JSONFileLoader(value)
			if err != nil {
				return newExitError(err, 7)
			}
			loaderFuncs = append(loaderFuncs, loader)
		}

		if value := c.String("aws-secret"); value != "" {
			logger.Printf("Registering AWS SecretManager loader with SecretID %s", c.String("aws-secret"))
			loader, err := valuesloader.AWSSecretsManagerLoader(value)
			if err != nil {
				return newExitError(err, 8)
			}
			loaderFuncs = append(loaderFuncs, loader)
		}

		logger.Printf("Creating ValuesLoader")
		vl, err = valuesloader.New(loaderFuncs...)
		if err != nil {
			return newExitError(err, 2)
		}

		if input != "" {
			logger.Printf("Reading input file")
			inputData, err := ioutil.ReadFile(input)
			if err != nil {
				return newExitError(err, 1)
			}

			logger.Printf("Finding input tokens")
			inputTokens := text.Tokens(inputData)

			logger.Printf("Rendering input data")
			inputRender, err = render(inputData, inputTokens, vl)
			if err != nil {
				newExitError(err, 10)
			}

			if output != "" {
				logger.Printf("Writing output file")
				err = ioutil.WriteFile(output, inputRender, 0777)
				if err != nil {
					return newExitError(err, 3)
				}
			}
		}

		if c.String("env-file") != "" {
			logger.Printf("Reading env file %s", c.String("env-file"))
			envData, err = ioutil.ReadFile(c.String("env-file"))
			if err != nil {
				return newExitError(err, 9)
			}

			logger.Printf("Finding env file tokens")
			envTokens = text.Tokens(envData)

			logger.Printf("Rendering env file")
			envRender, err = render(envData, envTokens, vl)
			if err != nil {
				newExitError(err, 11)
			}

			logger.Printf("Getting complete environment values")
			envSlice, err = environ(envRender)
			if err != nil {
				return newExitError(err, 12)
			}
		}

		if c.String("env-output-var") != "" && inputRender != nil {
			logger.Printf("Creating output environment variable with value:")
			logger.Printf(string(inputRender))
			pair := c.String("env-output-var") + "=" + string(inputRender)
			if envSlice == nil {
				logger.Printf("Getting complete environment values")
				envSlice, err = environ([]byte(pair))
				if err != nil {
					return newExitError(err, 12)
				}
			} else {
				logger.Printf("Adding output environment variable")
				envSlice = append(envSlice, pair)
			}
		}

		if len(c.Args()) == 0 {
			logger.Printf("No command to run. Done")
			return nil
		}

		name := c.Args()[0]
		args := c.Args()[1:]

		logger.Printf("Preparing command %s with args %v", name, args)
		cmd := exec.Command(name, args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = app.Writer
		cmd.Stderr = cli.ErrWriter
		if envSlice != nil {
			logger.Printf("Adding enviroment variables to the command")
			cmd.Env = envSlice
		}

		logger.Printf("Running command")
		return cmd.Run()
	}

	return app
}

func render(in []byte, tks []*text.Token, vl *valuesloader.ValuesLoader) ([]byte, error) {
	out := make([]byte, len(in))
	copy(out, in)

TokensLoop:
	for _, token := range tks {
		for _, key := range token.Keys {
			if value, ok := vl.Lookup(key); ok {
				out = text.Replace(out, token.Raw, value)
				continue TokensLoop
			}
		}
		out = text.Replace(out, token.Raw, "")
	}
	return out, nil
}

func environ(envData []byte) ([]string, error) {
	r := bytes.NewReader(envData)
	em, err := godotenv.Parse(r)
	if err != nil {
		return nil, err
	}

	out := os.Environ()
	for k, v := range em {
		out = append(out, k+"="+v)
	}

	return out, nil
}

func newExitError(err error, code int) error {
	return cli.NewExitError(err.Error(), code)
}
