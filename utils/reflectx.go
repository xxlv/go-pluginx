package utils

import (
	"reflect"
	"strings"
)

func ParseTag(tag reflect.StructTag) map[string]string {
	// 空格问题
	tagMap := make(map[string]string)

	// Split the tag string into slices
	tags := strings.Split(string(tag), " ")

	for _, v := range tags {
		// Split each tag into key and value
		tagParts := strings.SplitN(v, ":", 2)
		if len(tagParts) == 2 {
			// Remove quotes from the value, then add to the map
			key := tagParts[0]
			value := tag.Get(key)
			tagMap[key] = value
		}
	}

	return tagMap
}
