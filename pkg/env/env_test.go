package env

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"testing/quick"
)

func TestList(t *testing.T) {
	t.Parallel()

	r := rand.New(rand.NewSource(0))
	f := func() bool {
		var values []string
		for i := 0; i < r.Intn(20); i++ {
			values = append(values, fmt.Sprintf("value-%d", r.Int()))
		}

		joined := strings.Join(values, ";")
		parsed := List(joined)

		if len(values) != len(parsed) {
			t.Log("length must match")
			return false
		}

		for _, value := range values {
			if !parsed[value] {
				t.Log("parsed list must contain all values from input string")
				return false
			}
		}

		return true
	}

	if err := quick.Check(f, &quick.Config{Rand: r}); err != nil {
		t.Fatal(err)
	}
}
