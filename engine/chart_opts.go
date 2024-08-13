package engine

import "helm.sh/helm/v3/pkg/action"

type ChartOptions struct {
	action.ChartPathOptions
	Name string

	ValuesCustomizers []ValuesCustomizer
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
