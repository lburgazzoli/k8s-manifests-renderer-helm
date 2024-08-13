package engine_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/lburgazzoli/gomega-matchers/pkg/matchers/jq"
	"github.com/lburgazzoli/helm-libs/engine"
	"github.com/rs/xid"
)

func TestEngine(t *testing.T) {
	t.Parallel()

	g := NewWithT(t)

	e := engine.New()
	g.Expect(e).ShouldNot(BeNil())

	c, err := e.Load(engine.ChartSpec{
		Repo:    "https://dapr.github.io/helm-charts/",
		Name:    "dapr",
		Version: "1.13.5",
	})

	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(c).ShouldNot(BeNil())

	r, err := c.Render(
		t.Name(),
		xid.New().String(),
		0,
		map[string]interface{}{
			"dapr_operator": map[string]interface{}{
				"replicaCount": 5,
			},
		},
		nil)

	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(r).ShouldNot(BeEmpty())

	g.Expect(r).To(
		ContainElement(
			jq.Match(`.metadata.name == "dapr-operator" and .spec.replicas == 5`)))
}
