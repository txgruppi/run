package valuesloader_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/txgruppi/run/valuesloader"
)

func TestValuesLoader(t *testing.T) {
	t.Run("nil loader", func(t *testing.T) {
		loader, err := valuesloader.New(nil)
		require.Nil(t, loader)
		require.EqualError(t, err, "nil loader at 0")
	})

	t.Run("EnvLoader", func(t *testing.T) {
		loader, err := valuesloader.EnvironmentLoader()
		require.Nil(t, err)
		require.NotNil(t, loader)

		t.Run("existing key", func(t *testing.T) {
			key := "testing_run_env_loader"
			value := "It works!"

			require.Nil(t, os.Setenv(key, value))
			loaded, ok := loader(key)
			require.True(t, ok)
			require.Equal(t, value, loaded)
			require.Nil(t, os.Unsetenv(key))
		})

		t.Run("missing key", func(t *testing.T) {
			loaded, ok := loader("some_key_that_will_not_be_set_in_the_enviroment")
			require.False(t, ok)
			require.Equal(t, "", loaded)
		})
	})

	t.Run("JSONLoader", func(t *testing.T) {
		data := []byte(`{"database":{"driver":"mysql","dsn":"user:password@tcp(host:port)/database"}}`)
		loader, err := valuesloader.JSONLoader(data)
		require.Nil(t, err)
		require.NotNil(t, loader)

		t.Run("existing props", func(t *testing.T) {
			pairs := map[string]string{
				"database.driver": "mysql",
				"database.dsn":    "user:password@tcp(host:port)/database",
			}

			for key, value := range pairs {
				t.Run(key, func(t *testing.T) {
					loaded, ok := loader(key)
					require.True(t, ok)
					require.Equal(t, value, loaded)
				})
			}
		})

		t.Run("missing or invalid props", func(t *testing.T) {
			pairs := map[string]string{
				"database":               "",
				"some_non_existing_prop": "",
			}

			for key, value := range pairs {
				t.Run(key, func(t *testing.T) {
					loaded, ok := loader(key)
					require.False(t, ok)
					require.Equal(t, value, loaded)
				})
			}
		})
	})

	t.Run("RemoteJSONLoader", func(t *testing.T) {
		data := []byte(`{"database":{"driver":"mysql","dsn":"user:password@tcp(host:port)/database"}}`)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write(data)
		}))
		defer server.Close()

		loader, err := valuesloader.RemoteJSONLoader(server.URL)
		require.Nil(t, err)
		require.NotNil(t, loader)

		t.Run("existing props", func(t *testing.T) {
			pairs := map[string]string{
				"database.driver": "mysql",
				"database.dsn":    "user:password@tcp(host:port)/database",
			}

			for key, value := range pairs {
				t.Run(key, func(t *testing.T) {
					loaded, ok := loader(key)
					require.True(t, ok)
					require.Equal(t, value, loaded)
				})
			}
		})

		t.Run("missing or invalid props", func(t *testing.T) {
			pairs := map[string]string{
				"database":               "",
				"some_non_existing_prop": "",
			}

			for key, value := range pairs {
				t.Run(key, func(t *testing.T) {
					loaded, ok := loader(key)
					require.False(t, ok)
					require.Equal(t, value, loaded)
				})
			}
		})
	})

	t.Run("JSONFileLoader", func(t *testing.T) {
		data := []byte(`{"database":{"driver":"mysql","dsn":"user:password@tcp(host:port)/database"}}`)

		file, err := ioutil.TempFile(os.TempDir(), "run-test")
		require.Nil(t, err)
		defer file.Close()

		_, err = file.Write(data)
		require.Nil(t, err)

		loader, err := valuesloader.JSONFileLoader(file.Name())
		require.Nil(t, err)
		require.NotNil(t, loader)

		t.Run("existing props", func(t *testing.T) {
			pairs := map[string]string{
				"database.driver": "mysql",
				"database.dsn":    "user:password@tcp(host:port)/database",
			}

			for key, value := range pairs {
				t.Run(key, func(t *testing.T) {
					loaded, ok := loader(key)
					require.True(t, ok)
					require.Equal(t, value, loaded)
				})
			}
		})

		t.Run("missing or invalid props", func(t *testing.T) {
			pairs := map[string]string{
				"database":               "",
				"some_non_existing_prop": "",
			}

			for key, value := range pairs {
				t.Run(key, func(t *testing.T) {
					loaded, ok := loader(key)
					require.False(t, ok)
					require.Equal(t, value, loaded)
				})
			}
		})
	})

	t.Run("multiple loaders", func(t *testing.T) {
		dataLocal := []byte(`{"database":{"driver":"mysql","dsn":"user:password@tcp(host:port)/database"}}`)
		dataRemote := []byte(`{"server":{"bind":"0.0.0.0","port":80,"just_some_float":1.234},"types":{"null":null,"true":true,"false":false}}`)

		pairs := map[string]string{
			"database.driver":        "mysql",
			"database.dsn":           "user:password@tcp(host:port)/database",
			"server.bind":            "0.0.0.0",
			"server.port":            "80",
			"PORT@env":               "3000",
			"server.just_some_float": "1.234",
			"types.null":             "",
			"types.true":             "true",
			"types.false":            "false",
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write(dataRemote)
		}))
		defer server.Close()

		envLoader, err := valuesloader.EnvironmentLoader()
		require.Nil(t, err)

		jsonLoader, err := valuesloader.JSONLoader(dataLocal)
		require.Nil(t, err)

		remoteJSONLoader, err := valuesloader.RemoteJSONLoader(server.URL)
		require.Nil(t, err)

		loader, err := valuesloader.New(
			envLoader,
			jsonLoader,
			remoteJSONLoader,
		)
		require.Nil(t, err)
		require.NotNil(t, loader)

		t.Run("existing key", func(t *testing.T) {
			for key, value := range pairs {
				t.Run(key, func(t *testing.T) {
					isEnv := false
					if strings.HasSuffix(key, "@env") {
						isEnv = true
						key = strings.TrimSuffix(key, "@env")
					}

					if isEnv {
						require.Nil(t, os.Setenv(key, value))
					}

					loadedFromLookup, ok := loader.Lookup(key)
					require.True(t, ok)
					require.Equal(t, value, loadedFromLookup)

					loadedFromGet := loader.Get(key)
					require.Equal(t, value, loadedFromGet)

					if isEnv {
						require.Nil(t, os.Unsetenv(key))
					}
				})
			}
		})

		t.Run("missing key", func(t *testing.T) {
			key := "some_non_existing_key"

			loadedFromLookup, ok := loader.Lookup(key)
			require.False(t, ok)
			require.Equal(t, "", loadedFromLookup)

			loadedFromGet := loader.Get(key)
			require.Equal(t, "", loadedFromGet)
		})
	})
}
