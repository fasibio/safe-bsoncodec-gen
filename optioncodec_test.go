package safebsoncodecgen_test

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/fasibio/safe"
	safebsoncodecgen "github.com/fasibio/safe-bsoncodec-gen"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
)

type BsonTestStruct struct {
	A int                              `bson:"a" json:"a,omitempty"`
	B safe.Option[int]                 `bson:"b" json:"b,omitempty"`
	C string                           `bson:"c" json:"c,omitempty"`
	F safe.Option[BsonSubTestStruct]   `bson:"f" json:"f,omitempty"`
	G safe.Option[[]BsonSubTestStruct] `bson:"g"`
}

type BsonSubTestStruct struct {
	D safe.Option[string] `bson:"d,omitempty" json:"d,omitempty"`
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

	reg := bson.NewRegistry()

	reg.RegisterTypeDecoder(reflect.TypeOf(safe.Option[string]{}), safebsoncodecgen.OptionCodec[string]{})
	reg.RegisterTypeEncoder(reflect.TypeOf(safe.Option[string]{}), safebsoncodecgen.OptionCodec[string]{})

	reg.RegisterTypeDecoder(reflect.TypeOf(safe.Option[int]{}), safebsoncodecgen.OptionCodec[int]{})
	reg.RegisterTypeEncoder(reflect.TypeOf(safe.Option[int]{}), safebsoncodecgen.OptionCodec[int]{})

	reg.RegisterTypeDecoder(reflect.TypeOf(safe.Option[BsonSubTestStruct]{}), safebsoncodecgen.OptionCodec[BsonSubTestStruct]{})
	reg.RegisterTypeEncoder(reflect.TypeOf(safe.Option[BsonSubTestStruct]{}), safebsoncodecgen.OptionCodec[BsonSubTestStruct]{})
	reg.RegisterTypeDecoder(reflect.TypeOf(safe.Option[[]BsonSubTestStruct]{}), safebsoncodecgen.OptionCodec[[]BsonSubTestStruct]{})
	reg.RegisterTypeEncoder(reflect.TypeOf(safe.Option[[]BsonSubTestStruct]{}), safebsoncodecgen.OptionCodec[[]BsonSubTestStruct]{})

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			require := require.New(t)
			buf := new(bytes.Buffer)
			vw, err := bsonrw.NewBSONValueWriter(buf)
			if err != nil {
				panic(err)
			}
			enc, err := bson.NewEncoder(vw)
			if err != nil {
				panic(err)
			}
			err = enc.SetRegistry(reg)
			require.NoError(err)
			err = enc.Encode(test.Data)
			// res, err := bson.MarshalWithRegistry(reg, test.Data)
			require.NoError(err)
			var got BsonTestStruct

			dec, err := bson.NewDecoder(bsonrw.NewBSONDocumentReader(buf.Bytes()))
			if err != nil {
				panic(err)
			}
			err = dec.SetRegistry(reg)
			require.NoError(err)
			err = dec.Decode(&got)
			require.NoError(err)

			// err = bson.UnmarshalWithRegistry(reg, res, &got)
			snaps.MatchSnapshot(t, got)
		})

	}

}
