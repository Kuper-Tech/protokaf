package tracing

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	jaegerConfig "github.com/uber/jaeger-client-go/config"
)

type Env struct {
	usage        string
	defaultValue string
}

//nolint:gofmt
var jaegerEnvs = map[string]*Env{
	"JAEGER_SERVICE_NAME":            {"The service name", "protokaf"},
	"JAEGER_TAGS":                    {"A comma separated list of name = value tracer level tags, which get added to all reported spans", ""}, //nolint:lll
	"JAEGER_SAMPLER_TYPE":            {"The sampler type", "const"},
	"JAEGER_SAMPLER_PARAM":           {"The sampler parameter (number)", "1.0"},
	"JAEGER_SAMPLING_ENDPOINT":       {"The url for the remote sampling conf when using sampler type remote", ""},
	"JAEGER_REPORTER_MAX_QUEUE_SIZE": {"The reporter's maximum queue size", ""},
	"JAEGER_REPORTER_FLUSH_INTERVAL": {"The reporter's flush interval (ms)", ""},
	"JAEGER_ENDPOINT":                {"Send spans to jaeger-collector at this URL", ""},
	"JAEGER_USER":                    {"User for basic http authentication when sending spans to jaeger-collector", ""},
	"JAEGER_PASSWORD":                {"Password for basic http authentication when sending spans to jaeger-collector", ""},
	"JAEGER_AGENT_HOST":              {"The hostname for communicating with agent via UDP", "0.0.0.0"},
	"JAEGER_AGENT_PORT":              {"The port for communicating with agent via UDP", "6831"},
}

func NewJaegerConfig() (*jaegerConfig.Configuration, error) {
	setupJaegerEnv()

	jaegerCfg, err := jaegerConfig.FromEnv()
	if err != nil {
		return nil, fmt.Errorf("could not parse Jaeger env vars: %w", err)
	}

	return jaegerCfg, nil
}

func SetJaegerFlags(flags *pflag.FlagSet) {
	for envName, env := range jaegerEnvs {
		var flag string
		name := flagName(envName)

		flags.StringVar(&flag, name, env.defaultValue, env.usage)
		_ = viper.BindPFlag(name, flags.Lookup(name))
	}
}

func setupJaegerEnv() {
	for envName := range jaegerEnvs {
		if os.Getenv(envName) == "" {
			name := flagName(envName)
			os.Setenv(envName, viper.GetString(name))
		}
	}
}

func flagName(envName string) string {
	return strings.ReplaceAll(strings.ToLower(envName), "_", "-")
}
