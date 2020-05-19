package stream

import (
	"math"
	"reflect"
	"sort"
	"sync"
)

type list struct {
	items          reflect.Value // holds data
	workerCount    int
	filters        []Filter // filters array
	allMatchFilter Filter // allMatch filter
	allMatched     bool
	skip           int // skip first x element(s)
	limit          int
	findEdge       bool // min max
	fFindEdge      CompareConditional //min max function
	selectedValue  reflect.Value // selected item
	kind           reflect.Kind //items kind
	format         reflect.Type //items format
	iMutex         sync.RWMutex  //mutex that used in parallel execution
}

func (s *list) parallelProcess() {
	// make new slice
	newContent := reflect.MakeSlice(s.format, 0, 0)

	// skip limit init
	length := s.items.Len()
	start := 0
	if s.skip > 0 && s.skip < length {
		start = s.skip
	}

	if s.limit > 0 && s.limit+start < length {
		length = s.limit + start
	}

	// all matched flag
	s.allMatched = true

	// chunk size [worker(chunk), worker1(chunk1), ...]
	chunkSize := int(math.Ceil(float64(s.items.Len()-start) / float64(s.workerCount)))

	// worker func
	worker := func(result chan reflect.Value, st, end int) {
		newContent := reflect.MakeSlice(s.format, 0, 0)

		for i := st; i < end; i++ {
			item := s.items.Index(i)

			currentContent := Content{
				Data: item.Interface(),
			}

			// apply filter
			ok := true
			for _, f := range s.filters {
				ok = ok && f(currentContent)
				if !ok {
					break
				}
			}

			if ok {
				s.iMutex.Lock()

				if s.allMatchFilter != nil {
					if !s.allMatchFilter(currentContent) {
						s.allMatched = false
						s.allMatchFilter = nil
					}
				}

				if !s.selectedValue.CanSet() {
					s.selectedValue = item
				}

				selectedContent := Content{
					Data: s.selectedValue.Interface(),
				}

				if s.findEdge && s.fFindEdge(currentContent, selectedContent) {
					s.selectedValue = item
				}

				s.iMutex.Unlock()

				kind := item.Kind()
				// slice in slice
				if kind == reflect.Slice || kind == reflect.Array {
					newContent = reflect.AppendSlice(newContent, item)
				} else {
					newContent = reflect.Append(newContent, item)
				}
			}
		}

		result <- newContent
	}

	c := make(chan reflect.Value, s.workerCount)

	for i := 0; i < s.workerCount; i++ {
		end := start + (i+1)*chunkSize
		if end > length {
			end = length
		}

		go worker(c, start+i*chunkSize, end)
	}

	for i := 0; i < s.workerCount; i++ {
		newContent = reflect.AppendSlice(newContent, <-c)
	}

	s.items = newContent
}

func (s *list) process() {
	if s.workerCount > 1 {
		s.parallelProcess()

		return
	}

	// make new slice
	newContent := reflect.MakeSlice(s.format, 0, 0)

	// skip and limit adjustments
	length := s.items.Len()
	start := 0
	if s.skip > 0 && s.skip < length {
		start = s.skip
	}

	if s.limit > 0 && s.limit+start < length {
		length = s.limit + start
	}

	s.allMatched = true

	for i := start; i < length; i++ {
		item := s.items.Index(i)

		currentContent := Content{
			Data: item.Interface(),
		}

		// apply filters
		ok := true
		for _, f := range s.filters {
			ok = ok && f(currentContent)
			if !ok {
				break
			}
		}

		if ok {
			if s.allMatchFilter != nil {
				if !s.allMatchFilter(currentContent) {
					s.allMatched = false
					s.allMatchFilter = nil
				}
			}

			// for findEdge function
			if !s.selectedValue.CanSet() {
				s.selectedValue = item
			}

			selectedContent := Content{
				Data: s.selectedValue.Interface(),
			}

			if s.findEdge && s.fFindEdge(currentContent, selectedContent) {
				s.selectedValue = item
			}

			kind := item.Kind()
			// slice in slice
			if kind == reflect.Slice || kind == reflect.Array {
				newContent = reflect.AppendSlice(newContent, item)
			} else {
				newContent = reflect.Append(newContent, item)
			}
		}
	}

	s.items = newContent
}

// append new filter
// thread count optional default is one. More thread breaks order of list items
// use multiple thread if filter function execution takes too much time and order is not important
func (s *list) Filter(f Filter, threadCount ...int) IStream {
	s.workerCount = getThreadCount(threadCount...)
	s.filters = append(s.filters, f)

	return s
}

// process filter and apply action to all item
// new type required and it should be array slice or map
// thread count optional default is one. More thread breaks order of list items
// use multiple thread if Action function execution takes too much time and order is not important
func (s *list) Map(f Action, newType interface{}, threadCount ...int) IStream {
	s.workerCount = getThreadCount(threadCount...)

	s.process()

	v := reflect.ValueOf(newType)
	kind := v.Kind()
	typeOf := reflect.TypeOf(newType)

	if (s.kind == reflect.Slice || s.kind == reflect.Array) && (kind == reflect.Slice || kind == reflect.Array) {
		return listToList(s, typeOf, f)
	} else if (s.kind == reflect.Slice || s.kind == reflect.Array) && kind == reflect.Map {
		return listToMap(s, typeOf, f)
	} else {
		panic("newType should be slice,array or map")
	}

	return s
}

// skip first i elements
func (s *list) Skip(i int) IStream {
	s.skip = i

	return s
}

// read i element from start
func (s *list) Limit(i int) IStream {
	s.limit = i

	return s
}

// sorting
func (s *list) SortBy(f Compare) IStream {
	s.process()
	sort.Slice(s.items.Interface(), s.makeLess(f))

	return s
}

func (s *list) makeLess(f Compare) func(i, j int) bool {
	return func(x, y int) bool {
		c1 := Content{
			Data: s.items.Index(x).Interface(),
		}

		c2 := Content{
			Data: s.items.Index(y).Interface(),
		}

		comp := f(c1, c2)

		return comp > 0
	}
}

// min max
func (s *list) FindEdge(f CompareConditional) interface{} {
	s.findEdge = true
	s.fFindEdge = f

	s.process()

	if s.selectedValue.CanSet() {
		return s.selectedValue.Interface()
	}

	return nil
}

// list size
func (s *list) Count() int {
	return s.items.Len()
}

func (s *list) AnyMatch(f Filter) bool {
	s.filters = append(s.filters, f)
	s.process()

	return s.items.Len() > 0
}

func (s *list) AllMatch(f Filter) bool {
	s.allMatchFilter = f
	s.process()

	return s.allMatched
}

func (s *list) FindFirst() interface{} {
	s.process()

	if s.items.Len() > 0 {
		first := s.items.Index(0)

		return first.Interface()
	}

	return nil
}

func (s *list) FindLast() interface{} {
	s.process()

	if s.items.Len() > 0 {
		last := s.items.Index(s.items.Len() - 1)

		return last.Interface()
	}

	return nil
}

func (s *list) Interface() interface{} {
	s.process()

	return s.items.Interface()
}
