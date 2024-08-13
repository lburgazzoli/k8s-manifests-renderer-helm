package engine

import (
	"context"
	"maps"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"helm.sh/helm/v3/pkg/action"
)

type ValuesCustomizer func(context.Context, map[string]interface{}) (map[string]interface{}, error)

type ChartOptions struct {
	pathOptions action.ChartPathOptions
	name        string

	valuesCustomizers    []ValuesCustomizer
	resourcesCustomizers []ResourcesCustomizer
	overrides            map[string]interface{}
}

type ChartOption func(*ChartOptions)

func WithUsername(value string) ChartOption {
	return func(opts *ChartOptions) {
		opts.pathOptions.Username = value
	}
}

func WithPassword(value string) ChartOption {
	return func(opts *ChartOptions) {
		opts.pathOptions.Password = value
	}
}

func WithValuesCustomizers(values ...ValuesCustomizer) ChartOption {
	return func(opts *ChartOptions) {
		opts.valuesCustomizers = append(opts.valuesCustomizers, values...)
	}
}

func WithResourcesCustomizers(values ...ResourcesCustomizer) ChartOption {
	return func(opts *ChartOptions) {
		opts.resourcesCustomizers = append(opts.resourcesCustomizers, values...)
	}
}

func WithOverrides(value map[string]interface{}) ChartOption {
	return func(opts *ChartOptions) {
		opts.overrides = maps.Clone(value)
	}
}

type ResourcesCustomizer interface {
	Configure(ctx context.Context) error
	Apply(ctx context.Context, in unstructured.Unstructured) (unstructured.Unstructured, error)
}
