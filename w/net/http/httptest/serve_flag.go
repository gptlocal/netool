package httptest

import (
	"flag"
	"os"
	"strings"
)

var serveFlag string

func init() {
	if strSliceContainsPrefix(os.Args, "-httptest.serve=") || strSliceContainsPrefix(os.Args, "--httptest.serve=") {
		flag.StringVar(&serveFlag, "httptest.serve", "", "if non-empty, httptest.NewServer serves on this address and blocks.")
	}
}

func strSliceContainsPrefix(v []string, pre string) bool {
	for _, s := range v {
		if strings.HasPrefix(s, pre) {
			return true
		}
	}
	return false
}
