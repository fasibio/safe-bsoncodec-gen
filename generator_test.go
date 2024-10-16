package safebsoncodecgen_test

import (
	"os"
	"testing"

	safebsoncodecgen "github.com/fasibio/safe-bsoncodec-gen"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/require"
)

func TestGenerator(t *testing.T) {

	gen := safebsoncodecgen.NewBsonGenerator(&safebsoncodecgen.Config{
		PackageName: "test_model",
		PackagePath: "github.com/fasibio/safe-bsoncodec-gen/test_model",
		Direction:   "./test_model",
		FileName:    "test_gen.go",
	}, "github.com/fasibio/safe-bsoncodec-gen/test_model")
	require.NoError(t, gen.Run())
	f, err := os.ReadFile("./test_model/test_gen.go")
	require.NoError(t, err)
	snaps.MatchSnapshot(t, string(f))
	os.Remove("./test_model/test_gen.go")
}
