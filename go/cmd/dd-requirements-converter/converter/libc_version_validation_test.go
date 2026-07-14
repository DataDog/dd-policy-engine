package converter

import (
	"encoding/json"
	"testing"

	flatbuffers "github.com/google/flatbuffers/go"
)

func TestInvalidLibcVersionsReturnErrors(t *testing.T) {
	testCases := []struct {
		name    string
		version string
	}{
		{name: "empty"},
		{name: "major only", version: "2"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var requirement JSONlibc
			if err := json.Unmarshal(
				[]byte(`{"arch":"x64","supported":true,"min":"`+testCase.version+`"}`),
				&requirement,
			); err != nil {
				t.Fatal(err)
			}

			if _, err := requirement.ConvertToWLS(flatbuffers.NewBuilder(128), "glibc"); err == nil {
				t.Fatalf("expected version %q to be rejected", testCase.version)
			}
		})
	}
}
