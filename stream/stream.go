package stream

import (
	"math"
	"reflect"
)

type IStream interface {
	Filter(f Filter, threadCount ...int) IStream
	Map(f Action, newType interface{}, threadCount ...int) IStream
	Skip(i int) IStream
	Limit(i int) IStream
	SortBy(f Compare) IStream
	FindEdge(f CompareConditional) interface{}
	Count() int
	AnyMatch(f Filter) bool
	AllMatch(f Filter) bool
	FindFirst() interface{}
	FindLast() interface{}
	Interface() interface{}
}

func Of(data interface{}) IStream {
	v := reflect.ValueOf(data)
	kind := v.Kind()
	switch kind {
	case reflect.Slice:
		fallthrough
	case reflect.Array:
		stream := &list{
			format: reflect.TypeOf(data),
			kind:   kind,
			items:  v,
		}
		return stream
	case reflect.Map:
		stream := &mapping{
			format: reflect.TypeOf(data),
			kind:   kind,
			items:  v,
		}
		return stream
	default:
		panic("it should be slice,array or map")
	}
	return nil
}

func parallelListToList(l *list, newType reflect.Type, f Action) IStream {
	newContent := reflect.MakeSlice(newType, 0, 0)
	length := l.items.Len()
	chunkSize := int(math.Ceil(float64(length) / float64(l.workerCount)))

	c := make(chan reflect.Value, l.workerCount)

	worker := func(result chan reflect.Value, st, end int) {
		newContent := reflect.MakeSlice(newType, 0, 0)
		for i := st; i < end; i++ {
			content := Content{
				Data: l.items.Index(i).Interface(),
			}
			c := f(content)
			v := reflect.ValueOf(c.Data)
			kind := v.Kind()
			if kind == reflect.Slice || kind == reflect.Array {
				newContent = reflect.AppendSlice(newContent, v)
			} else {
				newContent = reflect.Append(newContent, v)
			}
		}
		result <- newContent
	}

	for i := 0; i < l.workerCount; i++ {
		end := (i + 1) * chunkSize
		if end > length {
			end = length
		}
		go worker(c, i*chunkSize, end)
	}

	for i := 0; i < l.workerCount; i++ {
		newContent = reflect.AppendSlice(newContent, <-c)
	}

	stream := &list{
		format: newType,
		kind:   reflect.Slice,
		items:  newContent,
	}
	return stream
}

func listToList(l *list, newType reflect.Type, f Action) IStream {
	if l.workerCount > 1 {
		return parallelListToList(l, newType, f)
	}

	newContent := reflect.MakeSlice(newType, 0, 0)
	for i := 0; i < l.items.Len(); i++ {
		content := Content{
			Data: l.items.Index(i).Interface(),
		}
		c := f(content)
		v := reflect.ValueOf(c.Data)
		kind := v.Kind()
		if kind == reflect.Slice || kind == reflect.Array {
			newContent = reflect.AppendSlice(newContent, v)
		} else {
			newContent = reflect.Append(newContent, v)
		}
	}

	stream := &list{
		format: newType,
		kind:   reflect.Slice,
		items:  newContent,
	}
	return stream
}

func parallelListToMap(l *list, newType reflect.Type, f Action) IStream {
	newContent := reflect.MakeMap(newType)
	c := make(chan []keyValue, l.workerCount)
	length := l.items.Len()

	worker := func(result chan []keyValue, st, end int) {
		var newContent []keyValue
		for i := st; i < end; i++ {
			content := Content{
				Data: l.items.Index(i).Interface(),
			}
			c := f(content)
			k := reflect.ValueOf(c.Key)
			v := reflect.ValueOf(c.Data)
			kind := v.Kind()
			if kind == reflect.Map {
				for _, ik := range v.MapKeys() {
					newContent = append(newContent, keyValue{Key: ik, Value: v.MapIndex(ik)})
				}
			} else {
				newContent = append(newContent, keyValue{Key: k, Value: v})
			}

		}
		result <- newContent
	}

	chunkSize := int(math.Ceil(float64(length) / float64(l.workerCount)))
	for i := 0; i < l.workerCount; i++ {
		end := (i + 1) * chunkSize
		if end > length {
			end = length
		}
		go worker(c, i*chunkSize, end)
	}

	for i := 0; i < l.workerCount; i++ {
		keyValues := <-c
		for _, kv := range keyValues {
			newContent.SetMapIndex(kv.Key, kv.Value)
		}
	}

	stream := &mapping{
		format: newType,
		kind:   reflect.Map,
		items:  newContent,
	}

	return stream
}

