package executor_test

import (
	"context"
	"time"

	"github.com/Oyal2/tcp-server/internal/constant"
	"github.com/Oyal2/tcp-server/internal/model"
	"github.com/Oyal2/tcp-server/pkg/executor"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CommandExecutor", func() {
	var exe *executor.CommandExecutor

	BeforeEach(func() {
		exe = executor.NewCommandExecutor()
	})

	It("should execute a simple command", func() {
		request := &model.TaskRequest{
			Command: []string{printerPath},
			Timeout: 1000,
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(request.Timeout)*time.Millisecond)
		defer cancel()
		result := exe.ExecuteTask(ctx, request)

		Expect(result.ExitCode).To(Equal(0))
		Expect(result.Output).To(Equal("test\n"))
		Expect(result.Error).To(BeEmpty())
	})

	It("should handle command timeout", func() {
		request := &model.TaskRequest{
			Command: []string{"sleep", "2"},
			Timeout: 1000,
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(request.Timeout)*time.Millisecond)
		defer cancel()
		result := exe.ExecuteTask(ctx, request)

		Expect(result.ExitCode).To(Equal(-1))
		Expect(result.Error).NotTo(BeEmpty())
	})

	It("should handle non-existent command", func() {
		request := &model.TaskRequest{
			Command: []string{"fake_command"},
			Timeout: 1000,
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(request.Timeout)*time.Millisecond)
		defer cancel()
		result := exe.ExecuteTask(ctx, request)

		Expect(result.ExitCode).To(Equal(-1))
		Expect(result.Error).NotTo(BeEmpty())
	})

	It("should handle nil command", func() {
		request := &model.TaskRequest{
			Command: nil,
			Timeout: 1000,
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(request.Timeout)*time.Millisecond)
		defer cancel()
		result := exe.ExecuteTask(ctx, request)

		Expect(result.ExitCode).To(Equal(-1))
		Expect(result.Error).NotTo(BeEmpty())
		Expect(result.Error).To(Equal(constant.TaskResultCommandNilError))
	})

	It("should handle empty command", func() {
		request := &model.TaskRequest{
			Command: []string{""},
			Timeout: 1000,
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(request.Timeout)*time.Millisecond)
		defer cancel()
		result := exe.ExecuteTask(ctx, request)

		Expect(result.ExitCode).To(Equal(-1))
		Expect(result.Error).NotTo(BeEmpty())
		Expect(result.Error).To(Equal("exec: no command"))
	})

})
