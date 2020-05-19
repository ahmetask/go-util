# Go-util

Go-util is a Golang library for iterating arrays or maps by using reflection

## Installation

You need [golang 1.13](https://golang.org/dl/) to use go-util.


```go
go get github.com/ahmetask/go-util
```

## Functions
```go
stream.Of(data interface)
.Filter(f Filter, threadCount ...int) IStream
.Map(f Action, newType interface{}, threadCount ...int) IStream
.Skip(i int) IStream
.Limit(i int) IStream
.SortBy(f Compare) IStream
.FindEdge(f CompareConditional) interface{}
.Count() int
.AnyMatch(f Filter) bool
.AllMatch(f Filter) bool
.FindFirst() interface{}
.FindLast() interface{}
.Interface() interface{}

```
## Action Functions and model

```go
type Content struct {
    Key interface{}
    Data interface{}
}

// Filter function
type Filter func(Content) bool

// Map Function
type Action func(Content) Content

// Sorting function
type Compare func(Content, Content) int

// Min Max Function
type CompareConditional func(Content, Content) bool

```

## Usage
- You can do lots of things like skip, limit, min, max, allMatch etc.
- The key point is that managing interface correctly otherwise it panics
- It only supports array, slice and maps
- Mapping does not provide sortBy since order of keys changes in runtime.
- Library provides synchronized option, but it changes order of list. So use if you don't need order and execution of action 
takes too much time
- You can see combinations in tests or example

Wrap Your Data
```go
    stream.Of(yourData)
```
And iterate over by using stream library functions
```go
    stream.Of(yourData).Filter(f).Map(f).SortBy(f).FindEdge(f)....
```

## Example
```go
package main

import (
	"fmt"
	"github.com/ahmetask/go-util/stream"
)

func main() {
	type St struct {
		Name string
	}

	type MyStruct struct {
		Name         string
		Integer      int
		InternalList []St
		internalMap  map[string]St
	}

	var mapping = map[string]MyStruct{
		"keyA": {
			Name:    "stringA",
			Integer: 1,
			InternalList: []St{
				{
					Name: "internalListA1",
				}, {
					Name: "internalListA2",
				},
				{
					Name: "a",
				},
			},
			internalMap: map[string]St{
				"internalKeyA": {
					Name: "internalMapA1",
				},
				"internalKeyA2": {
					Name: "internalMapA2",
				},
			},
		},
		"keyB": {
			Name:    "stringB",
			Integer: 1,
			InternalList: []St{
				{
					Name: "internalListB1",
				}, {
					Name: "internalListB2",
				},
			},
			internalMap: map[string]St{
				"internalKeyB": {
					Name: "internalMapB1",
				},
				"internalKeyB2": {
					Name: "internalMapB2",
				},
			},
		},
	}

	var st = stream.Of(mapping).Interface()

	fmt.Printf("---Interface:\n%v\n", st)

	var arr = stream.Of(mapping).
		Map(func(content stream.Content) stream.Content {
			return stream.Content{Data: content.Data.(MyStruct).Name}
		}, []string{}).Interface()

	fmt.Printf("---Mapping\n%v\n", arr)

	var filteredArray = stream.Of(mapping).
		Filter(func(content stream.Content) bool {
			return content.Key.(string) == "keyA"
		}).
		Map(func(content stream.Content) stream.Content {
			return stream.Content{Data: content.Data.(MyStruct).Name}
		}, []string{}).Interface()

	fmt.Printf("---Filtered Mapping\n%v\n", filteredArray)

	var filteredInternalArray = stream.Of(mapping).
		Filter(func(content stream.Content) bool {
			return content.Key.(string) == "keyA"
		}).
		Map(func(content stream.Content) stream.Content {
			return stream.Content{Data: content.Data.(MyStruct).InternalList}
		}, []St{}).
		Map(func(content stream.Content) stream.Content {
			return stream.Content{Data: content.Data.(St).Name}
		}, []string{}).
		Interface()

	fmt.Printf("---Filtered Internal List Mapping\n%v\n", filteredInternalArray)

	var filteredInternalMapping = stream.Of(mapping).
		Filter(func(content stream.Content) bool {
			return content.Key.(string) == "keyA"
		}).
		Map(func(content stream.Content) stream.Content {
			return stream.Content{Key: content.Key.(string), Data: content.Data.(MyStruct).internalMap}
		}, map[string]St{}).
		Interface()

	fmt.Printf("---Filtered Internal Map Mapping\n%v\n", filteredInternalMapping)

	var sortedInternalArray = stream.Of(mapping).
		Filter(func(content stream.Content) bool {
			return content.Key.(string) == "keyA"
		}).
		Map(func(content stream.Content) stream.Content {
			return stream.Content{Data: content.Data.(MyStruct).InternalList}
		}, []St{}).
		Map(func(content stream.Content) stream.Content {
			return stream.Content{Data: content.Data.(St).Name}
		}, []string{}).
		SortBy(func(content stream.Content, content2 stream.Content) int {
			if content.Data.(string) < content2.Data.(string) {
				return 1
			}
			if content.Data.(string) == content2.Data.(string) {
				return 0
			}
			return -1
		}).Interface()

	fmt.Printf("---Sorted Internal List Mapping\n%v\n", sortedInternalArray)
}
```

## TODO 
- more generic utility

## Contributing
Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.