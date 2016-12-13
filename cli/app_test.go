package cli_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"testing"

	rcli "github.com/nproc/run/cli"
	"github.com/nproc/run/valuesloader"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

const (
	template = `[database]
url = "{{MONGO_URL}}"

[jwt]
secret = "{{JWT_SECRET}}"

[server]
bind = "{{SERVER_BIND}}"
port = "{{SERVER_PORT}}"`
	fullReplace = `[database]
url = "mongo://user:password@my.server.com:56789/thedatabase"

[jwt]
secret = "#&EYR%Zdv%ta&3f&KHNW"

[server]
bind = "0.0.0.0"
port = "3456"`
	partialReplace = `[database]
url = "mongo://user:password@my.server.com:56789/thedatabase"

[jwt]
secret = ""

[server]
bind = "0.0.0.0"
port = ""`
)

var (
	fullEnv = map[string]string{
		"MONGO_URL":   "mongo://user:password@my.server.com:56789/thedatabase",
		"JWT_SECRET":  "#&EYR%Zdv%ta&3f&KHNW",
		"SERVER_BIND": "0.0.0.0",
		"SERVER_PORT": "3456",
	}
	partialEnv = map[string]string{
		"MONGO_URL":   "mongo://user:password@my.server.com:56789/thedatabase",
		"SERVER_BIND": "0.0.0.0",
	}
)

