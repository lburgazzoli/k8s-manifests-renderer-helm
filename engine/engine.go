package engine

import (
	"fmt"

	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	k8syaml "k8s.io/apimachinery/pkg/runtime/serializer/yaml"

	"helm.sh/helm/v3/pkg/engine"
)

type ValuesCustomizer func(map[string]interface{}) (map[string]interface{}, error)

func New() *Instance {
	return &Instance{
		e:   engine.Engine{},
		env: cli.New(),
	}
}

type Instance struct {
	e   engine.Engine
	env *cli.EnvSettings
}

func (in *Instance) Load(cs ChartSpec, opts ...ChartOption) (*Chart, error) {
	options := ChartOptions{}
	options.Name = cs.Name
	options.RepoURL = cs.Repo
	options.Version = cs.Version

	for i := range opts {
		opts[i](&options)
	}

	path, err := options.LocateChart(options.Name, in.env)
	if err != nil {
		return nil, fmt.Errorf("unable to load chart (repo: %s, name: %s, version: %s): %w", options.RepoURL, options.Name, options.Version, err)
	}

	c, err := loader.Load(path)
	if err != nil {
		return nil, fmt.Errorf("unable to load chart (repo: %s, name: %s, version: %s): %w", options.RepoURL, options.Name, options.Version, err)
	}

	rv := Chart{
		helmEngine: in.e,
		helmChart:  c,
		options:    options,
		decoder:    k8syaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme),
	}

	return &rv, nil
}
