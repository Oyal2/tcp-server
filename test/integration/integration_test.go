package integration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/Oyal2/tcp-server/internal/constant"
	"github.com/Oyal2/tcp-server/internal/model"
	"github.com/Oyal2/tcp-server/internal/server"
	"github.com/Oyal2/tcp-server/pkg/executor"
	"github.com/Oyal2/tcp-server/pkg/ratelimit"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Integration", func() {
	var (
		s    *server.TCPServer
		port int
	)

	const (
		readTimeout  = time.Second * 3
		writeTimeout = time.Second * 3
	)

	BeforeEach(func() {
		exe := &executor.CommandExecutor{}
		port = 0
		rateLimiter, err := ratelimit.NewIPRateLimiter(100, constant.DefaultRateInterval)
		Expect(err).NotTo(HaveOccurred())

		params := server.TCPServerParams{
			Port:         port,
			Executor:     exe,
			WaitGroup:    &sync.WaitGroup{},
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
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

	It("should handle a real command execution", func() {
		conn, err := net.Dial("tcp", s.Addr().String())
		Expect(err).NotTo(HaveOccurred())
		defer conn.Close()

		request := model.TaskRequest{
			Command: []string{printerPath, "-message=simple_test"},
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
		Expect(response.Output).To(Equal("simple_test\n"))
		Expect(response.ExitCode).To(Equal(0))
	})

	It("should handle multiple concurrent requests", func() {
		numRequests := 100
		var wg sync.WaitGroup
		wg.Add(numRequests)

		for i := 0; i < numRequests; i++ {
			go func() {
				defer wg.Done()
				conn, err := net.Dial("tcp", s.Addr().String())
				Expect(err).NotTo(HaveOccurred())
				defer conn.Close()

				request := model.TaskRequest{
					Command: []string{printerPath, "-message=concurrency_test"},
					Timeout: 0,
				}
				requestJSON, err := json.Marshal(request)
				Expect(err).NotTo(HaveOccurred())

				_, err = conn.Write(append(requestJSON, '\n'))
				Expect(err).NotTo(HaveOccurred())

				var response model.TaskResult
				err = json.NewDecoder(conn).Decode(&response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response.Command).To(Equal(request.Command))
				Expect(response.Output).To(Equal("concurrency_test\n"))
				Expect(response.ExitCode).To(Equal(0))
			}()
		}

		wg.Wait()
	})

	Context("when executing the printer command", func() {
		It("should handle a simple printer command", func() {
			conn, err := net.Dial("tcp", s.Addr().String())
			Expect(err).NotTo(HaveOccurred())
			defer conn.Close()

			request := model.TaskRequest{
				Command: []string{printerPath},
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
			Expect(response.Output).To(Equal("test\n"))
			Expect(response.ExitCode).To(Equal(0))
		})

		It("should handle printer command with custom message and repeat", func() {
			conn, err := net.Dial("tcp", s.Addr().String())
			Expect(err).NotTo(HaveOccurred())
			defer conn.Close()

			request := model.TaskRequest{
				Command: []string{printerPath, "-message=repeated", "-repeat=3"},
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
			Expect(response.Output).To(Equal("repeated\nrepeated\nrepeated\n"))
			Expect(response.ExitCode).To(Equal(0))
		})

		It("should handle printer command timeout", func() {
			conn, err := net.Dial("tcp", s.Addr().String())
			Expect(err).NotTo(HaveOccurred())
			defer conn.Close()

			request := model.TaskRequest{
				Command: []string{printerPath, "-sleep=100"},
				Timeout: 10,
			}
			requestJSON, err := json.Marshal(request)
			Expect(err).NotTo(HaveOccurred())

			_, err = conn.Write(append(requestJSON, '\n'))
			Expect(err).NotTo(HaveOccurred())

			var response model.TaskResult
			err = json.NewDecoder(conn).Decode(&response)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Command).To(Equal(request.Command))
			Expect(response.ExitCode).To(Equal(-1))
			Expect(response.Error).To(ContainSubstring(constant.TaskResultTimeoutError))
		})

		It("should handle printer with no timeout", func() {
			conn, err := net.Dial("tcp", s.Addr().String())
			Expect(err).NotTo(HaveOccurred())
			defer conn.Close()

			request := model.TaskRequest{
				Command: []string{printerPath, "-message=no_timeout_test", "-repeat=100000", "-sleep=100"},
				Timeout: 0,
			}
			requestJSON, err := json.Marshal(request)
			Expect(err).NotTo(HaveOccurred())

			_, err = conn.Write(append(requestJSON, '\n'))
			Expect(err).NotTo(HaveOccurred())

			var response model.TaskResult
			err = json.NewDecoder(conn).Decode(&response)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Command).To(Equal(request.Command))
			Expect(response.ExitCode).To(Equal(0))
			Expect(response.Error).To(BeEmpty())
		})

		It("should handle printer with missing timeout", func() {
			conn, err := net.Dial("tcp", s.Addr().String())
			Expect(err).NotTo(HaveOccurred())
			defer conn.Close()

			command := []string{printerPath, "-message=no_timeout_test", "-repeat=100000", "-sleep=100"}
			commandJSON, err := json.Marshal(command)
			Expect(err).NotTo(HaveOccurred())

			taskReqJSON := fmt.Sprintf(`{"command":%s}`, string(commandJSON))
			_, err = conn.Write(append([]byte(taskReqJSON), '\n'))
			Expect(err).NotTo(HaveOccurred())

			var response model.TaskResult
			err = json.NewDecoder(conn).Decode(&response)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.ExitCode).To(Equal(0))
			Expect(response.Error).To(BeEmpty())
		})

		It("should handle non-existent command", func() {
			conn, err := net.Dial("tcp", s.Addr().String())
			Expect(err).NotTo(HaveOccurred())
			defer conn.Close()

			request := model.TaskRequest{
				Command: []string{"fake_command"},
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
			Expect(response.ExitCode).To(Equal(-1))
			Expect(response.Error).To(ContainSubstring("executable file not found"))
		})
	})
})