func listToMap(l *list, newType reflect.Type, f Action) IStream {
	if l.workerCount > 1 {
		return parallelListToMap(l, newType, f)
	}

	newContent := reflect.MakeMap(newType)
	for i := 0; i < l.items.Len(); i++ {
		content := Content{
			Data: l.items.Index(i).Interface(),
		}
		c := f(content)
		k := reflect.ValueOf(c.Key)
		v := reflect.ValueOf(c.Data)
		kind := v.Kind()
		if kind == reflect.Map {
			for _, ik := range v.MapKeys() {
				newContent.SetMapIndex(ik, v.MapIndex(ik))
			}
		} else {
			newContent.SetMapIndex(k, v)
		}
	}

	stream := &mapping{
		format: newType,
		kind:   reflect.Map,
		items:  newContent,
	}
	return stream
}

func parallelMapToMap(m *mapping, newType reflect.Type, f Action) IStream {
	newContent := reflect.MakeMap(newType)
	keys := m.items.MapKeys()
	c := make(chan []keyValue, m.workerCount)
	length := len(keys)

	worker := func(result chan []keyValue, st, end int) {
		var newContent []keyValue
		for i := st; i < end; i++ {
			content := Content{
				Key:  keys[i].Interface(),
				Data: m.items.MapIndex(keys[i]).Interface(),
			}
			c := f(content)
			k := reflect.ValueOf(c.Key)
			v := reflect.ValueOf(c.Data)

			kind := v.Kind()
			if kind == reflect.Map {
				for _, ik := range v.MapKeys() {
					newContent = append(newContent, keyValue{Key: ik, Value: v.MapIndex(ik)})
				}
			} else {
				newContent = append(newContent, keyValue{Key: k, Value: v})
			}
		}
		result <- newContent
	}

	chunkSize := int(math.Ceil(float64(length) / float64(m.workerCount)))
	for i := 0; i < m.workerCount; i++ {
		end := (i + 1) * chunkSize
		if end > length {
			end = length
		}
		go worker(c, i*chunkSize, end)
	}

	for i := 0; i < m.workerCount; i++ {
		keyValues := <-c
		for _, kv := range keyValues {
			newContent.SetMapIndex(kv.Key, kv.Value)
		}
	}

	stream := &mapping{
		format: newType,
		kind:   reflect.Map,
		items:  newContent,
	}

	return stream
}

func mapToMap(m *mapping, newType reflect.Type, f Action) IStream {
	if m.workerCount > 1 {
		return parallelMapToMap(m, newType, f)
	}

	newContent := reflect.MakeMap(newType)
	for _, key := range m.items.MapKeys() {
		content := Content{
			Key:  key.Interface(),
			Data: m.items.MapIndex(key).Interface(),
		}
		c := f(content)
		k := reflect.ValueOf(c.Key)
		v := reflect.ValueOf(c.Data)

		kind := v.Kind()
		if kind == reflect.Map {
			for _, ik := range v.MapKeys() {
				newContent.SetMapIndex(ik, v.MapIndex(ik))
			}
		} else {
			newContent.SetMapIndex(k, v)
		}
	}

	stream := &mapping{
		format: newType,
		kind:   reflect.Map,
		items:  newContent,
	}
	return stream
}

func parallelMapToList(m *mapping, newType reflect.Type, f Action) IStream {
	newContent := reflect.MakeSlice(newType, 0, 0)
	keys := m.items.MapKeys()
	c := make(chan reflect.Value, m.workerCount)
	length := len(keys)

	worker := func(result chan reflect.Value, st, end int) {
		newContent := reflect.MakeSlice(newType, 0, 0)
		for i := st; i < end; i++ {
			content := Content{
				Data: m.items.MapIndex(keys[i]).Interface(),
			}
			c := f(content)
			v := reflect.ValueOf(c.Data)
			kind := v.Kind()
			if kind == reflect.Slice || kind == reflect.Array {
				newContent = reflect.AppendSlice(newContent, v)
			} else {
				newContent = reflect.Append(newContent, v)
			}
		}
		result <- newContent
	}

	chunkSize := int(math.Ceil(float64(length) / float64(m.workerCount)))
	for i := 0; i < m.workerCount; i++ {
		end := (i + 1) * chunkSize
		if end > length {
			end = length
		}
		go worker(c, i*chunkSize, end)
	}

	for i := 0; i < m.workerCount; i++ {
		newContent = reflect.AppendSlice(newContent, <-c)
	}

	stream := &list{
		format: newType,
		kind:   reflect.Slice,
		items:  newContent,
	}
	return stream

}

func mapToList(m *mapping, newType reflect.Type, f Action) IStream {
	if m.workerCount > 1 {
		return parallelMapToList(m, newType, f)
	}

	newContent := reflect.MakeSlice(newType, 0, 0)
	for _, key := range m.items.MapKeys() {
		content := Content{
			Key:  key.Interface(),
			Data: m.items.MapIndex(key).Interface(),
		}
		c := f(content)
		v := reflect.ValueOf(c.Data)
		kind := v.Kind()
		if kind == reflect.Slice || kind == reflect.Array {
			newContent = reflect.AppendSlice(newContent, v)
		} else {
			newContent = reflect.Append(newContent, v)
		}
	}

	stream := &list{
		format: newType,
		kind:   reflect.Slice,
		items:  newContent,
	}
	return stream
}
