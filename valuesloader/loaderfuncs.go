package valuesloader

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/valyala/fastjson"
)

func EnvironmentLoader() (ValueLoaderFunc, error) {
	return func(key string) (string, bool) {
		return os.LookupEnv(key)
	}, nil
}

func JSONLoader(data []byte) (ValueLoaderFunc, error) {
	parsed, err := fastjson.ParseBytes(data)
	if err != nil {
		return nil, err
	}
	return func(key string) (string, bool) {
		keys := strings.Split(key, ".")
		if !parsed.Exists(keys...) {
			return "", false
		}

		value := parsed.Get(keys...)
		switch value.Type() {
		case fastjson.TypeNull:
			return "", true

		case fastjson.TypeTrue:
			return "true", true

		case fastjson.TypeFalse:
			return "false", true

		case fastjson.TypeString:
			str := value.String()
			return str[1 : len(str)-1], true

		case fastjson.TypeNumber:
			if n, err := value.Int64(); err == nil {
				return strconv.FormatInt(n, 10), true
			}
			if n, err := value.Float64(); err == nil {
				return strconv.FormatFloat(n, 'f', -1, 64), true
			}
			return "", false

		default:
			return "", false
		}
	}, nil
}

func RemoteJSONLoader(url string) (ValueLoaderFunc, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return JSONLoader(data)
}

func JSONFileLoader(filepath string) (ValueLoaderFunc, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return JSONLoader(data)
}

func AWSSecretsManagerLoader(secretArn string) (ValueLoaderFunc, error) {
	sess := session.New()
	sm := secretsmanager.New(sess)
	out, err := sm.GetSecretValue(&secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretArn),
	})
	if err != nil {
		return nil, err
	}
	if out == nil {
		return nil, fmt.Errorf("got unexpected nil value")
	}

	return JSONLoader([]byte(*out.SecretString))
}
