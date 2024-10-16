package safebsoncodecgen

import (
	"reflect"

	"github.com/fasibio/safe"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
)

// Custom Codec for Option[T] to integrate with BSON operations.
type OptionCodec[T any] struct{}

// EncodeValue encodes the Option[T] for BSON storage.
func (oc OptionCodec[T]) EncodeValue(
	ctx bsoncodec.EncodeContext, vw bsonrw.ValueWriter, val reflect.Value,
) error {
	opt := val.Interface().(safe.Option[T])
	if opt.IsNone() {
		return vw.WriteNull()
	}
	// Use bson.Marshal to encode the inner value.
	en, err := bson.NewEncoder(vw)
	if err != nil {
		return err
	}
	return en.Encode(opt.Unwrap())
}

// DecodeValue decodes the BSON data into Option[T].
func (oc OptionCodec[T]) DecodeValue(
	ctx bsoncodec.DecodeContext, vr bsonrw.ValueReader, val reflect.Value,
) error {
	if vr.Type() == bson.TypeNull {
		vr.ReadNull()
		val.Set(reflect.ValueOf(safe.None[T]()))
		return nil
	}

	// Decode the inner value.
	var v T
	de, err := bson.NewDecoder(vr)
	if err != nil {
		return err
	}
	de.Decode(&v)

	val.Set(reflect.ValueOf(safe.Some(&v)))
	return nil
}
