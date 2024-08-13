package engine

import (
	"context"
	"fmt"
	"slices"
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
	ctx context.Context,
	name string,
	namespace string,
	revision int,
	values map[string]interface{},
) ([]unstructured.Unstructured, error) {
	rv, err := c.renderValues(ctx, name, namespace, revision, values)
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

	answer, err := c.customizeResources(ctx, result)
	if err != nil {
		return nil, err
	}

	return answer, nil
}

func (c *Chart) renderValues(
	ctx context.Context,
	name string,
	namespace string,
	revision int,
	values map[string]interface{},
) (chartutil.Values, error) {
	if values == nil {
		values = make(map[string]interface{})
	}

	overrides := c.options.overrides
	if overrides == nil {
		overrides = make(map[string]interface{})
	}

	for i := range c.options.valuesCustomizers {
		nv, err := c.options.valuesCustomizers[i](ctx, values)
		if err != nil {
			return chartutil.Values{}, fmt.Errorf("unable to cusomise values: %w", err)
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

func (c *Chart) customizeResources(ctx context.Context, resources []unstructured.Unstructured) ([]unstructured.Unstructured, error) {
	if len(resources) == 0 {
		return resources, nil
	}

	if len(c.options.resourcesCustomizers) == 0 {
		return resources, nil
	}

	res := slices.Clone(resources)

	for _, rc := range c.options.resourcesCustomizers {
		for i := range res {
			resource, err := rc.Apply(ctx, res[i])

			if err != nil {
				return nil, fmt.Errorf(
					"cannot customise resource %s:%s/%s: %w",
					res[i].GroupVersionKind().String(),
					res[i].GetNamespace(),
					res[i].GetName(),
					err)
			}

			res[i] = resource
		}
	}

	return res, nil
}
