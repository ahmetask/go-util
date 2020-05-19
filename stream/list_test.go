package stream

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test Stream", func() {
	Describe("List", func() {
		var testArray []testModel
		BeforeEach(func() {
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
		})
		Context("when stream.Interface", func() {
			It("should get an interface", func() {
				res := Of(testArray).Interface()
				v, ok := res.([]testModel)
				Expect(ok).Should(BeTrue())
				Expect(testArray).Should(Equal(v))
			})
		})
		Context("when stream", func() {
			It("should apply filter", func() {
				res := Of(testArray).Filter(func(content Content) bool {
					return content.Data.(testModel).Name == testArray[2].Name
				})
				v, ok := res.Interface().([]testModel)

				Expect(ok).To(Equal(true))
				Expect(len(v)).To(Equal(1))
				Expect(v[0]).To(Equal(testArray[2]))
			})
			It("should apply multiple filter", func() {
				res := Of(testArray).Filter(func(content Content) bool {
					return content.Data.(testModel).Id > 5
				}).Filter(func(content Content) bool {
					return content.Data.(testModel).Name == "i"
				})
				v, ok := res.Interface().([]testModel)

				Expect(ok).To(Equal(true))
				Expect(len(v)).To(Equal(1))
				Expect(v[0].Name).To(Equal("i"))
			})
			It("should apply multiple filter and map to new list", func() {
				res := Of(testArray).Filter(func(content Content) bool {
					return content.Data.(testModel).Id > 5
				}).Filter(func(content Content) bool {
					return content.Data.(testModel).Name == "i"
				}).Map(func(content Content) Content {
					return Content{Data: content.Data.(testModel).Name}
				}, []string{})
				v, ok := res.Interface().([]string)

				Expect(ok).To(Equal(true))
				Expect(len(v)).To(Equal(1))
				Expect(v[0]).To(Equal("i"))
			})
			It("should apply multiple filter and parallel map to new list", func() {
				res := Of(testArray).Filter(func(content Content) bool {
					return content.Data.(testModel).Id > 5
				}).Filter(func(content Content) bool {
					return content.Data.(testModel).Id < 9
				}).Map(func(content Content) Content {
					return Content{Data: content.Data.(testModel).Name}
				}, []string{}, 4)
				v, ok := res.Interface().([]string)

				Expect(ok).To(Equal(true))
				Expect(len(v)).To(Equal(3))
				for _, i := range v {
					Expect(i == "h" || i == "g" || i == "f").To(BeTrue())
				}
			})
			It("should apply multiple filter parallel and parallel map to new list", func() {
				res := Of(testArray).Filter(func(content Content) bool {
					return content.Data.(testModel).Id > 5
				}).Filter(func(content Content) bool {
					return content.Data.(testModel).Id < 9
				}, 4).Map(func(content Content) Content {
					return Content{Data: content.Data.(testModel).Name}
				}, []string{}, 4)
				v, ok := res.Interface().([]string)

				Expect(ok).To(Equal(true))
				Expect(len(v)).To(Equal(3))
				for _, i := range v {
					Expect(i == "h" || i == "g" || i == "f").To(BeTrue())
				}
			})
			It("should apply filter and map to new map", func() {
				res := Of(testArray).
					Filter(func(content Content) bool {
						return content.Data.(testModel).Id > 5
					}).
					Map(func(content Content) Content {
						c := content.Data.(testModel)

						return Content{Key: c.Name, Data: c.Name}
					}, map[string]string{})

				v, ok := res.Interface().(map[string]string)
				Expect(ok).To(Equal(true))
				Expect(len(v)).To(Equal(4))
				Expect(v["f"]).To(Equal("f"))
				Expect(v["g"]).To(Equal("g"))
				Expect(v["h"]).To(Equal("h"))
				Expect(v["i"]).To(Equal("i"))
			})
			It("should apply filter and parallel map to new map", func() {
				res := Of(testArray).
					Filter(func(content Content) bool {
						return content.Data.(testModel).Id > 5
					}).
					Map(func(content Content) Content {
						c := content.Data.(testModel)

						return Content{Key: c.Name, Data: c.Name}
					}, map[string]string{}, 4)

				v, ok := res.Interface().(map[string]string)
				Expect(ok).To(Equal(true))
				Expect(len(v)).To(Equal(4))
				Expect(v["f"]).To(Equal("f"))
				Expect(v["g"]).To(Equal("g"))
				Expect(v["h"]).To(Equal("h"))
				Expect(v["i"]).To(Equal("i"))
			})
			It("should apply skip", func() {
				res := Of(testArray).
					Skip(2)
				v, ok := res.Interface().([]testModel)

				Expect(ok).To(Equal(true))
				Expect(len(v)).To(Equal(7))
			})
			It("should apply limit", func() {
				res := Of(testArray).
					Limit(2)
				v, ok := res.Interface().([]testModel)

				Expect(ok).To(Equal(true))
				Expect(len(v)).To(Equal(2))
			})
			It("should apply skip, limit and sort", func() {
				res := Of(testArray).
					Skip(2).
					Limit(4).
					SortBy(func(content Content, content2 Content) int {
						if content.Data.(testModel).Id < content2.Data.(testModel).Id {
							return 1
						}
						if content.Data.(testModel).Id == content2.Data.(testModel).Id {
							return 0
						}
						return -1
					})
				v, ok := res.Interface().([]testModel)
				Expect(ok).To(Equal(true))
				Expect(len(v)).To(Equal(2))
				Expect(v[0]).To(Equal(testArray[4]))
				Expect(v[1]).To(Equal(testArray[5]))
			})
			It("should apply filter skip, limit, map and sort", func() {
				res := Of(testArray).
					Skip(1).
					Limit(6).
					Filter(func(content Content) bool {
						return content.Data.(testModel).Id > 3
					}).
					Map(func(content Content) Content {
						c := content.Data.(testModel)

						return Content{Data: c.Id}
					}, []int{}, 4).
					SortBy(func(content Content, content2 Content) int {
						if content.Data.(int) < content2.Data.(int) {
							return 1
						}
						if content.Data.(int) == content2.Data.(int) {
							return 0
						}
						return -1
					})
				v, ok := res.Interface().([]int)
				Expect(ok).To(Equal(true))
				Expect(len(v)).To(Equal(4))
				Expect(v[0]).To(Equal(4))
				Expect(v[1]).To(Equal(5))
				Expect(v[2]).To(Equal(6))
				Expect(v[3]).To(Equal(7))
			})
			It("should apply filter skip, limit, map and any match true", func() {
				res := Of(testArray).
					Skip(2).
					Limit(9).
					Filter(func(content Content) bool {
						return content.Data.(testModel).Id > 3
					}).
					Map(func(content Content) Content {
						c := content.Data.(testModel)

						return Content{Data: c.Id}
					}, []int{}, 4).
					AnyMatch(func(content Content) bool {
						return content.Data.(int) == 8
					})

				Expect(res).To(Equal(true))
			})
			It("should apply filter skip, limit, map and any match false", func() {
				res := Of(testArray).
					Skip(2).
					Limit(9).
					Filter(func(content Content) bool {
						return content.Data.(testModel).Id > 3
					}).
					Map(func(content Content) Content {
						c := content.Data.(testModel)

						return Content{Data: c.Id}
					}, []int{}, 4).
					AnyMatch(func(content Content) bool {
						return content.Data.(int) == 1
					})

				Expect(res).To(Equal(false))
			})
			It("should apply filter map and all match", func() {
				res := Of(testArray).
					Filter(func(content Content) bool {
						return content.Data.(testModel).Id > 3
					}).
					Map(func(content Content) Content {
						c := content.Data.(testModel)

						return Content{Data: c.Id}
					}, []int{}, 4).
					AllMatch(func(content Content) bool {
						return content.Data.(int) > 3
					})

				Expect(res).To(Equal(true))
			})
			It("should apply filter map and all match false", func() {
				res := Of(testArray).
					Filter(func(content Content) bool {
						return content.Data.(testModel).Id > 3
					}).
					Map(func(content Content) Content {
						c := content.Data.(testModel)

						return Content{Data: c.Id}
					}, []int{}, 4).
					AllMatch(func(content Content) bool {
						return content.Data.(int) < 6
					})

				Expect(res).To(Equal(false))
			})
			It("should apply map and find min", func() {
				res := Of(testArray).
					Map(func(content Content) Content {
						c := content.Data.(testModel)

						return Content{Data: c.Id}
					}, []int{}, 4).
					FindEdge(func(content Content, content2 Content) bool {
						return content.Data.(int) < content2.Data.(int)
					})
				v, ok := res.(int)
				Expect(ok).To(Equal(true))
				Expect(v).To(Equal(1))
			})
			It("should apply map and find max", func() {
				res := Of(testArray).
					Map(func(content Content) Content {
						c := content.Data.(testModel)

						return Content{Data: c.Id}
					}, []int{}, 4).
					FindEdge(func(content Content, content2 Content) bool {
						return content.Data.(int) > content2.Data.(int)
					})
				v, ok := res.(int)
				Expect(ok).To(Equal(true))
				Expect(v).To(Equal(9))
			})
			It("should get count", func() {
				res := Of(testArray).
					Map(func(content Content) Content {
						c := content.Data.(testModel)

						return Content{Data: c.Id}
					}, []int{}, 4).
					Count()
				Expect(res).To(Equal(9))
			})
			It("should get first", func() {
				res := Of(testArray).
					Map(func(content Content) Content {
						c := content.Data.(testModel)

						return Content{Data: c.Id}
					}, []int{}).
					SortBy(func(content Content, content2 Content) int {
						if content.Data.(int) < content2.Data.(int) {
							return 1
						}
						if content.Data.(int) == content2.Data.(int) {
							return 0
						}
						return -1
					}).FindFirst()
				Expect(res.(int)).To(Equal(1))
			})
			It("should get last", func() {
				res := Of(testArray).
					Map(func(content Content) Content {
						c := content.Data.(testModel)

						return Content{Data: c.Id}
					}, []int{}).
					SortBy(func(content Content, content2 Content) int {
						if content.Data.(int) < content2.Data.(int) {
							return 1
						}
						if content.Data.(int) == content2.Data.(int) {
							return 0
						}
						return -1
					}).FindLast()
				Expect(res.(int)).To(Equal(9))
			})
			It("should filter and map internal array", func() {
				internalArray := []testModel{
					{Id: 2, Name: "b"},
					{Id: 3, Name: "c"},
				}
				testArray = []testModel{
					{Id: 1, Name: "a", List: internalArray},
				}
				res := Of(testArray).
					Map(func(content Content) Content {
						c := content.Data.(testModel)

						return Content{Data: c.List}
					}, []testModel{})

				v, ok := res.Interface().([]testModel)
				Expect(ok).To(Equal(true))
				Expect(len(v)).To(Equal(2))
				Expect(v[0]).To(Equal(testArray[0].List[0]))
				Expect(v[1]).To(Equal(testArray[0].List[1]))
			})
			It("should filter and map internal array", func() {
				internalArray := []testModel{
					{Id: 2, Name: "b"},
					{Id: 3, Name: "c"},
				}
				testArray = []testModel{
					{Id: 1, Name: "a", List: internalArray},
				}
				res := Of(testArray).
					Map(func(content Content) Content {
						c := content.Data.(testModel)

						return Content{Data: c.List}
					}, []testModel{})

				v, ok := res.Interface().([]testModel)
				Expect(ok).To(Equal(true))
				Expect(len(v)).To(Equal(2))
				Expect(v[0]).To(Equal(testArray[0].List[0]))
				Expect(v[1]).To(Equal(testArray[0].List[1]))
			})
			It("should filter and map internal array parallel", func() {
				internalMap := map[string]testModel{
					"x": {Id: 2, Name: "b", List: []testModel{{Name: "internal"}, {Name: "internal2"}}},
					"y": {Id: 3, Name: "c", List: []testModel{{Name: "internal3"}}},
				}
				testArray = []testModel{
					{Id: 1, Name: "a", Map: internalMap},
					{Id: 1, Name: "b", Map: internalMap},
				}
				res := Of(testArray).
					Filter(func(content Content) bool {
						return content.Data.(testModel).Name == "a"
					}).
					Map(func(content Content) Content {
						c := content.Data.(testModel)
						return Content{Key: "map", Data: c.Map}
					}, map[string]testModel{}, 4).
					Map(func(content Content) Content {
						c := content.Data.(testModel)
						return Content{Data: c.List}
					}, []testModel{}, 4)

				v, ok := res.Interface().([]testModel)
				Expect(ok).To(Equal(true))
				Expect(len(v)).To(Equal(3))
				for _, v := range v {
					Expect(v.Name == "internal" || v.Name == "internal2" || v.Name == "internal3").To(BeTrue())
				}
			})
			It("should filter and map internal array", func() {
				internalMap := map[string]testModel{
					"x": {Id: 2, Name: "b", List: []testModel{{Name: "internal"}, {Name: "internal2"}}},
					"y": {Id: 3, Name: "c", List: []testModel{{Name: "internal3"}}},
				}
				testArray = []testModel{
					{Id: 1, Name: "a", Map: internalMap},
					{Id: 1, Name: "b", Map: internalMap},
				}
				res := Of(testArray).
					Filter(func(content Content) bool {
						return content.Data.(testModel).Name == "a"
					}).
					Map(func(content Content) Content {
						c := content.Data.(testModel)
						return Content{Key: "map", Data: c.Map}
					}, map[string]testModel{}).
					Map(func(content Content) Content {
						c := content.Data.(testModel)
						return Content{Data: c.List}
					}, []testModel{})

				v, ok := res.Interface().([]testModel)
				Expect(ok).To(Equal(true))
				Expect(len(v)).To(Equal(3))
				for _, v := range v {
					Expect(v.Name == "internal" || v.Name == "internal2" || v.Name == "internal3").To(BeTrue())
				}
			})

			It("should filter and map internal map", func() {
				internalMap := map[string]testModel{
					"x": {Id: 2, Name: "b", Map: map[string]testModel{"a": {Name: "t1"}, "c": {Name: "t2"}}},
					"y": {Id: 3, Name: "c", Map: map[string]testModel{"t": {Name: "y1"}, "t2": {Name: "y2"}}},
				}
				testArray = []testModel{
					{Id: 1, Name: "a", Map: internalMap},
					{Id: 1, Name: "b", Map: internalMap},
				}
				res := Of(testArray).
					Filter(func(content Content) bool {
						return content.Data.(testModel).Name == "a"
					}).
					Map(func(content Content) Content {
						c := content.Data.(testModel)
						return Content{Key: "map", Data: c.Map}
					}, map[string]testModel{}, 3).
					Map(func(content Content) Content {
						c := content.Data.(testModel)
						k := content.Key.(string)
						return Content{Key: k, Data: c.Map}
					}, map[string]testModel{}).
					Map(func(content Content) Content {
						c := content.Data.(testModel)
						return Content{Data: c.Name}
					}, []string{}, 4)

				v, ok := res.Interface().([]string)
				Expect(ok).To(Equal(true))
				Expect(len(v)).To(Equal(4))
				for _, s := range v {
					Expect(s == "t1" || s == "t2" || s == "y1" || s == "y2").To(BeTrue())
				}
			})
		})
	})
})
