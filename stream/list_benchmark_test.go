package stream

import (
	"testing"
	"time"
)

var testArray []testModel

func init() {
	testArray = []testModel{
		{Id: 1, Name: "a"},
		{Id: 2, Name: "b"},
		{Id: 3, Name: "c"},
		{Id: 4, Name: "d"},
		{Id: 5, Name: "e"},
		{Id: 6, Name: "f"},
		{Id: 7, Name: "g"},
		{Id: 8, Name: "h"},
		{Id: 9, Name: "i"},
	}
}

func BenchmarkList_Filter(b *testing.B) {
	stream := Of(testArray)
	for n := 0; n < b.N; n++ {
		stream.Filter(func(content Content) bool {
			time.Sleep(100 * time.Millisecond)
			return content.Data.(testModel).Name == "c"
		}).Interface()
	}
}

func BenchmarkList_FilterParallel(b *testing.B) {
	stream := Of(testArray)
	for n := 0; n < b.N; n++ {
		stream.Filter(func(content Content) bool {
			time.Sleep(100 * time.Millisecond)
			return content.Data.(testModel).Name == "c"
		}, 4).Interface()
	}
}

func BenchmarkList_Map(b *testing.B) {
	stream := Of(testArray)
	for n := 0; n < b.N; n++ {
		stream.Map(func(content Content) Content {
			time.Sleep(10 * time.Millisecond)
			c := content.Data.(testModel)

			return Content{Data: c.Name}
		}, []string{})
	}
}

func BenchmarkList_MapParallel(b *testing.B) {
	stream := Of(testArray)
	for n := 0; n < b.N; n++ {
		stream.Map(func(content Content) Content {
			time.Sleep(10 * time.Millisecond)
			c := content.Data.(testModel)
			return Content{Data: c.Name}
		}, []string{}, 4)
	}
}
