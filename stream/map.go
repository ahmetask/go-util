package stream

import (
	"math"
	"reflect"
	"sync"
)

type mapping struct {
	items          reflect.Value //holds data
	workerCount    int
	filters        []Filter //filters array
	allMatchFilter Filter   //alMatch Filter
	allMatched     bool
	skip           int // skip first x element(s)
	limit          int
	findEdge       bool               // min max
	fFindEdge      CompareConditional //min max function
	selectedKey    reflect.Value      // selected item
	selectedValue  reflect.Value      //items kind
	kind           reflect.Kind       //items kind
	format         reflect.Type       //items kind
	iMutex         sync.RWMutex       //mutex that used in parallel execution
}

type keyValue struct {
	Key   reflect.Value
	Value reflect.Value
}

func (s *mapping) parallelProcess() {
	// make new map
	newContent := reflect.MakeMap(s.format)
	// get keys as array
	keys := s.items.MapKeys()
	// make channel that hold key and value
	c := make(chan []keyValue, s.workerCount)
	s.allMatched = true

	worker := func(result chan []keyValue, st, end int) {
		var newContent []keyValue

		for i := st; i < end; i++ {
			key := keys[i]
			item := s.items.MapIndex(key)

			currentContent := Content{
				Key:  key.Interface(),
				Data: item.Interface(),
			}

			currentKeyValue := keyValue{
				Key:   key,
				Value: item,
			}

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

				newContent = append(newContent, currentKeyValue)
			}
		}

		result <- newContent
	}

	// skip limit init
	length := len(keys)
	start := 0
	if s.skip > 0 && s.skip < length {
		start = s.skip
	}

	if s.limit > 0 && s.limit+start < length {
		length = s.limit + start
	}

	// chunk size according to keys[worker(chunk), worker1(chunk1), ...]
	chunkSize := int(math.Ceil(float64(length-start) / float64(s.workerCount)))
	for i := 0; i < s.workerCount; i++ {
		end := start + (i+1)*chunkSize
		if end > length {
			end = length
		}
		go worker(c, start+i*chunkSize, end)
	}

	for i := 0; i < s.workerCount; i++ {
		keyValues := <-c
		for _, kv := range keyValues {
			newContent.SetMapIndex(kv.Key, kv.Value)
		}
	}

	s.items = newContent
}

func (s *mapping) process() {
	if s.workerCount > 1 {
		s.parallelProcess()
		return
	}

	newContent := reflect.MakeMap(s.format)

	keys := s.items.MapKeys()
	length := len(keys)
	start := 0
	if s.skip > 0 && s.skip < length {
		start = s.skip
	}

	if s.limit > 0 && s.limit+start < length {
		length = s.limit + start
	}

	for i := start; i < length; i++ {
		key := keys[i]
		item := s.items.MapIndex(key)

		currentContent := Content{
			Key:  key.Interface(),
			Data: item.Interface(),
		}

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

			if !s.selectedValue.CanSet() {
				s.selectedKey = key
				s.selectedValue = item
			}

			selectedContent := Content{
				Key:  s.selectedKey.Interface(),
				Data: s.selectedValue.Interface(),
			}

			if s.findEdge && s.fFindEdge(currentContent, selectedContent) {
				s.selectedKey = key
				s.selectedValue = item
			}
			newContent.SetMapIndex(key, item)
		}
	}

	s.items = newContent
}

func (s *mapping) Filter(f Filter, threadCount ...int) IStream {
	s.workerCount = getThreadCount(threadCount...)
	s.filters = append(s.filters, f)

	return s
}

func (s *mapping) Map(f Action, newType interface{}, threadCount ...int) IStream {
	s.workerCount = getThreadCount(threadCount...)
	s.process()
	v := reflect.ValueOf(newType)
	kind := v.Kind()
	typeOf := reflect.TypeOf(newType)
	if s.kind == reflect.Map && (kind == reflect.Slice || kind == reflect.Array) {
		return mapToList(s, typeOf, f)
	} else if s.kind == reflect.Map && kind == reflect.Map {
		return mapToMap(s, typeOf, f)
	} else {
		panic("newType should be slice,array or map")
	}
	return s
}

// since key order change in run time it is not advised
func (s *mapping) Skip(i int) IStream {
	s.skip = i
	return s
}

// since key order change in run time it is not advised
func (s *mapping) Limit(i int) IStream {
	s.limit = i
	return s
}

//it is not suitable since key order change in run time.
func (s *mapping) SortBy(_ Compare) IStream {
	return s
}

func (s *mapping) FindEdge(f CompareConditional) interface{} {
	s.findEdge = true
	s.fFindEdge = f
	s.process()
	if s.selectedValue.CanSet() {
		return s.selectedValue.Interface()
	}
	return nil
}

func (s *mapping) Count() int {
	return len(s.items.MapKeys())
}

func (s *mapping) AnyMatch(f Filter) bool {
	s.filters = append(s.filters, f)
	s.process()
	return len(s.items.MapKeys()) > 0
}

func (s *mapping) AllMatch(f Filter) bool {
	s.allMatchFilter = f
	s.process()
	return s.allMatched
}

//it gives random result since MapKeys list order changes in runtime
func (s *mapping) FindFirst() interface{} {
	s.process()
	if len(s.items.MapKeys()) > 0 {
		keys := s.items.MapKeys()
		return s.items.MapIndex(keys[0])
	}
	return nil
}

//it gives random result since MapKeys list order changes in runtime
func (s *mapping) FindLast() interface{} {
	s.process()
	if len(s.items.MapKeys()) > 0 {
		keys := s.items.MapKeys()
		return s.items.MapIndex(keys[len(keys)-1])
	}
	return nil
}

func (s *mapping) Interface() interface{} {
	s.process()
	return s.items.Interface()
}
