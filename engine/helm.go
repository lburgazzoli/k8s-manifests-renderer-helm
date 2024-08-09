package engine

import (
	"fmt"
	"sort"
	"strings"

	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	k8syaml "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
)

type ValuesCustomizer func(map[string]any) (map[string]any, error)

func New() *Instance {
	return &Instance{
		e:                 engine.Engine{},
		env:               cli.New(),
		decoder:           k8syaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme),
		valuesCustomizers: make([]ValuesCustomizer, 0),
	}
}

type Instance struct {
	e                 engine.Engine
	env               *cli.EnvSettings
	decoder           runtime.Serializer
	valuesCustomizers []ValuesCustomizer
}

func (e *Instance) Customizer(customizer ValuesCustomizer, customizers ...ValuesCustomizer) {
	e.valuesCustomizers = append(e.valuesCustomizers, customizer)
	e.valuesCustomizers = append(e.valuesCustomizers, customizers...)
}

func (e *Instance) Load(cs ChartSpec, opts ...ChartOption) (*chart.Chart, error) {
	options := ChartOptions{}
	options.Name = cs.Name
	options.RepoURL = cs.Repo
	options.Version = cs.Version

	for i := range opts {
		opts[i](&options)
	}

	path, err := options.LocateChart(options.Name, e.env)
	if err != nil {
		return nil, fmt.Errorf("unable to load chart (repo: %s, name: %s, version: %s): %w", options.RepoURL, options.Name, options.Version, err)
	}

	c, err := loader.Load(path)
	if err != nil {
		return nil, fmt.Errorf("unable to load chart (repo: %s, name: %s, version: %s): %w", options.RepoURL, options.Name, options.Version, err)
	}

	return c, nil
}

func (e *Instance) Render(
	c *chart.Chart,
	name string,
	namespace string,
	revision int,
	values map[string]interface{},
	overrides map[string]interface{},
) ([]unstructured.Unstructured, error) {
	rv, err := e.renderValues(c, name, namespace, revision, values, overrides)
	if err != nil {
		return nil, fmt.Errorf("cannot render values: %w", err)
	}

	files, err := e.e.Render(c, rv)
	if err != nil {
		return nil, fmt.Errorf("cannot render a chart: %w", err)
	}

	keys := make([]string, 0, len(files))

	for k := range files {
		if !strings.HasSuffix(k, ".yaml") && !strings.HasSuffix(k, ".yml") {
			continue
		}

		keys = append(keys, k)
	}

	sort.Strings(keys)

	result := make([]unstructured.Unstructured, 0)

	for _, k := range keys {
		v := files[k]

		ul, err := toUnstructured(e.decoder, []byte(v))
		if err != nil {
			return nil, fmt.Errorf("cannot decode %s: %w", k, err)
		}

		if ul == nil {
			continue
		}

		result = append(result, ul...)
	}

	return result, nil
}

func (e *Instance) renderValues(
	c *chart.Chart,
	name string,
	namespace string,
	revision int,
	values map[string]interface{},
	overrides map[string]interface{},
) (chartutil.Values, error) {
	for i := range e.valuesCustomizers {
		nv, err := e.valuesCustomizers[i](values)
		if err != nil {
			return chartutil.Values{}, fmt.Errorf("unable to cusomize values: %w", err)
		}

		values = nv
	}

	values = mergeMaps(values, overrides)

	err := chartutil.ProcessDependencies(c, values)
	if err != nil {
		return chartutil.Values{}, fmt.Errorf("cannot process dependencies: %w", err)
	}

	rv, err := chartutil.ToRenderValues(
		c,
		values,
		chartutil.ReleaseOptions{
			Name:      name,
			Namespace: namespace,
			Revision:  revision,
			IsInstall: false,
			IsUpgrade: true,
		},
		nil,
	)
	if err != nil {
		return chartutil.Values{}, fmt.Errorf("cannot render values: %w", err)
	}

	return rv, nil
}
