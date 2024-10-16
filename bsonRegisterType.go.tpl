package {{.PackageName}}

import (
  "reflect"
  "github.com/fasibio/safe"
  safebsoncodecgen "github.com/fasibio/safe-bsoncodec-gen"
  "go.mongodb.org/mongo-driver/bson/bsoncodec"
  {{- range $value := .OptionTypes.GetPackagePaths .PackagePath}}
      "{{$value}}"
  {{- end}}
)

func RegisterOptionCodec(registry *bsoncodec.Registry){
  {{- range $key,$value := .OptionTypes}}
      {{- $snippet := call $value.GetCodecSnippet }}
      registry.RegisterTypeDecoder(
        reflect.TypeOf(safe.Option[{{ $snippet }}]{}), 
        safebsoncodecgen.OptionCodec[{{ $snippet }}]{},
      )
      registry.RegisterTypeEncoder(
        reflect.TypeOf(safe.Option[{{ $snippet }}]{}), 
        safebsoncodecgen.OptionCodec[{{ $snippet }}]{},
      )
  {{- end}}
}