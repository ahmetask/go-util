package stream

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"runtime"
)

var _ = Describe("Test Util", func() {
	Describe("getThreadCount", func() {
		Context("It should get thread count", func() {
			It("when empty arg", func() {
				count := getThreadCount()
				Expect(count).To(Equal(1))
			})
			It("when arg exist and less than maxCore", func() {
				maxCore := runtime.NumCPU()
				count := getThreadCount(maxCore - 1)
				Expect(count).To(Equal(maxCore - 1))
			})
			It("when arg exist and greater than maxCore", func() {
				maxCore := runtime.NumCPU()
				count := getThreadCount(maxCore + 1)
				Expect(count).To(Equal(maxCore))
			})
		})
	})
})
