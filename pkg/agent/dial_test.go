package agent

import (
	"fmt"
	"github.com/mutagen-io/mutagen/pkg/mutagen"
	"testing"
)

func TestAgentInvocationPath(t *testing.T) {
	path := agentInvocationPath(false)
	expected := fmt.Sprintf("~/.mutagen/agents/%s/mutagen-agent", mutagen.Version)
	if path != expected {
		t.Errorf("Invocation path cmdExe=false, expected %s, got %s", expected, path)
	}

	path = agentInvocationPath(true)
	expected = fmt.Sprintf(".mutagen\\agents\\%s\\mutagen-agent", mutagen.Version)
	if path != expected {
		t.Errorf("Invocation path cmdExe=true, expected %s, got %s", expected, path)
	}
}
