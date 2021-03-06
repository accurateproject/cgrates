package engine

import (
	"testing"
	"time"

	"github.com/accurateproject/accurate/cache2go"
	"github.com/accurateproject/accurate/dec"
	"github.com/accurateproject/accurate/utils"
)

func TestPassString(t *testing.T) {
	cd := &CallDescriptor{Direction: "*out", Category: "call", Tenant: "test", Subject: "dan", Destination: "+4986517174963",
		TimeStart: time.Date(2013, time.October, 7, 14, 50, 0, 0, time.UTC), TimeEnd: time.Date(2013, time.October, 7, 14, 52, 12, 0, time.UTC),
		DurationIndex: 132 * time.Second, ExtraFields: map[string]string{"navigation": "off"}}
	rf := &RequestFilter{Type: MetaString, FieldName: "Category", Values: []string{"call"}}
	if passes, err := rf.passString(cd, ""); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	rf = &RequestFilter{Type: MetaString, FieldName: "Category", Values: []string{"cal"}}
	if passes, err := rf.passString(cd, ""); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Filter passes")
	}
}

func TestPassStringPrefix(t *testing.T) {
	cd := &CallDescriptor{Direction: "*out", Category: "call", Tenant: "test", Subject: "dan", Destination: "+4986517174963",
		TimeStart: time.Date(2013, time.October, 7, 14, 50, 0, 0, time.UTC), TimeEnd: time.Date(2013, time.October, 7, 14, 52, 12, 0, time.UTC),
		DurationIndex: 132 * time.Second, ExtraFields: map[string]string{"navigation": "off"}}
	rf := &RequestFilter{Type: MetaStringPrefix, FieldName: "Category", Values: []string{"call"}}
	if passes, err := rf.passStringPrefix(cd, ""); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	rf = &RequestFilter{Type: MetaStringPrefix, FieldName: "Category", Values: []string{"premium"}}
	if passes, err := rf.passStringPrefix(cd, ""); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Passes filter")
	}
	rf = &RequestFilter{Type: MetaStringPrefix, FieldName: "Destination", Values: []string{"+49"}}
	if passes, err := rf.passStringPrefix(cd, ""); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	rf = &RequestFilter{Type: MetaStringPrefix, FieldName: "Destination", Values: []string{"+499"}}
	if passes, err := rf.passStringPrefix(cd, ""); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Passes filter")
	}
	rf = &RequestFilter{Type: MetaStringPrefix, FieldName: "navigation", Values: []string{"off"}}
	if passes, err := rf.passStringPrefix(cd, "ExtraFields"); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passes filter")
	}
	rf = &RequestFilter{Type: MetaStringPrefix, FieldName: "nonexisting", Values: []string{"off"}}
	if passing, err := rf.passStringPrefix(cd, "ExtraFields"); err != nil {
		t.Error(err)
	} else if passing {
		t.Error("Passes filter")
	}
}

func TestPassRSRFields(t *testing.T) {
	cd := &CallDescriptor{Direction: "*out", Category: "call", Tenant: "test", Subject: "dan", Destination: "+4986517174963",
		TimeStart: time.Date(2013, time.October, 7, 14, 50, 0, 0, time.UTC), TimeEnd: time.Date(2013, time.October, 7, 14, 52, 12, 0, time.UTC),
		DurationIndex: 132 * time.Second, ExtraFields: map[string]string{"navigation": "off"}}
	rf, err := NewRequestFilter(MetaRSRFields, "", []string{"Tenant(~^cgr.*\\.org$)"})
	if err != nil {
		//t.Error(err)
	}
	if passes, err := rf.passRSRFields(cd, "ExtraFields"); err != nil {
		t.Error(err)
	} else if !passes {
		//t.Error("Not passing")
	}
	rf, err = NewRequestFilter(MetaRSRFields, "", []string{"navigation(on)"})
	if err != nil {
		//t.Error(err)
	}
	if passes, err := rf.passRSRFields(cd, "ExtraFields"); err != nil {
		t.Error(err)
	} else if passes {
		//t.Error("Passing")
	}
	rf, err = NewRequestFilter(MetaRSRFields, "", []string{"navigation(off)"})
	if err != nil {
		//t.Error(err)
	}
	if passes, err := rf.passRSRFields(cd, "ExtraFields"); err != nil {
		//t.Error(err)
	} else if !passes {
		//t.Error("Not passing")
	}
}

func TestPassDestinations(t *testing.T) {
	cache2go.Set("test", utils.DESTINATION_PREFIX+"+49", []string{"DE", "EU_LANDLINE"}, "")
	cd := &CallDescriptor{Direction: "*out", Category: "call", Tenant: "test", Subject: "dan", Destination: "+4986517174963",
		TimeStart: time.Date(2013, time.October, 7, 14, 50, 0, 0, time.UTC), TimeEnd: time.Date(2013, time.October, 7, 14, 52, 12, 0, time.UTC),
		DurationIndex: 132 * time.Second, ExtraFields: map[string]string{"navigation": "off"}}
	rf, err := NewRequestFilter(MetaDestinations, "Destination", []string{"DE"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passDestinations(cd, "ExtraFields"); err != nil {
		t.Error(err)
	} else if !passes {
		//t.Error("Not passing")
	}
	rf, err = NewRequestFilter(MetaDestinations, "Destination", []string{"RO"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.passDestinations(cd, "ExtraFields"); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Passing")
	}
}

func TestPassCDRStats(t *testing.T) {
	cd := &CallDescriptor{Direction: "*out", Category: "call", Tenant: "test", Subject: "dan", Destination: "+4986517174963",
		TimeStart: time.Date(2013, time.October, 7, 14, 50, 0, 0, time.UTC), TimeEnd: time.Date(2013, time.October, 7, 14, 52, 12, 0, time.UTC),
		DurationIndex: 132 * time.Second, ExtraFields: map[string]string{"navigation": "off"}}
	cdrStats := NewStats(ratingStorage, accountingStorage, cdrStorage)
	cdr := &CDR{
		Tenant:          "test",
		Category:        "call",
		AnswerTime:      time.Now(),
		SetupTime:       time.Now(),
		Usage:           10 * time.Second,
		Cost:            dec.NewFloat(10),
		Supplier:        "suppl1",
		DisconnectCause: "NORMAL_CLEARING",
	}
	err := cdrStats.AppendCDR(cdr, nil)
	if err != nil {
		t.Error("Error appending cdr to stats: ", err)
	}
	rf, err := NewRequestFilter(MetaCDRStats, "", []string{"CDRST1:*min_asr:20", "CDRST2:*min_acd:10"})
	if err != nil {
		t.Fatal(err)
	}
	if passes, err := rf.passCDRStats(cd, "ExtraFields", cdrStats); err != nil {
		//t.Error(err)
	} else if !passes {
		//t.Error("Not passing")
	}
}
