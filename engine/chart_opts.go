package engine

import (
	"maps"

	"helm.sh/helm/v3/pkg/action"
)

type ChartOptions struct {
	action.ChartPathOptions
	Name string

	ValuesCustomizers []ValuesCustomizer
	Overrides         map[string]interface{}
}

type ChartOption func(*ChartOptions)

func WithUsername(value string) ChartOption {
	return func(opts *ChartOptions) {
		opts.Username = value
	}
}

func WithPassword(value string) ChartOption {
	return func(opts *ChartOptions) {
		opts.Password = value
	}
}

func WithCustomizer(value ValuesCustomizer) ChartOption {
	return func(opts *ChartOptions) {
		opts.ValuesCustomizers = append(opts.ValuesCustomizers, value)
	}
}

func WithOverrides(value map[string]interface{}) ChartOption {
	return func(opts *ChartOptions) {
		opts.Overrides = maps.Clone(value)
	}
}
