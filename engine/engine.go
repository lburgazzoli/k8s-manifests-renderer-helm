package engine

import (
	"context"
	"fmt"

	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	k8syaml "k8s.io/apimachinery/pkg/runtime/serializer/yaml"

	"helm.sh/helm/v3/pkg/engine"
)

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

func (in *Instance) Load(ctx context.Context, cs ChartSpec, opts ...ChartOption) (*Chart, error) {
	options := ChartOptions{}
	options.name = cs.Name
	options.pathOptions.RepoURL = cs.Repo
	options.pathOptions.Version = cs.Version

	for i := range opts {
		opts[i](&options)
	}

	path, err := options.pathOptions.LocateChart(options.name, in.env)
	if err != nil {
		return nil, fmt.Errorf(
			"unable to load chart (repo: %s, name: %s, version: %s): %w",
			options.pathOptions.RepoURL,
			options.name,
			options.pathOptions.Version, err)
	}

	c, err := loader.Load(path)
	if err != nil {
		return nil, fmt.Errorf(
			"unable to load chart (repo: %s, name: %s, version: %s): %w",
			options.pathOptions.RepoURL,
			options.name,
			options.pathOptions.Version, err)
	}

	for i := range options.resourcesCustomizers {
		if err := options.resourcesCustomizers[i].Configure(ctx); err != nil {
			return nil, fmt.Errorf("unable to initialize customizer: %w", err)
		}
	}

	rv := Chart{
		helmEngine: in.e,
		helmChart:  c,
		options:    options,
		decoder:    k8syaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme),
	}

	return &rv, nil
}
