package dockerversion

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/howtowhale/dvm/dvm-helper/internal/test"
)

func TestVersion_findLatestEdgeVersion(t *testing.T) {
	releaseListing := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintln(w, test.LoadTestData("edge_releases.html"))
	}))

	v, err := findLatestEdgeVersion(releaseListing.URL)
	if err != nil {
		t.Fatalf("%#v", err)
	}

	wantV := "17.06.0-ce"
	gotV := v.String()
	if wantV != gotV {
		t.Fatalf("Expected '%s', got '%s'", wantV, gotV)
	}
}
