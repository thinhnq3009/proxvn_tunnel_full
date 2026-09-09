package tunnel

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestMessageForceJSON(t *testing.T) {
	withoutForce, err := json.Marshal(Message{Type: "register", Subdomain: "factorio"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(withoutForce), `"force"`) {
		t.Fatalf("force=false must be omitted for backward compatibility: %s", withoutForce)
	}

	withForce, err := json.Marshal(Message{Type: "register", Subdomain: "factorio", Force: true})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(withForce), `"force":true`) {
		t.Fatalf("force=true missing from registration payload: %s", withForce)
	}
}
