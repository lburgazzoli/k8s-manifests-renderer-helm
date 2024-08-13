package engine

import (
	"maps"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"helm.sh/helm/v3/pkg/action"
)

type ValuesCustomizer func(map[string]interface{}) (map[string]interface{}, error)
type ResourcesCustomizer func(unstructured.Unstructured) (unstructured.Unstructured, error)

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

func WithValuesCustomizer(value ValuesCustomizer) ChartOption {
	return func(opts *ChartOptions) {
		opts.valuesCustomizers = append(opts.valuesCustomizers, value)
	}
}

func WithResourcesCustomizer(value ResourcesCustomizer) ChartOption {
	return func(opts *ChartOptions) {
		opts.resourcesCustomizers = append(opts.resourcesCustomizers, value)
	}
}

func WithOverrides(value map[string]interface{}) ChartOption {
	return func(opts *ChartOptions) {
		opts.overrides = maps.Clone(value)
	}
}
