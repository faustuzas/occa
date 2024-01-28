package service

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

type SampleService struct {
	number int
}

type SampleServiceConfiguration struct {
	Num int `yaml:"num"`
}

func (s SampleServiceConfiguration) Build() (SampleService, error) {
	return SampleService{number: s.Num}, nil
}

func TestExternalServiceFromYAML(t *testing.T) {
	yamlCfg := `num: 12`

	var cfg ExternalService[SampleService, SampleServiceConfiguration]

	err := yaml.Unmarshal([]byte(yamlCfg), &cfg)
	require.NoError(t, err)

	service, err := cfg.GetService()
	require.NoError(t, err)

	require.Equal(t, 12, service.number)
}

func TestExternalServiceFromImpl(t *testing.T) {
	cfg := FromImpl[SampleService, SampleServiceConfiguration](SampleService{number: 19})

	service, err := cfg.GetService()
	require.NoError(t, err)

	require.Equal(t, 19, service.number)
}

func TestExternalServiceFromConfig(t *testing.T) {
	cfg := FromConfig[SampleService, SampleServiceConfiguration](SampleServiceConfiguration{
		Num: 19,
	})

	service, err := cfg.GetService()
	require.NoError(t, err)

	require.Equal(t, 19, service.number)
}
