package stream

type Content struct {
	Key  interface{} // for map type
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
