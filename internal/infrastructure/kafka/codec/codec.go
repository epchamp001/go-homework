package codec

import "encoding/json"

const ContentTypeJSON = "application/json"

type Codec interface {
	Marshal(v any) ([]byte, error)
	Unmarshal(data []byte, v any) error
	ContentType() string
}

type JSONCodec struct{}

func (JSONCodec) Marshal(v any) ([]byte, error)      { return json.Marshal(v) }
func (JSONCodec) Unmarshal(data []byte, v any) error { return json.Unmarshal(data, v) }
func (JSONCodec) ContentType() string                { return ContentTypeJSON }

var JSON Codec = JSONCodec{}
