package mongoextjson_test

import (
	"fmt"

	"github.com/feliixx/mongoextjson"
	"go.mongodb.org/mongo-driver/bson"
)

func ExampleUnmarshal() {

	doc := bson.M{}
	data := []byte(`
	{
		"_id":        ObjectId("5a934e000102030405000000"),
		"binary":     BinData(2,"YmluYXJ5"),
		"date":       ISODate("2016-05-15T01:02:03.004Z"),
		"decimal128": NumberDecimal("1.8446744073709551617E-6157"),
		"double":     2.2,
		"false":      false,
		"int32":      32,
		"int64":      NumberLong(64),
		"string":     "string",
		"timestamp":  Timestamp(12,0),
		"true":       true,
		"undefined":  undefined,
		unquoted:     "keys can be unquoted"
	}`)
	err := mongoextjson.Unmarshal(data, &doc)
	if err != nil {
		fmt.Printf("fail to unmarshal %+v: %v", data, err)
	}
	fmt.Printf("%+v", doc)
	// Output:
	//map[_id:ObjectID("5a934e000102030405000000") binary:{Subtype:2 Data:[98 105 110 97 114 121]} date:2016-05-15 01:02:03.004 +0000 UTC decimal128:1.8446744073709551617E-6157 double:2.2 false:false int32:32 int64:64 string:string timestamp:{T:12 I:0} true:true undefined:{} unquoted:keys can be unquoted]
}