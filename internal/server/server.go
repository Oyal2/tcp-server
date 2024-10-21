package server

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/Oyal2/tcp-server/internal/model"
	"github.com/Oyal2/tcp-server/pkg/executor"
	"github.com/Oyal2/tcp-server/pkg/ratelimit"
)

type TCPServer struct {
	executor    executor.TaskExecutor
	wg          *sync.WaitGroup
	rateLimiter ratelimit.RateLimiter

	mu           sync.RWMutex
	listener     net.Listener
	readTimeout  time.Duration
	writeTimeout time.Duration
}

type TCPServerParams struct {
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	Executor     executor.TaskExecutor
	WaitGroup    *sync.WaitGroup
	RateLimiter  ratelimit.RateLimiter
}

func NewTCPServer(params TCPServerParams) (*TCPServer, error) {
	// Create a tcp listener
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", params.Port))
	if err != nil {
		return nil, fmt.Errorf("error listening: %w", err)
	}

	// If there is not waitgroup assinged, then we will assign it one
	if params.WaitGroup == nil {
		params.WaitGroup = &sync.WaitGroup{}
	}

	ts := TCPServer{
		listener:     listener,
		readTimeout:  params.ReadTimeout,
		writeTimeout: params.WriteTimeout,
		executor:     params.Executor,
		wg:           params.WaitGroup,
		rateLimiter:  params.RateLimiter,
	}

	return &ts, nil
}

func (s *TCPServer) Start(ctx context.Context) {
	// Get ready to close the server when we end this function
	defer s.listener.Close()
	log.Printf("Server listening on %s", s.listener.Addr())

	// Asynchronously run a cleanup on our rate limiter if we have one assigned.
	if s.rateLimiter != nil {
		go s.handleRateLimitCleanup(ctx)
	}

	for {
		// Look out for any client connections
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				log.Printf("listener closed: %v", err)
				return
			}
		}

		// Add to our waitgroup a new process that will be running
		s.wg.Add(1)
		// Asynchronously handle the incoming connection
		go s.handleConnection(ctx, conn)
	}
}

func (s *TCPServer) Stop() {
	s.listener.Close()
	s.wg.Wait()
}

func (s *TCPServer) Addr() net.Addr {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.listener.Addr()
}

func (s *TCPServer) ReadTimeout() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.readTimeout
}

func (s *TCPServer) WriteTimeout() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.writeTimeout
}

func (s *TCPServer) handleConnection(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	defer s.wg.Done()

	// Get the IP Address without the port if they have one. (Limits malicious actors from requesting from different ports)
	ip, _, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		log.Printf("Error parsing remote address: %v", err)
		return
	}

	// Check if the ip is rate limited or not
	if !s.rateLimiter.Allow(ip) {
		log.Printf("Rate limit exceeded for IP: %s", ip)
		return
	}

	// Start extracting the information, lets run this until our cancel context is activated.
	scanner := bufio.NewScanner(conn)
	for ctx.Err() == nil {
		// set a reading deadline
		if err := conn.SetReadDeadline(time.Now().Add(s.readTimeout)); err != nil {
			log.Printf("Error setting read deadline: %v", err)
			return
		}
		// wait for a message with a new line
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				if err != io.EOF {
					log.Printf("Error reading from connection: %v", err)
				}
				fmt.Fprint(conn, err)
			}
			return
		}
		b := scanner.Bytes()
		// Unmarshal the incoming request. We expect only the TaskRequest json
		var request model.TaskRequest
		if err := json.Unmarshal(b, &request); err != nil {
			err := fmt.Sprintf("Error parsing request: %v", err)
			log.Print(err)
			fmt.Fprint(conn, err)
			return
		}

		if request.Timeout > 0 {
			// Create a timeout with the new timeout
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, time.Duration(request.Timeout)*time.Millisecond)
			defer cancel()
		}
		// Execute the task
		result := s.executor.ExecuteTask(ctx, &request)

		// Set a writing deadline
		if err := conn.SetWriteDeadline(time.Now().Add(s.writeTimeout)); err != nil {
			log.Printf("Error setting write deadline: %v", err)
			return
		}

		// Unmarshal the result from executing the task
		response, err := json.Marshal(result)
		if err != nil {
			err := fmt.Sprintf("Error marshaling response: %v", err)
			log.Print(err)
			fmt.Fprint(conn, err)
			continue
		}

		// Write out the marshalled response.
		if _, err := conn.Write(response); err != nil {
			log.Printf("Error sending response: %v", err)
			fmt.Fprintf(conn, "Error sending response: %v", err)
			continue
		}
	}
}

func (s *TCPServer) handleRateLimitCleanup(ctx context.Context) {
	ticker := time.NewTicker(time.Minute * 5)
	defer ticker.Stop()

	// Every 5 minutes we run the rate limiter to clean up any "old" IPs
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.rateLimiter.Clean()
		}
	}
}
