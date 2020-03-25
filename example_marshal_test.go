package mongoextjson_test

import (
	"fmt"
	"time"

	"github.com/feliixx/mongoextjson"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func ExampleMarshal() {

	objectID, _ = primitive.ObjectIDFromHex("5a934e000102030405000000")

	doc := bson.M{
		"_id":        objectID,
		"string":     "string",
		"int32":      int32(32),
		"int64":      int64(64),
		"double":     2.2,
		"decimal128": primitive.NewDecimal128(1, 1),
		"false":      false,
		"true":       true,
		"binary":     primitive.Binary{Subtype: 2, Data: []byte("binary")},
		"date":       time.Date(2016, 5, 15, 1, 2, 3, 4000000, time.UTC),
		"timestamp":  primitive.Timestamp{T: 12, I: 0},
		"undefined":  primitive.Undefined{},
	}
	b, err := mongoextjson.Marshal(doc)
	if err != nil {
		fmt.Printf("fail to marshal %+v: %v", doc, err)
	}
	fmt.Printf("%s", b)
	// Output:
	//{"_id":ObjectId("5a934e000102030405000000"),"binary":BinData(2,"YmluYXJ5"),"date":ISODate("2016-05-15T01:02:03.004Z"),"decimal128":NumberDecimal("1.8446744073709551617E-6157"),"double":2.2,"false":false,"int32":32,"int64":NumberLong(64),"string":"string","timestamp":Timestamp(12,0),"true":true,"undefined":undefined}
}
