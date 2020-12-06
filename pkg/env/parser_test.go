package env

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name string
		args string
		want map[string]int
	}{
		{"labels", "app=app1=-1;app=app2=2;app=app3=0", map[string]int{"app=app1": -1, "app=app2": 2, "app=app3": 0}},
		{"names", "app1=-1;app2=2;app3=0", map[string]int{"app1": -1, "app2": 2, "app3": 0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseNames(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseNames() = %v, want %v", got, tt.want)
			}
		})
	}
}
