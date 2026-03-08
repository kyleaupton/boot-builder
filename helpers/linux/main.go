package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	socketPathPrefix = "/tmp/flashit-helper-"
	idleTimeout      = 60 * time.Second
)

// session wraps a single client connection and provides thread-safe response writing.
type session struct {
	conn    net.Conn
	scanner *bufio.Scanner
	mu      sync.Mutex
}

func newSession(conn net.Conn) *session {
	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 64*1024), 64*1024)
	return &session{
		conn:    conn,
		scanner: scanner,
	}
}

func (s *session) readRequest() (*Request, error) {
	if !s.scanner.Scan() {
		if err := s.scanner.Err(); err != nil {
			return nil, fmt.Errorf("read error: %w", err)
		}
		return nil, fmt.Errorf("client disconnected")
	}

	var req Request
	if err := json.Unmarshal(s.scanner.Bytes(), &req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}
	return &req, nil
}

func (s *session) sendResponse(resp Response) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.Marshal(resp)
	if err != nil {
		log.Printf("failed to marshal response: %v", err)
		return
	}
	data = append(data, '\n')
	s.conn.Write(data)
}

func (s *session) sendResult(id string, success bool) {
	s.sendResponse(Response{
		ID:      id,
		Type:    RespTypeResult,
		Success: success,
	})
}

func (s *session) sendError(id string, msg string) {
	s.sendResponse(Response{
		ID:    id,
		Type:  RespTypeError,
		Error: msg,
	})
}

func (s *session) sendProgress(id string, written, total uint64) {
	s.sendResponse(Response{
		ID:      id,
		Type:    RespTypeProgress,
		Written: written,
		Total:   total,
	})
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <socket-path>\n", os.Args[0])
		os.Exit(1)
	}

	socketPath := os.Args[1]

	// Validate socket path format
	if !strings.HasPrefix(socketPath, socketPathPrefix) || !strings.HasSuffix(socketPath, ".sock") {
		fmt.Fprintf(os.Stderr, "Error: invalid socket path format\n")
		os.Exit(1)
	}

	// Clean up stale socket file if it exists
	os.Remove(socketPath)

	// Set up signal handling for clean shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	// Create the Unix domain socket
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to create socket: %v\n", err)
		os.Exit(1)
	}

	// Restrict socket permissions to owner only
	if err := os.Chmod(socketPath, 0600); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to set socket permissions: %v\n", err)
		listener.Close()
		os.Remove(socketPath)
		os.Exit(1)
	}

	defer func() {
		listener.Close()
		os.Remove(socketPath)
	}()

	log.Printf("FlashIt Helper: listening on %s", socketPath)

	// Handle signals in background
	go func() {
		sig := <-sigCh
		log.Printf("FlashIt Helper: received signal %v, shutting down", sig)
		listener.Close()
		os.Remove(socketPath)
		os.Exit(0)
	}()

	// Accept a single connection
	conn, err := listener.Accept()
	if err != nil {
		log.Printf("FlashIt Helper: accept failed: %v", err)
		return
	}

	// Close listener — only one client allowed
	listener.Close()

	log.Printf("FlashIt Helper: client connected")

	sess := newSession(conn)
	defer conn.Close()

	// Idle timeout timer
	idleTimer := time.NewTimer(idleTimeout)
	defer idleTimer.Stop()

	// Run message loop with idle timeout
	reqCh := make(chan *Request)
	errCh := make(chan error)

	go func() {
		for {
			req, err := sess.readRequest()
			if err != nil {
				errCh <- err
				return
			}
			reqCh <- req
		}
	}()

	for {
		select {
		case req := <-reqCh:
			idleTimer.Reset(idleTimeout)
			handleRequest(sess, req)

			// Check if we should exit after handling
			if req.Command == CmdShutdown {
				log.Printf("FlashIt Helper: shutdown requested, exiting")
				return
			}

		case err := <-errCh:
			log.Printf("FlashIt Helper: %v", err)
			return

		case <-idleTimer.C:
			log.Printf("FlashIt Helper: idle timeout reached, exiting")
			return
		}
	}
}

func handleRequest(sess *session, req *Request) {
	switch req.Command {
	case CmdPing:
		sess.sendResult(req.ID, true)

	case CmdWriteISO:
		if req.Device == "" || req.ISOPath == "" {
			sess.sendError(req.ID, "device and iso_path are required")
			return
		}
		writeISO(sess, req)

	case CmdFormatDisk:
		if req.Device == "" || req.Filesystem == "" {
			sess.sendError(req.ID, "device and filesystem are required")
			return
		}
		formatDisk(sess, req)

	case CmdCancel:
		cancelCurrentOperation()
		sess.sendResult(req.ID, true)

	case CmdShutdown:
		sess.sendResult(req.ID, true)

	default:
		sess.sendError(req.ID, fmt.Sprintf("unknown command: %s", req.Command))
	}
}
