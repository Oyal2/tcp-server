package ipratelimit_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestIpratelimit(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ipratelimit Suite")
}
