package util_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stv0g/cunicu/pkg/util"
)

var _ = Context("reindent json", func() {
	It("works", func() {
		original := []byte(`{ "a": { "b": { "c": 5 } } }`)
		indented := []byte(`{
  "a": {
    "b": {
      "c": 5
    }
  }
}`)

		Expect(util.ReIndentJSON(original, "", "  ")).To(Equal(indented))
	})

	It("fails for invalid json", func() {
		_, err := util.ReIndentJSON([]byte("{"), "", "  ")
		Expect(err).To(HaveOccurred())
	})
})
