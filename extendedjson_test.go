// Copyright (c) 2020 - Adrien Petel

package mongoextjson_test

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/feliixx/mongoextjson"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestMarshalUnmarshal(t *testing.T) {

	t.Parallel()
	objectID, _ := primitive.ObjectIDFromHex("5a934e000102030405000000")

	marshalTests := []struct {
		name          string
		value         interface{}
		data          string
		canonical     string
		skipMarshal   bool
		skipUnmarshal bool
	}{
		{
			name:      "objectID",
			value:     objectID,
			data:      `ObjectId("5a934e000102030405000000")`,
			canonical: `{"$oid":"5a934e000102030405000000"}`,
		},
		{
			name:          "DateTime",
			value:         primitive.DateTime(778846633334),
			data:          `ISODate("1994-09-06T10:17:13.334Z")`,
			canonical:     `{"$date":"1994-09-06T10:17:13.334Z"}`,
			skipUnmarshal: true, // what is this new primitive.DateTime time ?
		},
		{
			name:      "Timestamp",
			value:     primitive.Timestamp{T: 1, I: 2},
			data:      `Timestamp(1,2)`,
			canonical: `{"$timestamp":{"t":1,"i":2}}`,
		},
		{
			name:      "time.Date UTC",
			value:     time.Date(2016, 5, 15, 1, 2, 3, 4000000, time.UTC),
			data:      `ISODate("2016-05-15T01:02:03.004Z")`,
			canonical: `{"$date":"2016-05-15T01:02:03.004Z"}`,
		}, {
			name:          "time.Date with zone",
			value:         time.Date(2016, 5, 15, 1, 2, 3, 4000000, time.FixedZone("CET", 60*60)),
			data:          `ISODate("2016-05-15T01:02:03.004+01:00")`,
			canonical:     `{"$date":"2016-05-15T01:02:03.004+01:00"}`,
			skipUnmarshal: true, // TODO: why this doesn't work ?
		},
		{
			name:      "Binary",
			value:     primitive.Binary{Subtype: 2, Data: []byte("foo")},
			data:      `BinData(2,"Zm9v")`,
			canonical: `{"$binary":"Zm9v","$type":"0x2"}`,
		},
		{
			name:      "Undefined",
			value:     primitive.Undefined{},
			data:      `undefined`,
			canonical: `{"$undefined":true}`,
		},
		{
			name:      "Decimal 128",
			value:     primitive.NewDecimal128(3385858588484, 3333),
			data:      `NumberDecimal("6.2458066851535814488338301193477E-6145")`,
			canonical: `{"$numberDecimal":"6.2458066851535814488338301193477E-6145"}`,
		},
		{
			name:      "string",
			value:     bson.M{"str": "hello"},
			data:      `{"str":"hello"}`,
			canonical: `{"str":"hello"}`,
		},
		{
			name:      "single quoted string",
			value:     bson.M{"str": "'hello'"},
			data:      `{"str":"'hello'"}`,
			canonical: `{"str":"'hello'"}`,
		},
		{
			name:      "double quoted string",
			value:     bson.M{"str": "\"hello\""},
			data:      `{"str":"\"hello\""}`,
			canonical: `{"str":"\"hello\""}`,
		},
		{
			name: "string with line return",
			value: bson.M{"str": `"he
			llo"`},
			data:      `{"str":"\"he\n\t\t\tllo\""}`,
			canonical: `{"str":"\"he\n\t\t\tllo\""}`,
		},
		{
			name:      "int64",
			value:     int64(10),
			data:      `10`,
			canonical: `{"$numberLong":10}`,
		},
		{
			name:      "int",
			value:     int(1),
			data:      `1`,
			canonical: `1`,
		},
		{
			name:      "int32",
			value:     int32(26),
			data:      `NumberInt(26)`,
			canonical: `{"$numberInt":26}`,
		},
		{
			name:      "float32",
			value:     float32(2.32),
			data:      `2.32`,
			canonical: `2.32`,
		},
		{
			name:      "float64",
			value:     float64(2.6464),
			data:      `2.6464`,
			canonical: `2.6464`,
		},
		{
			name:      "regex",
			value:     primitive.Regex{Pattern: "/test/", Options: "i"},
			data:      `{"$regex":"/test/","$options":"i"}`,
			canonical: `{"$regex":"/test/","$options":"i"}`,
		},
		{
			name:      "object",
			value:     bson.M{"key": "value"},
			data:      `{"key":"value"}`,
			canonical: `{"key":"value"}`,
		},
		{
			name:        "object with unquoted keys",
			value:       bson.M{"key": "value", "obj": bson.M{"sub": 1, "f": 0.0}},
			data:        `{key :"value",obj:{sub:1,f:0.0}}`,
			canonical:   `{key :"value",obj:{sub:1,f:0.0}}`,
			skipMarshal: true,
		},
		{
			name:      "empty object",
			value:     bson.M{},
			data:      `{}`,
			canonical: `{}`,
		},
		{
			name:      "boolean true",
			value:     true,
			data:      `true`,
			canonical: `true`,
		},
		{
			name:      "boolean false",
			value:     false,
			data:      `false`,
			canonical: `false`,
		},
		{
			name:          "array with null value",
			value:         []bson.M{nil},
			data:          `[null]`,
			canonical:     `[null]`,
			skipUnmarshal: true,
		},
		{
			name:      "empty array",
			value:     []bson.M{},
			data:      `[]`,
			canonical: `[]`,
		},
		{
			name:      "object with array",
			value:     bson.M{"key": bson.A{"one", "two"}},
			data:      `{"key":["one","two"]}`,
			canonical: `{"key":["one","two"]}`,
		},
		{
			name:      "array of objects",
			value:     []bson.M{bson.M{"k": "v1"}, bson.M{"k": "v2"}},
			data:      `[{"k":"v1"},{"k":"v2"}]`,
			canonical: `[{"k":"v1"},{"k":"v2"}]`,
		},
		{
			name:          "min key",
			value:         bson.M{"k": primitive.MinKey{}},
			data:          `{"k":{}}`, // TODO: is this normal ?
			canonical:     `{"k":{}}`,
			skipUnmarshal: true,
		},
		{
			name:          "max key",
			value:         bson.M{"k": primitive.MaxKey{}},
			data:          `{"k":{}}`, // TODO: is this normal ?
			canonical:     `{"k":{}}`,
			skipUnmarshal: true,
		},
		{
			name:          "DBRef",
			value:         primitive.DBPointer{DB: "test", Pointer: objectID},
			data:          `{"DB":"test","Pointer":ObjectId("5a934e000102030405000000")}`,
			canonical:     `{"DB":"test","Pointer":{"$oid":"5a934e000102030405000000"}}`,
			skipUnmarshal: true,
		},
		{
			name:        "data with space",
			value:       bson.M{"key": bson.A{"one", "two"}},
			data:        `{ "key" : [ "one", "two" ] }`,
			canonical:   `{ "key"  :["one","two"]}`,
			skipMarshal: true,
		},
		{
			name:  "data with line return",
			value: bson.M{"key": bson.A{1, 2}},
			data: `{ 
				"key" : [ 
					1,
					2
				]
			}`,
			canonical: `{
				 "key"  :[1,2
				 ]}`,
			skipMarshal: true,
		},
		{
			name:  "data with tab",
			value: bson.M{"key": bson.A{"one", "two"}},
			data: `{ "key"	:	["one",	"two"]	}`,
			canonical: `{	"key":[	"one","two"]}`,
			skipMarshal: true,
		},
		{
			name:  "bson data with tab",
			value: bson.M{"key": bson.A{objectID, int32(0)}},
			data: `{ "key"	:	[ObjectId("5a934e000102030405000000"),	NumberInt(0) ]	}`,
			canonical: `{	"key":[	{"$oid":"5a934e000102030405000000"},{"$numberInt":0} ] }`,
			skipMarshal: true,
		},
	}

	for _, tt := range marshalTests {
		t.Run(tt.name, func(t *testing.T) {

			if !tt.skipMarshal {
				data, err := mongoextjson.Marshal(tt.value)
				if err != nil {
					t.Errorf("fail to marshal %v: %v", tt.value, err)
				}
				if want, got := tt.data, string(data); want != got {
					t.Errorf("marshal failed: expected %s, but got %s", want, got)
				}

				data, err = mongoextjson.MarshalCanonical(tt.value)
				if err != nil {
					t.Errorf("fail to marshal canonical %v: %v", tt.value, err)
				}
				if want, got := tt.canonical, string(data); want != got {
					t.Errorf("marshal canonical failed: expected %s, but got %s", want, got)
				}
			}

			if !tt.skipUnmarshal {

				value := reflect.New(reflect.TypeOf(tt.value)).Elem().Interface()
				err := mongoextjson.Unmarshal([]byte(tt.data), &value)
				if err != nil {
					t.Errorf("fail to unmarshal %s: %v", tt.data, err)
				}
				if want, got := fmt.Sprintf("%v", tt.value), fmt.Sprintf("%v", value); want != got {
					t.Errorf("unmarshal failed: expected %v, but got %v", want, got)
				}

				value = reflect.New(reflect.TypeOf(tt.value)).Elem().Interface()
				err = mongoextjson.Unmarshal([]byte(tt.canonical), &value)
				if err != nil {
					t.Errorf("fail to unmarshal canonical %s: %v", tt.data, err)
				}
				if want, got := fmt.Sprintf("%v", tt.value), fmt.Sprintf("%v", value); want != got {
					t.Errorf("unmarshal canonical failed: expected %v, but got %v", want, got)
				}

			}
		})
	}
}
