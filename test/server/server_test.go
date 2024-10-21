package server_test

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net"
	"sync"
	"time"

	"github.com/Oyal2/tcp-server/internal/constant"
	"github.com/Oyal2/tcp-server/internal/model"
	"github.com/Oyal2/tcp-server/internal/server"
	"github.com/Oyal2/tcp-server/pkg/ratelimit"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type mockExecutor struct {
	ExecuteTaskFunc func(context context.Context, request *model.TaskRequest) *model.TaskResult
}

func (m *mockExecutor) ExecuteTask(context context.Context, request *model.TaskRequest) *model.TaskResult {
	return m.ExecuteTaskFunc(context, request)
}

var _ = Describe("TCPServer", func() {
	var (
		s       *server.TCPServer
		mockExe *mockExecutor
		port    int
	)

	const (
		readTimeout  = time.Second * 3
		writeTimeout = time.Second * 3
	)

	BeforeEach(func() {
		mockExe = &mockExecutor{}
		port = 0
		rateLimiter, err := ratelimit.NewIPRateLimiter(constant.DefaultRateLimit, constant.DefaultRateInterval)
		Expect(err).NotTo(HaveOccurred())
		params := server.TCPServerParams{
			Port:         port,
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
			Executor:     mockExe,
			WaitGroup:    &sync.WaitGroup{},
			RateLimiter:  rateLimiter,
		}
		s, err = server.NewTCPServer(params)
		Expect(err).NotTo(HaveOccurred())

		go s.Start(context.Background())
		time.Sleep(100 * time.Millisecond)
	})

	AfterEach(func() {
		s.Stop()
	})

	It("should handle a valid request", func() {
		mockExe.ExecuteTaskFunc = func(ctx context.Context, request *model.TaskRequest) *model.TaskResult {
			return &model.TaskResult{
				Command:    request.Command,
				ExecutedAt: time.Now().Unix(),
				DurationMs: 100,
				ExitCode:   0,
				Output:     "test output",
			}
		}

		conn, err := net.Dial("tcp", s.Addr().String())
		Expect(err).NotTo(HaveOccurred())
		defer conn.Close()

		request := model.TaskRequest{
			Command: []string{"test", "command"},
			Timeout: 1000,
		}
		requestJSON, err := json.Marshal(request)
		Expect(err).NotTo(HaveOccurred())

		_, err = conn.Write(append(requestJSON, '\n'))
		Expect(err).NotTo(HaveOccurred())

		var response model.TaskResult
		err = json.NewDecoder(conn).Decode(&response)
		Expect(err).NotTo(HaveOccurred())
		Expect(response.Command).To(Equal(request.Command))
		Expect(response.Output).To(Equal("test output"))
		Expect(response.ExitCode).To(Equal(0))
	})

	It("should handle a timeout", func() {
		mockExe.ExecuteTaskFunc = func(ctx context.Context, request *model.TaskRequest) *model.TaskResult {
			time.Sleep(2 * time.Second)
			return &model.TaskResult{
				Command:    request.Command,
				ExecutedAt: time.Now().Unix(),
				DurationMs: 2000,
				ExitCode:   -1,
				Error:      "timeout exceeded",
			}
		}

		conn, err := net.Dial("tcp", s.Addr().String())
		Expect(err).NotTo(HaveOccurred())
		defer conn.Close()

		request := model.TaskRequest{
			Command: []string{"slow", "command"},
			Timeout: 1000,
		}
		requestJSON, err := json.Marshal(request)
		Expect(err).NotTo(HaveOccurred())

		_, err = conn.Write(append(requestJSON, '\n'))
		Expect(err).NotTo(HaveOccurred())

		var response model.TaskResult
		err = json.NewDecoder(conn).Decode(&response)
		Expect(err).NotTo(HaveOccurred())
		Expect(response.ExitCode).To(Equal(-1))
		Expect(response.Error).To(Equal("timeout exceeded"))
	})

	It("should handle malformed JSON", func() {
		const malformedErr = "Error parsing request"
		conn, err := net.Dial("tcp", s.Addr().String())
		Expect(err).NotTo(HaveOccurred())
		defer conn.Close()

		_, err = conn.Write([]byte("invalid json\n"))
		Expect(err).NotTo(HaveOccurred())

		out, err := bufio.NewReader(conn).ReadString('\n')
		Expect(err).To(HaveOccurred())
		Expect(out).To(ContainSubstring(malformedErr))
		Expect(conn.Close()).ToNot(HaveOccurred())
	})

	It("should handle read timeout", func() {
		conn, err := net.Dial("tcp", s.Addr().String())
		Expect(err).NotTo(HaveOccurred())
		defer conn.Close()

		_, err = conn.Write([]byte("no_new_line"))
		Expect(err).NotTo(HaveOccurred())

		time.Sleep(s.ReadTimeout() + 100*time.Millisecond)

		out, err := bufio.NewReader(conn).ReadString('\n')
		Expect(err).To(HaveOccurred())
		Expect(out).To(ContainSubstring("Error parsing request"))
	})

	It("should handle write timeout", func() {
		mockExe.ExecuteTaskFunc = func(ctx context.Context, request *model.TaskRequest) *model.TaskResult {
			return &model.TaskResult{
				Command:    request.Command,
				ExecutedAt: time.Now().Unix(),
				DurationMs: 100,
				ExitCode:   0,
				Output:     "test output test outputtest outputtest outputtest outputtest outputtest outputtest outputtest outputtest outputtest outputtest outputtest outputtest outputtest outputtest outputtest outputtest outputtest outputtest outputtest outputtest outputtest outputtest outputtest outputtest outputtest outputtest outputtest outputtest outputtest outputtest outputtest outputtest outputtest outputtest outputtest outputtest outputtest output",
			}
		}

		conn, err := net.Dial("tcp", s.Addr().String())
		Expect(err).NotTo(HaveOccurred())
		defer conn.Close()

		request := model.TaskRequest{
			Command: []string{"test", "command"},
			Timeout: 1000,
		}
		requestJSON, err := json.Marshal(request)
		Expect(err).NotTo(HaveOccurred())

		_, err = conn.Write(append(requestJSON, '\n'))
		Expect(err).NotTo(HaveOccurred())

		buffer := make([]byte, 10)
		_, err = conn.Read(buffer)
		Expect(err).NotTo(HaveOccurred())

		err = conn.SetReadDeadline(time.Now().Add(s.WriteTimeout() + s.WriteTimeout()))
		Expect(err).NotTo(HaveOccurred())
		time.Sleep(s.WriteTimeout() + s.WriteTimeout())

		_, err = io.ReadAll(conn)
		Expect(err).To(HaveOccurred())

		netErr, ok := err.(net.Error)
		Expect(ok).To(BeTrue())
		Expect(netErr.Timeout()).To(BeTrue())
	})
})
