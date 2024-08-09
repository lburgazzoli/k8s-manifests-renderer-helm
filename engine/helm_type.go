package engine

type ChartSpec struct {
	// +kubebuilder:validation:Required
	Repo string `json:"repo,omitempty"`
	// +kubebuilder:validation:Required
	Name string `json:"name,omitempty"`
	// +kubebuilder:validation:Optional
	Version string `json:"version,omitempty"`
}

type ChartMeta struct {
	Repo    string `json:"repo,omitempty"`
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
}
