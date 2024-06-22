package utils

import (
	"reflect"
	"testing"
)

type TestStruct struct {
	Name string `json:"name" xml:"name"`
	Age  int    `json:"age"`
}

func TestParseTag(t *testing.T) {
	// Get the Type of the TestStruct
	typ := reflect.TypeOf(TestStruct{})

	// Get the field "Name"
	field, _ := typ.FieldByName("Name")

	// Parse the tag of the field
	tagMap := ParseTag(field.Tag)

	// Check if the "json" key is "name"
	if tagMap["json"] != "name" {
		t.Errorf("Expected 'name', but got '%s'", tagMap["json"])
	}

	// Check if the "xml" key is "name"
	if tagMap["xml"] != "name" {
		t.Errorf("Expected 'name', but got '%s'", tagMap["xml"])
	}

	// Get the field "Age"
	field, _ = typ.FieldByName("Age")

	// Parse the tag of the field
	tagMap = ParseTag(field.Tag)

	// Check if the "json" key is "age"
	if tagMap["json"] != "age" {
		t.Errorf("Expected 'age', but got '%s'", tagMap["json"])
	}

	// Check if the "xml" key is not present
	if _, ok := tagMap["xml"]; ok {
		t.Errorf("Expected no 'xml' key, but got '%s'", tagMap["xml"])
	}
}
