package cli_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"os/user"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	rcli "github.com/txgruppi/run/cli"
	"github.com/urfave/cli"
)

const (
	template = `[database]
url = "{{RUN_TEST_ENV_MONGO_URL}}"

[jwt]
secret = "{{RUN_TEST_ENV_JWT_SECRET}}"

[server]
bind = "{{RUN_TEST_ENV_SERVER_BIND}}"
port = "{{RUN_TEST_ENV_SERVER_PORT}}"`

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

	allLoadersTemplate = `environment = "{{RUN_TEST_ENV_ENVIRONMENT}}"

[database]
driver = "{{database.driver}}"
dsn = "{{database.dsn}}"

[server]
bind = "{{RUN_TEST_ENV_SERVER_BIND || server.bind}}"
port = {{RUN_TEST_ENV_SERVER_PORT || server.port}}`

	allLoadersReplace = `environment = "development"

[database]
driver = "mysql"
dsn = "user:password@tcp(host:port)/database"

[server]
bind = "0.0.0.0"
port = 80`
)

var (
	fullEnv = map[string]string{
		"RUN_TEST_ENV_MONGO_URL":   "mongo://user:password@my.server.com:56789/thedatabase",
		"RUN_TEST_ENV_JWT_SECRET":  "#&EYR%Zdv%ta&3f&KHNW",
		"RUN_TEST_ENV_SERVER_BIND": "0.0.0.0",
		"RUN_TEST_ENV_SERVER_PORT": "3456",
	}
	partialEnv = map[string]string{
		"RUN_TEST_ENV_ENVIRONMENT": "development",
		"RUN_TEST_ENV_MONGO_URL":   "mongo://user:password@my.server.com:56789/thedatabase",
		"RUN_TEST_ENV_SERVER_BIND": "0.0.0.0",
	}
)

func TestApp(t *testing.T) {
	var lastExitCode int
	cli.OsExiter = func(code int) {
		lastExitCode = code
	}

	t.Run("invalid arguments", func(t *testing.T) {
		setEnv(fullEnv)
		app := rcli.NewApp()

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

		clearEnv(fullEnv)
	})

	t.Run("partial replace", func(t *testing.T) {
		setEnv(partialEnv)

		assert := assert.New(t)
		lastExitCode = 0

		app := rcli.NewApp()

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

		clearEnv(partialEnv)
	})

	t.Run("full replace", func(t *testing.T) {
		setEnv(fullEnv)

		assert := assert.New(t)
		lastExitCode = 0

		app := rcli.NewApp()

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

		clearEnv(fullEnv)
	})

	t.Run("command run successfully", func(t *testing.T) {
		setEnv(fullEnv)

		assert := assert.New(t)
		lastExitCode = 0

		app := rcli.NewApp()

		input, err := makeTempFile(template, 0777)
		assert.Nil(err)

		output, err := makeTempFile("", 0777)
		assert.Nil(err)

		var stdout bytes.Buffer
		var stderr bytes.Buffer

		app.Writer = &stdout
		cli.ErrWriter = &stderr

		args := []string{"run", "-i", input, "-o", output, "echo", "it", "works"}
		before := time.Now()
		err = app.Run(args)
		after := time.Now()
		assert.Nil(err)
		assert.Equal(0, lastExitCode)
		assert.True(after.Sub(before) < 1*time.Second)

		assert.Empty(stderr.String())
		assert.Equal("it works\n", stdout.String())

		clearEnv(fullEnv)
	})

	t.Run("run with delay", func(t *testing.T) {
		setEnv(fullEnv)

		assert := assert.New(t)
		lastExitCode = 0

		app := rcli.NewApp()

		input, err := makeTempFile(template, 0777)
		assert.Nil(err)

		output, err := makeTempFile("", 0777)
		assert.Nil(err)

		var stdout bytes.Buffer
		var stderr bytes.Buffer

		app.Writer = &stdout
		cli.ErrWriter = &stderr

		args := []string{"run", "-d", "1", "-i", input, "-o", output, "echo", "it", "works"}
		before := time.Now()
		err = app.Run(args)
		after := time.Now()
		assert.Nil(err)
		assert.Equal(0, lastExitCode)
		assert.True(after.Sub(before) > 1*time.Second)

		assert.Empty(stderr.String())
		assert.Equal("it works\n", stdout.String())

		clearEnv(fullEnv)
	})

	t.Run("all loaders", func(t *testing.T) {
		setEnv(partialEnv)

		assert := assert.New(t)
		lastExitCode = 0

		app := rcli.NewApp()

		input, err := makeTempFile(allLoadersTemplate, 0777)
		assert.Nil(err)

		output, err := makeTempFile("", 0777)
		assert.Nil(err)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"server":{"port":80}}`))
		}))
		defer server.Close()

		var stdout bytes.Buffer
		var stderr bytes.Buffer

		app.Writer = &stdout
		cli.ErrWriter = &stderr

		args := []string{"run", "-d", "1", "-j", `{"database":{"driver":"mysql","dsn":"user:password@tcp(host:port)/database"}}`, "-r", server.URL, "-i", input, "-o", output, "echo", "it", "works"}
		err = app.Run(args)
		assert.Nil(err)
		assert.Equal(0, lastExitCode)

		contents, err := ioutil.ReadFile(output)
		assert.Nil(err)
		assert.Equal(allLoadersReplace, string(contents))
		assert.Empty(stderr.String())

		clearEnv(partialEnv)
	})
}

func setEnv(m map[string]string) {
	for key, value := range m {
		os.Setenv(key, value)
	}
}

func clearEnv(m map[string]string) {
	for key, _ := range m {
		os.Unsetenv(key)
	}
}

func makeTempDir() (string, error) {
	dir := path.Join(os.TempDir(), "github.com", "txgruppi", "run")
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
	dir := path.Join(os.TempDir(), "github.com", "txgruppi", "run")
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
