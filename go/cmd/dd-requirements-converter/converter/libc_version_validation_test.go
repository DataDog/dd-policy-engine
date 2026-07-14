package converter

import (
	"encoding/json"
	"testing"
)

func TestInvalidLibcVersionsFailUnmarshal(t *testing.T) {
	testCases := []struct {
		name    string
		version string
	}{
		{name: "empty"},
		{name: "major only", version: "2"},
		{name: "major only bare prerelease", version: "2rc.1"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var requirement JSONlibc
			if err := json.Unmarshal(
				[]byte(`{"arch":"x64","supported":true,"min":"`+testCase.version+`"}`),
				&requirement,
			); err == nil {
				t.Fatalf("expected version %q to fail unmarshalling", testCase.version)
			}
		})
	}
}
