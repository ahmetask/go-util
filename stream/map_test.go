package stream

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test Stream", func() {
	Describe("Map", func() {
		var testMap map[string]testModel
		BeforeEach(func() {
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
		})
		Context("when stream.Interface", func() {
			It("should get an interface", func() {
				res := Of(testMap).Interface()
				v, ok := res.(map[string]testModel)
				Expect(ok).Should(BeTrue())
				Expect(testMap).Should(Equal(v))
			})
		})
		Context("when stream", func() {
			It("should apply filter", func() {
				res := Of(testMap).Filter(func(content Content) bool {
					return content.Data.(testModel).Name == testMap["c"].Name
				})
				v, ok := res.Interface().(map[string]testModel)

				Expect(ok).To(Equal(true))
				Expect(len(v)).To(Equal(1))
				Expect(v["c"]).To(Equal(testMap["c"]))
			})
			It("should apply multiple filter", func() {
				res := Of(testMap).Filter(func(content Content) bool {
					return content.Data.(testModel).Id > 5
				}).Filter(func(content Content) bool {
					return content.Data.(testModel).Name == "i"
				})
				v, ok := res.Interface().(map[string]testModel)

				Expect(ok).To(Equal(true))
				Expect(len(v)).To(Equal(1))
				Expect(v["i"].Name).To(Equal("i"))
			})

			It("should apply multiple filter and map to new list", func() {
				res := Of(testMap).Filter(func(content Content) bool {
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
				res := Of(testMap).Filter(func(content Content) bool {
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
				res := Of(testMap).Filter(func(content Content) bool {
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
				res := Of(testMap).
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
				res := Of(testMap).
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
				res := Of(testMap).
					Skip(2)
				v, ok := res.Interface().(map[string]testModel)

				Expect(ok).To(Equal(true))
				Expect(len(v)).To(Equal(7))
			})
			It("should apply limit", func() {
				res := Of(testMap).
					Limit(2)
				v, ok := res.Interface().(map[string]testModel)

				Expect(ok).To(Equal(true))
				Expect(len(v)).To(Equal(2))
			})
			It("should apply filter skip, limit, map", func() {
				res := Of(testMap).
					Filter(func(content Content) bool {
						return content.Data.(testModel).Id < 3
					}).
					Map(func(content Content) Content {
						c := content.Data.(testModel)

						return Content{Data: c.Id}
					}, []int{}, 4)
				v, ok := res.Interface().([]int)
				Expect(ok).To(Equal(true))
				Expect(len(v)).To(Equal(2))
				Expect(v[0] == 2 || v[0] == 1).To(BeTrue())
			})
			It("should apply filter map and any match true", func() {
				res := Of(testMap).
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
				res := Of(testMap).
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
			It("should apply filter map ann all match", func() {
				res := Of(testMap).
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
			It("should apply filter map ann all match false", func() {
				res := Of(testMap).
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
				res := Of(testMap).
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
				res := Of(testMap).
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
				res := Of(testMap).
					Map(func(content Content) Content {
						c := content.Data.(testModel)

						return Content{Data: c.Id}
					}, []int{}, 4).
					Count()
				Expect(res).To(Equal(9))
			})
			It("should get first", func() {
				res := Of(testMap).
					Map(func(content Content) Content {
						c := content.Data.(testModel)

						return Content{Data: c.Id}
					}, []int{}).
					FindFirst()
				Expect(res.(int)).ShouldNot(BeNil())
			})
			It("should get last", func() {
				res := Of(testMap).
					Map(func(content Content) Content {
						c := content.Data.(testModel)

						return Content{Data: c.Id}
					}, []int{}).
					FindLast()
				Expect(res.(int)).ShouldNot(BeNil())
			})
		})
	})
})
