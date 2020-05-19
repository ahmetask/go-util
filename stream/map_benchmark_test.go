package stream

import (
	"testing"
	"time"
)

var testMap map[string]testModel

func init() {
	testMap = map[string]testModel{
		"a": {Id: 1, Name: "a"},
		"b": {Id: 2, Name: "b"},
		"c": {Id: 3, Name: "c"},
		"d": {Id: 4, Name: "d"},
		"e": {Id: 5, Name: "e"},
		"f": {Id: 6, Name: "f"},
		"g": {Id: 7, Name: "g"},
		"h": {Id: 8, Name: "h"},
		"i": {Id: 9, Name: "i"},
	}
}

func BenchmarkMapping_Filter(b *testing.B) {
	stream := Of(testMap)
	for n := 0; n < b.N; n++ {
		stream.Filter(func(content Content) bool {
			time.Sleep(100 * time.Millisecond)
			return content.Data.(testModel).Name == "c"
		}).Interface()
	}
}

func BenchmarkMapping_FilterParallel(b *testing.B) {
	stream := Of(testArray)
	for n := 0; n < b.N; n++ {
		stream.Filter(func(content Content) bool {
			time.Sleep(100 * time.Millisecond)
			return content.Data.(testModel).Name == "c"
		}, 4).Interface()
	}
}

func BenchmarkMapping_Map(b *testing.B) {
	stream := Of(testArray)
	for n := 0; n < b.N; n++ {
		stream.Map(func(content Content) Content {
			time.Sleep(10 * time.Millisecond)
			c := content.Data.(testModel)

			return Content{Data: c.Name}
		}, []string{})
	}
}

func BenchmarkMapping_MapParallel(b *testing.B) {
	stream := Of(testArray)
	for n := 0; n < b.N; n++ {
		stream.Map(func(content Content) Content {
			time.Sleep(10 * time.Millisecond)
			c := content.Data.(testModel)
			return Content{Data: c.Name}
		}, []string{}, 4)
	}
}
