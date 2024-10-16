package safebsoncodecgen_test

import (
	"reflect"
	"testing"

	"github.com/fasibio/safe"
	safebsoncodecgen "github.com/fasibio/safe-bsoncodec-gen"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

type BsonTestStruct struct {
	A int                              `bson:"a" json:"a,omitempty"`
	B safe.Option[int]                 `bson:"b" json:"b,omitempty"`
	C string                           `bson:"c" json:"c,omitempty"`
	F safe.Option[BsonSubTestStruct]   `bson:"f" json:"f,omitempty"`
	G safe.Option[[]BsonSubTestStruct] `bson:"g"`
}

type BsonSubTestStruct struct {
	D safe.Option[string] `bson:"d" json:"d,omitempty"`
}

func SomePtr[T any](v T) safe.Option[T] {
	return safe.Some(safe.Ptr(v))
}

func TestBsonMarshal(t *testing.T) {

	tests := []struct {
		Name string
		Data BsonTestStruct
	}{
		{
			Name: "happy path",
			Data: BsonTestStruct{
				A: 1,
				B: SomePtr(10),
				C: "Fooo",
				F: SomePtr(BsonSubTestStruct{
					D: SomePtr("baar"),
				}),
				G: SomePtr([]BsonSubTestStruct{
					{D: SomePtr("1")},
					{D: SomePtr("2")},
				}),
			},
		},
		{
			Name: "nil values",
			Data: BsonTestStruct{
				A: 1,
				B: safe.None[int](),
				C: "Fooo",
				F: SomePtr(BsonSubTestStruct{
					D: safe.None[string](),
				}),
				G: SomePtr([]BsonSubTestStruct{
					{D: safe.None[string]()},
					{D: SomePtr("2")},
				}),
			},
		},
	}
	bson.DefaultRegistry.RegisterTypeDecoder(reflect.TypeOf(safe.Option[string]{}), safebsoncodecgen.OptionCodec[string]{})
	bson.DefaultRegistry.RegisterTypeEncoder(reflect.TypeOf(safe.Option[string]{}), safebsoncodecgen.OptionCodec[string]{})

	bson.DefaultRegistry.RegisterTypeDecoder(reflect.TypeOf(safe.Option[int]{}), safebsoncodecgen.OptionCodec[int]{})
	bson.DefaultRegistry.RegisterTypeEncoder(reflect.TypeOf(safe.Option[int]{}), safebsoncodecgen.OptionCodec[int]{})

	// bson.DefaultRegistry.RegisterTypeDecoder(reflect.TypeOf(safe.Option[any]{}), safe.GenericOptionCodec{})
	// bson.DefaultRegistry.RegisterTypeEncoder(reflect.TypeOf(safe.Option[any]{}), safe.GenericOptionCodec{})

	bson.DefaultRegistry.RegisterTypeDecoder(reflect.TypeOf(safe.Option[BsonSubTestStruct]{}), safebsoncodecgen.OptionCodec[BsonSubTestStruct]{})
	bson.DefaultRegistry.RegisterTypeEncoder(reflect.TypeOf(safe.Option[BsonSubTestStruct]{}), safebsoncodecgen.OptionCodec[BsonSubTestStruct]{})
	bson.DefaultRegistry.RegisterTypeDecoder(reflect.TypeOf(safe.Option[[]BsonSubTestStruct]{}), safebsoncodecgen.OptionCodec[[]BsonSubTestStruct]{})
	bson.DefaultRegistry.RegisterTypeEncoder(reflect.TypeOf(safe.Option[[]BsonSubTestStruct]{}), safebsoncodecgen.OptionCodec[[]BsonSubTestStruct]{})
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			res, err := bson.Marshal(test.Data)
			assert.NoError(t, err)
			var got BsonTestStruct

			err = bson.Unmarshal(res, &got)
			assert.NoError(t, err)
			snaps.MatchSnapshot(t, got)
		})

	}

}
