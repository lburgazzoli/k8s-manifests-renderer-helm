package engine

import (
	"fmt"
	"sort"
	"strings"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

type Chart struct {
	helmEngine engine.Engine
	helmChart  *chart.Chart
	options    ChartOptions
	decoder    runtime.Serializer
}

func (c *Chart) Render(
	name string,
	namespace string,
	revision int,
	values map[string]interface{},
	overrides map[string]interface{},
) ([]unstructured.Unstructured, error) {
	rv, err := c.renderValues(name, namespace, revision, values, overrides)
	if err != nil {
		return nil, fmt.Errorf("cannot render values: %w", err)
	}

	files, err := c.helmEngine.Render(c.helmChart, rv)
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

		ul, err := toUnstructured(c.decoder, []byte(v))
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

func (c *Chart) renderValues(
	name string,
	namespace string,
	revision int,
	values map[string]interface{},
	overrides map[string]interface{},
) (chartutil.Values, error) {
	if values == nil {
		values = make(map[string]interface{})
	}

	if overrides == nil {
		overrides = make(map[string]interface{})
	}

	for i := range c.options.ValuesCustomizers {
		nv, err := c.options.ValuesCustomizers[i](values)
		if err != nil {
			return chartutil.Values{}, fmt.Errorf("unable to cusomize values: %w", err)
		}

		values = nv
	}

	values = mergeMaps(values, overrides)

	err := chartutil.ProcessDependencies(c.helmChart, values)
	if err != nil {
		return chartutil.Values{}, fmt.Errorf("cannot process dependencies: %w", err)
	}

	rv, err := chartutil.ToRenderValues(
		c.helmChart,
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
