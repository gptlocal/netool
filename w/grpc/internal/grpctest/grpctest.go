package grpctest

import (
	"reflect"
	"strings"
	"testing"
)

// RunSubTests runs all "Test___" functions that are methods of x as subtests
// of the current test.  If x contains methods "Setup(*testing.T)" or
// "Teardown(*testing.T)", those are run before or after each of the test
// functions, respectively.
//
// For example usage, see example_test.go.  Run it using:
//
//	$ go test -v -run TestExample .
//
// To run a specific test/subtest:
//
//	$ go test -v -run 'TestExample/^Something$' .
func RunSubTests(t *testing.T, x interface{}) {
	xt := reflect.TypeOf(x)
	xv := reflect.ValueOf(x)

	setup := getTestFunc(t, xv, "Setup")
	teardown := getTestFunc(t, xv, "Teardown")

	for i := 0; i < xt.NumMethod(); i++ {
		methodName := xt.Method(i).Name
		if !strings.HasPrefix(methodName, "Test") {
			continue
		}
		tfunc := getTestFunc(t, xv, methodName)
		t.Run(strings.TrimPrefix(methodName, "Test"), func(t *testing.T) {
			// Run leakcheck in t.Cleanup() to guarantee it is run even if tfunc
			// or setup uses t.Fatal().
			//
			// Note that a defer would run before t.Cleanup, so if a goroutine
			// is closed by a test's t.Cleanup, a deferred leakcheck would fail.
			t.Cleanup(func() { teardown(t) })
			setup(t)
			tfunc(t)
		})
	}
}

func getTestFunc(t *testing.T, xv reflect.Value, name string) func(*testing.T) {
	if m := xv.MethodByName(name); m.IsValid() {
		if f, ok := m.Interface().(func(*testing.T)); ok {
			return f
		}
		// Method exists but has the wrong type signature.
		t.Fatalf("grpctest: function %v has unexpected signature (%T)", name, m.Interface())
	}
	return func(*testing.T) {}
}
