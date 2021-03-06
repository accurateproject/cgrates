//+build integration

package engine

import (
	"encoding/json"
	"flag"
	"testing"
	"time"

	"github.com/accurateproject/accurate/dec"
	"github.com/accurateproject/accurate/utils"
)

// Arguments received via test command
var testLocal = flag.Bool("local", false, "Perform the tests only on local test environment, not by default.") // This flag will be passed here via "go test -local" args
var dataDir = flag.String("data_dir", "/usr/share/cgrates", "CGR data dir path here")

// Sample HttpJsonPost, more for usage purposes
func TestHttpJsonPost(t *testing.T) {
	if !*testLocal {
		return
	}
	cdrOut := &ExternalCDR{UniqueID: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String()), OrderID: 123, ToR: utils.VOICE, OriginID: "dsafdsaf",
		OriginHost: "192.168.1.1",
		Source:     utils.UNIT_TEST, RequestType: utils.META_RATED, Direction: "*out", Tenant: "cgrates.org",
		Category: "call", Account: "account1", Subject: "tgooiscs0014", Destination: "1002",
		SetupTime: time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String(), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String(),
		RunID: utils.DEFAULT_RUNID,
		Usage: "0.00000001", ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: dec.NewFloat(1.01),
	}
	jsn, _ := json.Marshal(cdrOut)
	if _, err := utils.HttpJsonPost("http://localhost:8000", false, jsn); err == nil {
		t.Error(err)
	}
}
