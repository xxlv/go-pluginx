package utils

import (
	"reflect"
	"testing"
)

func TestRemoveElements(t *testing.T) {
	original := []string{"apple", "banana", "cherry", "date", "elderberry"}
	toRemove := []string{"banana", "date"}

	expected := []string{"apple", "cherry", "elderberry"}
	result := RemoveElements(original, toRemove)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, but got %v", expected, result)
	}
}

func TestRemoveDuplicates(t *testing.T) {
	elements := []string{"apple", "banana", "cherry", "apple", "banana", "banana", "date", "elderberry", "date"}

	expected := []string{"apple", "banana", "cherry", "date", "elderberry"}
	result := RemoveDuplicates(elements)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, but got %v", expected, result)
	}
}

func TestStringInList(t *testing.T) {
	myList := []string{"apple", "banana", "orange"}
	
	searchString1 := "banana"
	if !StringInList(searchString1, myList) {
		t.Errorf("%s should exist in the list", searchString1)
	}

	searchString2 := "grape"
	if StringInList(searchString2, myList) {
		t.Errorf("%s should not exist in the list", searchString2)
	}

}
