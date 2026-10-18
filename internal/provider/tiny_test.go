package provider

import (
	"strings"
	"testing"
)

func TestTinyRejectsIntentGroupWrites(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("MAGPIE_NO_ACCOUNTS", "1")
	if err := Save(Provider{ID: "fixture", Name: "Fixture", Chat: "http://127.0.0.1:1/v1", Key: "fixture", Models: []string{"model"}}); err != nil {
		t.Fatal(err)
	}
	for _, g := range []Group{
		{ID: "intent", Members: []string{"fixture/model"}, Rules: []Rule{{Use: "fixture/model", Intent: "code"}}},
		{ID: "classifier", Members: []string{"fixture/model"}, Classifier: "fixture/model"},
		{ID: "effort", Members: []string{"fixture/model"}, Effort: EffortAuto},
	} {
		if err := SaveGroup(g); err == nil || !strings.Contains(err.Error(), "intent routing is not available") {
			t.Fatalf("%s: %v", g.ID, err)
		}
	}
	if err := SaveGroup(Group{ID: "ordinary", Members: []string{"fixture/model"}}); err != nil {
		t.Fatal(err)
	}
}