func TestApp(t *testing.T) {
	var lastExitCode int
	cli.OsExiter = func(code int) {
		lastExitCode = code
	}

	t.Run("invalid arguments", func(t *testing.T) {
		app := rcli.NewApp(envLoader(fullEnv))

		t.Run("input do not exist", func(t *testing.T) {
			assert := assert.New(t)
			lastExitCode = 0

			dir, err := makeTempDir()
			assert.Nil(err)

			input := "/some/fake/path/to/a/fake/file.toml"
			output := path.Join(dir, "new-file.toml")

			var stdout bytes.Buffer
			var stderr bytes.Buffer

			app.Writer = &stdout
			cli.ErrWriter = &stderr

			args := []string{"run", "-i", input, "-o", output}
			err = app.Run(args)
			assert.EqualError(err, "open /some/fake/path/to/a/fake/file.toml: no such file or directory")
			exitErr, ok := err.(cli.ExitCoder)
			assert.NotNil(exitErr)
			assert.True(ok)
			assert.Equal(1, exitErr.ExitCode())
			assert.Equal(1, lastExitCode)
			assert.Empty(stdout.String())
			assert.Equal("open /some/fake/path/to/a/fake/file.toml: no such file or directory\n", stderr.String())
		})

		t.Run("input is a directory", func(t *testing.T) {
			assert := assert.New(t)
			lastExitCode = 0

			dir, err := makeTempDir()
			assert.Nil(err)

			input := dir
			output := path.Join(dir, "new-file.toml")

			var stdout bytes.Buffer
			var stderr bytes.Buffer

			app.Writer = &stdout
			cli.ErrWriter = &stderr

			args := []string{"run", "-i", input, "-o", output}
			err = app.Run(args)
			assert.EqualError(err, "read "+input+": is a directory")
			exitErr, ok := err.(cli.ExitCoder)
			assert.NotNil(exitErr)
			assert.True(ok)
			assert.Equal(1, exitErr.ExitCode())
			assert.Equal(1, lastExitCode)
			assert.Empty(stdout.String())
			assert.Equal("read "+input+": is a directory\n", stderr.String())
		})

		t.Run("input is not readable", func(t *testing.T) {
			u, err := user.Current()
			if err != nil {
				t.Error(err)
				t.FailNow()
			}
			if u.Uid == "0" {
				t.Skipf("When running as root this test always fail")
			}

			assert := assert.New(t)
			lastExitCode = 0

			dir, err := makeTempDir()
			assert.Nil(err)

			file, err := makeTempFile("", 0333)
			assert.Nil(err)

			input := file
			output := path.Join(dir, "new-file.toml")

			var stdout bytes.Buffer
			var stderr bytes.Buffer

			app.Writer = &stdout
			cli.ErrWriter = &stderr

			args := []string{"run", "-i", input, "-o", output}
			err = app.Run(args)
			assert.EqualError(err, "open "+input+": permission denied")
			exitErr, ok := err.(cli.ExitCoder)
			assert.NotNil(exitErr)
			assert.True(ok)
			assert.Equal(1, exitErr.ExitCode())
			assert.Equal(1, lastExitCode)
			assert.Empty(stdout.String())
			assert.Equal("open "+input+": permission denied\n", stderr.String())
		})

		t.Run("output is a directory", func(t *testing.T) {
			assert := assert.New(t)
			lastExitCode = 0

			dir, err := makeTempDir()
			assert.Nil(err)

			file, err := makeTempFile(template, 0777)
			assert.Nil(err)

			input := file
			output := dir

			var stdout bytes.Buffer
			var stderr bytes.Buffer

			app.Writer = &stdout
			cli.ErrWriter = &stderr

			args := []string{"run", "-i", input, "-o", output}
			err = app.Run(args)
			assert.EqualError(err, "open "+output+": is a directory")
			exitErr, ok := err.(cli.ExitCoder)
			assert.NotNil(exitErr)
			assert.True(ok)
			assert.Equal(3, exitErr.ExitCode())
			assert.Equal(3, lastExitCode)
			assert.Empty(stdout.String())
			assert.Equal("open "+output+": is a directory\n", stderr.String())
		})

		t.Run("output is not writable", func(t *testing.T) {
			u, err := user.Current()
			if err != nil {
				t.Error(err)
				t.FailNow()
			}
			if u.Uid == "0" {
				t.Skipf("When running as root this test always fail")
			}

			assert := assert.New(t)
			lastExitCode = 0

			input, err := makeTempFile(template, 0777)
			assert.Nil(err)

			output, err := makeTempFile("", 0555)
			assert.Nil(err)

			var stdout bytes.Buffer
			var stderr bytes.Buffer

			app.Writer = &stdout
			cli.ErrWriter = &stderr

			args := []string{"run", "-i", input, "-o", output}
			err = app.Run(args)
			assert.EqualError(err, "open "+output+": permission denied")
			exitErr, ok := err.(cli.ExitCoder)
			assert.NotNil(exitErr)
			assert.True(ok)
			assert.Equal(3, exitErr.ExitCode())
			assert.Equal(3, lastExitCode)
			assert.Empty(stdout.String())
			assert.Equal("open "+output+": permission denied\n", stderr.String())
		})
	})

	t.Run("partial replace", func(t *testing.T) {
		assert := assert.New(t)
		lastExitCode = 0

		app := rcli.NewApp(envLoader(partialEnv))

		input, err := makeTempFile(template, 0777)
		assert.Nil(err)

		output, err := makeTempFile("", 0777)
		assert.Nil(err)

		var stdout bytes.Buffer
		var stderr bytes.Buffer

		app.Writer = &stdout
		cli.ErrWriter = &stderr

		args := []string{"run", "-i", input, "-o", output}
		err = app.Run(args)
		assert.Nil(err)
		assert.Equal(0, lastExitCode)

		contents, err := ioutil.ReadFile(output)
		assert.Nil(err)
		assert.Equal(partialReplace, string(contents))
	})

	t.Run("full replace", func(t *testing.T) {
		assert := assert.New(t)
		lastExitCode = 0

		app := rcli.NewApp(envLoader(fullEnv))

		input, err := makeTempFile(template, 0777)
		assert.Nil(err)

		output, err := makeTempFile("", 0777)
		assert.Nil(err)

		var stdout bytes.Buffer
		var stderr bytes.Buffer

		app.Writer = &stdout
		cli.ErrWriter = &stderr

		args := []string{"run", "-i", input, "-o", output}
		err = app.Run(args)
		assert.Nil(err)
		assert.Equal(0, lastExitCode)

		contents, err := ioutil.ReadFile(output)
		assert.Nil(err)
		assert.Equal(fullReplace, string(contents))
	})

	t.Run("command run successfully", func(t *testing.T) {
		assert := assert.New(t)
		lastExitCode = 0

		app := rcli.NewApp(envLoader(fullEnv))

		input, err := makeTempFile(template, 0777)
		assert.Nil(err)

		output, err := makeTempFile("", 0777)
		assert.Nil(err)

		var stdout bytes.Buffer
		var stderr bytes.Buffer

		app.Writer = &stdout
		cli.ErrWriter = &stderr

		args := []string{"run", "-i", input, "-o", output, "echo", "it", "works"}
		err = app.Run(args)
		assert.Nil(err)
		assert.Equal(0, lastExitCode)

		assert.Empty(stderr.String())
		assert.Equal("it works\n", stdout.String())
	})

	t.Run("nil loaderFunc", func(t *testing.T) {
		assert := assert.New(t)
		lastExitCode = 0

		app := rcli.NewApp(nil)

		input, err := makeTempFile(template, 0777)
		assert.Nil(err)

		output, err := makeTempFile("", 0777)
		assert.Nil(err)

		var stdout bytes.Buffer
		var stderr bytes.Buffer

		app.Writer = &stdout
		cli.ErrWriter = &stderr

		args := []string{"run", "-i", input, "-o", output, "echo", "it", "works"}
		err = app.Run(args)
		assert.EqualError(err, "loader is required")
		exitErr, ok := err.(cli.ExitCoder)
		assert.NotNil(exitErr)
		assert.True(ok)
		assert.Equal(2, exitErr.ExitCode())
		assert.Equal(2, lastExitCode)
		assert.Empty(stdout.String())
		assert.Equal("loader is required\n", stderr.String())
	})
}

func envLoader(env map[string]string) valuesloader.ValueLoaderFunc {
	return func(key string) string {
		return env[key]
	}
}

func makeTempDir() (string, error) {
	dir := path.Join(os.TempDir(), "github.com", "nproc", "run")
	if err := os.MkdirAll(dir, 0777); err != nil {
		return "", err
	}
	tmpDir, err := ioutil.TempDir(dir, "app_test_")
	if err != nil {
		return "", err
	}
	return tmpDir, nil
}

func makeTempFile(contents string, permissions os.FileMode) (string, error) {
	dir := path.Join(os.TempDir(), "github.com", "nproc", "run")
	if err := os.MkdirAll(dir, 0777); err != nil {
		return "", err
	}
	file, err := ioutil.TempFile(dir, "app_test_")
	if err != nil {
		return "", err
	}
	if _, err := file.WriteString(contents); err != nil {
		return "", nil
	}
	if err := file.Chmod(permissions); err != nil {
		return "", err
	}
	return file.Name(), nil
}
