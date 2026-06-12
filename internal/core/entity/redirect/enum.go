package redirect

import (
	"encoding/json"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

type Type int

const (
	PROXY    Type = iota
	REDIRECT Type = iota
	IFRAME   Type = iota
)

var typeNames = map[Type]string{
	PROXY:    "PROXY",
	REDIRECT: "REDIRECT",
	IFRAME:   "IFRAME",
}

var typeValues = map[string]Type{
	"PROXY":    PROXY,
	"REDIRECT": REDIRECT,
	"IFRAME":   IFRAME,
}

func (t Type) String() string {
	if name, ok := typeNames[t]; ok {
		return name
	}
	return fmt.Sprintf("UNKNOWN(%d)", int(t))
}

func (t Type) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

func (t *Type) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err != nil {
		return err
	}
	val, ok := typeValues[name]
	if !ok {
		return fmt.Errorf("unknown redirect type: %s", name)
	}
	*t = val
	return nil
}

func (t Type) MarshalBSONValue() (bsontype.Type, []byte, error) {
	return bsontype.String, bsoncore.AppendString(nil, t.String()), nil
}

func (t *Type) UnmarshalBSONValue(bt bsontype.Type, data []byte) error {
	if bt != bsontype.String {
		return fmt.Errorf("expected BSON string, got %v", bt)
	}
	str, _, ok := bsoncore.ReadString(data)
	if !ok {
		return fmt.Errorf("failed to read BSON string for redirect type")
	}
	val, ok := typeValues[str]
	if !ok {
		return fmt.Errorf("unknown redirect type: %s", str)
	}
	*t = val
	return nil
}
