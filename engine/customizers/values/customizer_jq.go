package values

import (
	"context"
	"fmt"

	"github.com/itchyny/gojq"
)

func JQ(expression string) func(context.Context, map[string]interface{}) (map[string]interface{}, error) {
	return func(ctx context.Context, in map[string]interface{}) (map[string]interface{}, error) {
		query, err := gojq.Parse(expression)
		if err != nil {
			return nil, fmt.Errorf("unable to parse expression %s: %w", expression, err)
		}

		it := query.RunWithContext(ctx, in)

		v, ok := it.Next()
		if !ok {
			return in, nil
		}

		if err, ok := v.(error); ok {
			return in, err
		}

		if r, ok := v.(map[string]interface{}); ok {
			return r, nil
		}

		return in, nil
	}
}
