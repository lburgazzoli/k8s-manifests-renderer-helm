package resources

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/itchyny/gojq"
)

func JQ(expression string) func(unstructured.Unstructured) (unstructured.Unstructured, error) {
	return func(in unstructured.Unstructured) (unstructured.Unstructured, error) {
		query, err := gojq.Parse(expression)
		if err != nil {
			return in, fmt.Errorf("unable to parse expression %s: %w", expression, err)
		}

		it := query.Run(in.Object)

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
}
