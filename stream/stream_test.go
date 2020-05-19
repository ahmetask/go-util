package stream

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type testModel struct {
	Id   int
	Name string
	List []testModel
	Map  map[string]testModel
}

var _ = Describe("Test Stream", func() {
	Describe("Of", func() {
		Context("It should stream", func() {
			It("array", func() {
				arr := []testModel{
					{Id: 1, Name: "a"},
					{Id: 2, Name: "b"},
					{Id: 3, Name: "c"},
					{Id: 4, Name: "d"},
				}

				streamList := Of(arr)

				Expect(streamList).NotTo(BeNil())
			})
			It("map", func() {
				arr := map[string]testModel{
					"a": {Id: 1, Name: "a"},
					"b": {Id: 2, Name: "b"},
					"c": {Id: 3, Name: "c"},
					"d": {Id: 4, Name: "d"},
				}

				streamList := Of(arr)

				Expect(streamList).NotTo(BeNil())
			})
		})
	})
})
