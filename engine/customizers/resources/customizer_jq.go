package resources

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/itchyny/gojq"
)

type JQCustomizer struct {
	expression string
	code       *gojq.Code
}

func (c *JQCustomizer) Configure(_ context.Context) error {
	query, err := gojq.Parse(c.expression)
	if err != nil {
		return fmt.Errorf("unable to parse expression %s: %w", c.expression, err)
	}

	code, err := gojq.Compile(
		query,
		gojq.WithVariables([]string{
			"$gvk", "$gv", "$group", "$version", "$kind", "$name", "$namespace",
		}),
	)

	if err != nil {
		return fmt.Errorf("unable to compile expression %s: %w", c.expression, err)
	}

	c.code = code

	return nil
}

func (c *JQCustomizer) Apply(ctx context.Context, in unstructured.Unstructured) (unstructured.Unstructured, error) {
	it := c.code.RunWithContext(ctx,
		in.Object,
		in.GetObjectKind().GroupVersionKind().GroupVersion().String()+":"+in.GetKind(),
		in.GetObjectKind().GroupVersionKind().GroupVersion().String(),
		in.GetObjectKind().GroupVersionKind().Group,
		in.GetObjectKind().GroupVersionKind().Version,
		in.GetObjectKind().GroupVersionKind().Kind,
		in.GetName(),
		in.GetNamespace(),
	)

	v, ok := it.Next()
	if !ok {
		return in, nil
	}

	if err, ok := v.(error); ok {
		return in, err
	}

	if r, ok := v.(map[string]interface{}); ok {
		in.Object = r
	}

	return in, nil
}

func JQ(expression string) *JQCustomizer {
	return &JQCustomizer{
		expression: expression,
	}
}
