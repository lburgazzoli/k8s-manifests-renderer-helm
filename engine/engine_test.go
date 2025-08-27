package engine_test

import (
	"testing"

	"github.com/lburgazzoli/k8s-manifests-renderer-helm/engine/customizers/resources"
	"github.com/lburgazzoli/k8s-manifests-renderer-helm/engine/customizers/values"

	. "github.com/onsi/gomega"

	"github.com/lburgazzoli/gomega-matchers/pkg/matchers/jq"
	"github.com/lburgazzoli/k8s-manifests-renderer-helm/engine"
	"github.com/rs/xid"
)

//nolint:gochecknoglobals
var cs = engine.ChartSpec{
	Repo:    "https://dapr.github.io/helm-charts/",
	Name:    "dapr",
	Version: "1.13.5",
}

//nolint:lll
func TestEngine(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	id := xid.New().String()

	g := NewWithT(t)

	e := engine.New()
	g.Expect(e).ShouldNot(BeNil())

	c, err := e.Load(ctx, cs)
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(c).ShouldNot(BeNil())
	g.Expect(c.Name()).Should(Equal(cs.Name))
	g.Expect(c.Version()).Should(Equal(cs.Version))
	g.Expect(c.Repo()).Should(Equal(cs.Repo))
	g.Expect(c.Spec()).Should(Equal(cs))

	crds, err := c.CRDObjects()
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(crds).ShouldNot(BeEmpty())

	r, err := c.Render(
		ctx,
		t.Name(),
		xid.New().String(),
		0,
		map[string]interface{}{
			"dapr_operator": map[string]interface{}{
				"replicaCount": 5,
			},
			"dapr_sidecar_injector": map[string]any{
				"image": map[string]any{
					"name": "docker.io/daprio/daprd:" + id,
				},
			},
		})

	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(r).ShouldNot(BeEmpty())

	g.Expect(r).To(
		ContainElement(
			jq.Match(`.metadata.name == "dapr-operator" and .spec.replicas == 5`)))
	g.Expect(r).To(
		ContainElement(
			jq.Match(`.metadata.name == "dapr-sidecar-injector" and (.spec.template.spec.containers[0].env[] | select(.name == "SIDECAR_IMAGE") | .value == "docker.io/daprio/daprd:%s")`, id)))
}

func TestEngineWithValuesCustomizers(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	g := NewWithT(t)

	e := engine.New()
	g.Expect(e).ShouldNot(BeNil())

	c, err := e.Load(
		ctx,
		cs,
		engine.WithValuesCustomizers(values.JQ(`.dapr_operator.replicaCount = 6`)),
	)

	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(c).ShouldNot(BeNil())

	r, err := c.Render(
		ctx,
		t.Name(),
		xid.New().String(),
		0,
		map[string]interface{}{
			"dapr_operator": map[string]interface{}{
				"replicaCount": 5,
			},
		})

	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(r).ShouldNot(BeEmpty())

	g.Expect(r).To(
		ContainElement(
			jq.Match(`.metadata.name == "dapr-operator" and .spec.replicas == 6`)))
}

const customiseDaprOperatorReplicas1 = `
if ( $gvk == "apps/v1:Deployment" and $name == "dapr-operator" ) 
then 
  .spec.replicas = 4
end
`

const customiseDaprOperatorReplicas2 = `
if ( $gv == "apps/v1" and $kind == "Deployment" and $name == "dapr-operator" ) 
then 
  .spec.replicas = 4
end
`

const customiseDaprOperatorReplicas3 = `
if ( $group == "apps" and $version == "v1" and $kind == "Deployment" and $name == "dapr-operator" ) 
then 
  .spec.replicas = 4
end
`

func TestEngineWithResourcesCustomizers(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	g := NewWithT(t)

	e := engine.New()
	g.Expect(e).ShouldNot(BeNil())

	var flagtests = []struct {
		name       string
		expression string
	}{
		{"gvk", customiseDaprOperatorReplicas1},
		{"gv_k", customiseDaprOperatorReplicas2},
		{"g_v_k", customiseDaprOperatorReplicas3},
	}

	for _, tt := range flagtests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c, err := e.Load(
				ctx,
				cs,
				engine.WithResourcesCustomizers(resources.JQ(tt.expression)),
			)

			g.Expect(err).ShouldNot(HaveOccurred())
			g.Expect(c).ShouldNot(BeNil())

			r, err := c.Render(
				ctx,
				t.Name(),
				xid.New().String(),
				0,
				nil,
			)

			g.Expect(err).ShouldNot(HaveOccurred())
			g.Expect(r).ShouldNot(BeEmpty())

			g.Expect(r).To(
				ContainElement(
					jq.Match(`.metadata.name == "dapr-operator" and .spec.replicas == 4`)))
		})
	}
}
