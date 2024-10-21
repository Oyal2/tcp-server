package executor_test

import (
	"os"
	"testing"

	"github.com/Oyal2/tcp-server/test/helper"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var printerPath string

var _ = BeforeSuite(func() {
	var err error
	printerPath, err = helper.BuildPrinterExecutable()
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	if printerPath != "" {
		err := os.Remove(printerPath)
		if err != nil {
			GinkgoWriter.Printf("Failed to remove printer executable: %v\n", err)
		}
	}
})

func TestExecutor(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Executor Suite")
}
