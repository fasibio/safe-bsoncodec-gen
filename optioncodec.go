package safebsoncodecgen

import (
	"fmt"
	"reflect"

	"github.com/fasibio/safe"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
)

type OptionCodec[T any] struct{}

func (oc OptionCodec[T]) EncodeValue(
	ctx bsoncodec.EncodeContext, vw bsonrw.ValueWriter, val reflect.Value,
) error {
	if opt, ok := val.Interface().(safe.Option[T]); ok {
		if opt.IsNone() {
			return vw.WriteNull()
		}
		en, err := bson.NewEncoder(vw)
		if err != nil {
			return err
		}
		err = en.SetRegistry(ctx.Registry)
		if err != nil {
			return err
		}
		return en.Encode(opt.Unwrap())
	}
	return fmt.Errorf("type not match only safe.Option are allowed ")
}

func (oc OptionCodec[T]) DecodeValue(
	ctx bsoncodec.DecodeContext, vr bsonrw.ValueReader, val reflect.Value,
) error {
	if vr.Type() == bson.TypeNull {
		err := vr.ReadNull()
		if err != nil {
			return err
		}
		val.Set(reflect.ValueOf(safe.None[T]()))
		return nil
	}

	var v T
	de, err := bson.NewDecoder(vr)
	if err != nil {
		return err
	}
	err = de.SetRegistry(ctx.Registry)
	if err != nil {
		return err
	}
	err = de.Decode(&v)
	if err != nil {
		return err
	}

	val.Set(reflect.ValueOf(safe.Some(&v)))
	return nil
}
