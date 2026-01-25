//go:build windows

package windows

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"

	"github.com/google/uuid"
	"golang.org/x/sys/windows"
)

// Windows API constants
const (
	seeMaskNoCloseProcess = 0x00000040
	swHide                = 0
)

// SHELLEXECUTEINFOW structure for ShellExecuteExW
type shellExecuteInfo struct {
	cbSize         uint32
	fMask          uint32
	hwnd           uintptr
	lpVerb         *uint16
	lpFile         *uint16
	lpParameters   *uint16
	lpDirectory    *uint16
	nShow          int32
	hInstApp       uintptr
	lpIDList       uintptr
	lpClass        *uint16
	hkeyClass      uintptr
	dwHotKey       uint32
	hIconOrMonitor uintptr
	hProcess       windows.Handle
}

var (
	shell32             = syscall.NewLazyDLL("shell32.dll")
	kernel32            = syscall.NewLazyDLL("kernel32.dll")
	procShellExecuteExW = shell32.NewProc("ShellExecuteExW")
	procWaitNamedPipeW  = kernel32.NewProc("WaitNamedPipeW")
)

// shellExecuteEx wraps the ShellExecuteExW Windows API
func shellExecuteEx(sei *shellExecuteInfo) error {
	ret, _, err := procShellExecuteExW.Call(uintptr(unsafe.Pointer(sei)))
	if ret == 0 {
		return err
	}
	return nil
}

// waitNamedPipe wraps the WaitNamedPipeW Windows API
func waitNamedPipe(name *uint16, timeout uint32) error {
	ret, _, err := procWaitNamedPipeW.Call(
		uintptr(unsafe.Pointer(name)),
		uintptr(timeout),
	)
	if ret == 0 {
		return err
	}
	return nil
}

// pipeClient implements the Client interface using named pipes.
type pipeClient struct {
	pipeName   string
	helperPath string
	helperProc windows.Handle
	pipe       windows.Handle

	mu        sync.Mutex
	ready     bool
	requestID atomic.Uint64

	// reader is a buffered reader for the pipe
	reader *bufio.Reader

	// cancelMu protects the current operation's cancel channel
	cancelMu     sync.Mutex
	cancelChan   chan struct{}
	currentReqID string
}

// NewClient creates a new Windows privileged helper client.
// The helperPath is resolved automatically if empty.
func NewClient() Client {
	return &pipeClient{
		pipeName:   PipeNamePrefix + uuid.New().String(),
		pipe:       windows.InvalidHandle,
		helperProc: windows.InvalidHandle,
	}
}

// EnsureReady spawns the helper process with UAC elevation if not already running,
// and verifies communication is working.
func (c *pipeClient) EnsureReady(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.ready {
		return nil
	}

	// 1. Locate the helper executable
	helperPath, err := c.findHelperPath()
	if err != nil {
		return fmt.Errorf("failed to find helper: %w", err)
	}
	c.helperPath = helperPath

	// 2. Spawn helper with UAC elevation using ShellExecuteEx
	if err := c.spawnElevated(ctx); err != nil {
		return fmt.Errorf("failed to spawn elevated helper: %w", err)
	}

	// 3. Wait for pipe to become available
	if err := c.waitForPipe(ctx); err != nil {
		return fmt.Errorf("failed waiting for pipe: %w", err)
	}

	// 4. Connect to the named pipe
	if err := c.connectToPipe(ctx); err != nil {
		return fmt.Errorf("failed to connect to pipe: %w", err)
	}

	// 5. Create buffered reader for the pipe
	c.reader = bufio.NewReader(&pipeReader{handle: c.pipe})

	// 6. Send a ping request to verify communication
	pingReq := Request{
		ID:      c.nextRequestID(),
		Command: CmdPing,
	}

	resp, err := c.sendRequestUnlocked(ctx, pingReq)
	if err != nil {
		c.closePipeUnlocked()
		return fmt.Errorf("ping failed: %w", err)
	}

	if resp.Type == RespTypeError || !resp.Success {
		c.closePipeUnlocked()
		return fmt.Errorf("ping failed: %s", resp.Error)
	}

	c.ready = true
	return nil
}

// findHelperPath locates the helper executable.
// It searches in the following locations:
// 1. Same directory as the main executable
// 2. bin/helpers/ relative to executable
// 3. Current working directory
func (c *pipeClient) findHelperPath() (string, error) {
	const helperName = "flashit-helper.exe"

	// Get the main executable path
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	exeDir := filepath.Dir(exe)

	// Candidate paths to check
	candidates := []string{
		filepath.Join(exeDir, helperName),
		filepath.Join(exeDir, "helpers", helperName),
		filepath.Join(exeDir, "bin", "helpers", helperName),
		filepath.Join(".", "bin", "helpers", helperName),
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return filepath.Abs(path)
		}
	}

	return "", fmt.Errorf("helper executable not found in any of: %v", candidates)
}

// spawnElevated spawns the helper with UAC elevation using ShellExecuteEx.
func (c *pipeClient) spawnElevated(ctx context.Context) error {
	// Convert strings to UTF16 for Windows API
	verb, _ := windows.UTF16PtrFromString("runas")
	file, _ := windows.UTF16PtrFromString(c.helperPath)
	params, _ := windows.UTF16PtrFromString(c.pipeName)
	dir, _ := windows.UTF16PtrFromString(filepath.Dir(c.helperPath))

	sei := &shellExecuteInfo{
		cbSize:       uint32(unsafe.Sizeof(shellExecuteInfo{})),
		fMask:        seeMaskNoCloseProcess,
		lpVerb:       verb,
		lpFile:       file,
		lpParameters: params,
		lpDirectory:  dir,
		nShow:        swHide,
	}

	err := shellExecuteEx(sei)
	if err != nil {
		return fmt.Errorf("ShellExecuteEx failed: %w", err)
	}

	if sei.hProcess == 0 {
		return errors.New("ShellExecuteEx succeeded but no process handle returned")
	}

	c.helperProc = sei.hProcess
	return nil
}

// waitForPipe waits for the helper to create the named pipe.
func (c *pipeClient) waitForPipe(ctx context.Context) error {
	pipePath, _ := windows.UTF16PtrFromString(c.pipeName)

	// Poll with timeout for pipe to exist
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return errors.New("timeout waiting for helper to create pipe")
		case <-ticker.C:
			// Try WaitNamedPipe with a short timeout
			err := waitNamedPipe(pipePath, 100)
			if err == nil {
				// Pipe exists and is ready
				return nil
			}
			// ERROR_FILE_NOT_FOUND means pipe doesn't exist yet
			// Continue polling
		}
	}
}

// connectToPipe opens a connection to the named pipe.
func (c *pipeClient) connectToPipe(ctx context.Context) error {
	pipePath, _ := windows.UTF16PtrFromString(c.pipeName)

	handle, err := windows.CreateFile(
		pipePath,
		windows.GENERIC_READ|windows.GENERIC_WRITE,
		0,   // No sharing
		nil, // Default security
		windows.OPEN_EXISTING,
		windows.FILE_ATTRIBUTE_NORMAL,
		0, // No template
	)
	if err != nil {
		return fmt.Errorf("CreateFile failed: %w", err)
	}

	c.pipe = handle
	return nil
}

// pipeReader wraps a windows.Handle to implement io.Reader.
type pipeReader struct {
	handle windows.Handle
}

func (r *pipeReader) Read(p []byte) (int, error) {
	var bytesRead uint32
	err := windows.ReadFile(r.handle, p, &bytesRead, nil)
	if err != nil {
		return 0, err
	}
	return int(bytesRead), nil
}

// WriteISO writes an ISO image to a raw disk device.
func (c *pipeClient) WriteISO(ctx context.Context, isoPath string, device string, progress ProgressFunc) error {
	c.mu.Lock()
	if !c.ready {
		c.mu.Unlock()
		return ErrHelperNotRunning
	}
	c.mu.Unlock()

	reqID := c.nextRequestID()

	// Set up cancellation tracking
	c.cancelMu.Lock()
	c.currentReqID = reqID
	c.cancelChan = make(chan struct{})
	cancelCh := c.cancelChan
	c.cancelMu.Unlock()

	defer func() {
		c.cancelMu.Lock()
		c.currentReqID = ""
		c.cancelChan = nil
		c.cancelMu.Unlock()
	}()

	// Set up context cancellation
	done := make(chan struct{})
	defer close(done)

	go func() {
		select {
		case <-ctx.Done():
			c.CancelCurrentOperation()
		case <-cancelCh:
			// Operation was cancelled via CancelCurrentOperation
		case <-done:
			// Operation completed normally
		}
	}()

	req := Request{
		ID:      reqID,
		Command: CmdWriteISO,
		Device:  device,
		ISOPath: isoPath,
	}

	resp, err := c.sendRequestWithProgress(ctx, req, progress)
	if err != nil {
		return err
	}

	if resp.Type == RespTypeError || !resp.Success {
		if resp.Error != "" {
			// Check if cancelled
			if resp.Error == "cancelled: operation was cancelled" {
				return ErrCancelled
			}
			return errors.New(resp.Error)
		}
		return errors.New("operation failed")
	}

	return nil
}

// FormatDisk formats a disk with the specified filesystem and volume name.
func (c *pipeClient) FormatDisk(ctx context.Context, device string, filesystem string, volumeName string) error {
	c.mu.Lock()
	if !c.ready {
		c.mu.Unlock()
		return ErrHelperNotRunning
	}
	c.mu.Unlock()

	reqID := c.nextRequestID()

	req := Request{
		ID:         reqID,
		Command:    CmdFormatDisk,
		Device:     device,
		Filesystem: filesystem,
		VolumeName: volumeName,
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return err
	}

	if resp.Type == RespTypeError || !resp.Success {
		if resp.Error != "" {
			return errors.New(resp.Error)
		}
		return errors.New("format failed")
	}

	return nil
}

// CancelCurrentOperation cancels the currently running operation (if any).
func (c *pipeClient) CancelCurrentOperation() {
	c.cancelMu.Lock()
	reqID := c.currentReqID
	ch := c.cancelChan
	c.cancelMu.Unlock()

	if reqID == "" || ch == nil {
		return
	}

	// Signal the cancel channel
	select {
	case <-ch:
		// Already cancelled
	default:
		close(ch)
	}

	// Send cancel command to helper (fire-and-forget)
	req := Request{
		ID:       c.nextRequestID(),
		Command:  CmdCancel,
		TargetID: reqID,
	}

	// Best-effort send - ignore errors
	_ = c.writeRequest(req)
}

// Shutdown gracefully terminates the helper process.
func (c *pipeClient) Shutdown(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.ready {
		return nil
	}

	req := Request{
		ID:      c.nextRequestID(),
		Command: CmdShutdown,
	}

	// Send shutdown command
	resp, err := c.sendRequestUnlocked(ctx, req)
	if err != nil {
		// Still try to clean up
		c.closePipeUnlocked()
		c.ready = false
		return fmt.Errorf("shutdown request failed: %w", err)
	}

	if resp.Type == RespTypeError {
		c.closePipeUnlocked()
		c.ready = false
		return fmt.Errorf("shutdown failed: %s", resp.Error)
	}

	// Wait for helper process to exit (with timeout)
	if c.helperProc != 0 && c.helperProc != windows.InvalidHandle {
		event, _ := windows.WaitForSingleObject(c.helperProc, 5000) // 5 second timeout
		if event == uint32(windows.WAIT_TIMEOUT) {
			// Force terminate if it didn't exit gracefully
			windows.TerminateProcess(c.helperProc, 1)
		}
		windows.CloseHandle(c.helperProc)
		c.helperProc = windows.InvalidHandle
	}

	c.closePipeUnlocked()
	c.ready = false
	return nil
}

// nextRequestID generates a unique request ID.
func (c *pipeClient) nextRequestID() string {
	return fmt.Sprintf("req-%d", c.requestID.Add(1))
}

// sendRequest sends a request to the helper and waits for the result response.
// This acquires the mutex for thread safety.
func (c *pipeClient) sendRequest(ctx context.Context, req Request) (*Response, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.sendRequestUnlocked(ctx, req)
}

// sendRequestUnlocked sends a request without acquiring the mutex.
// Caller must hold the mutex.
func (c *pipeClient) sendRequestUnlocked(ctx context.Context, req Request) (*Response, error) {
	if c.pipe == windows.InvalidHandle {
		return nil, ErrHelperNotRunning
	}

	// Write request
	if err := c.writeRequest(req); err != nil {
		return nil, fmt.Errorf("failed to write request: %w", err)
	}

	// Read response
	resp, err := c.readResponse()
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return resp, nil
}

// sendRequestWithProgress sends a request and processes progress updates.
func (c *pipeClient) sendRequestWithProgress(ctx context.Context, req Request, progress ProgressFunc) (*Response, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.pipe == windows.InvalidHandle {
		return nil, ErrHelperNotRunning
	}

	// Write request
	if err := c.writeRequest(req); err != nil {
		return nil, fmt.Errorf("failed to write request: %w", err)
	}

	// Enter read loop for progress updates
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		resp, err := c.readResponse()
		if err != nil {
			return nil, fmt.Errorf("failed to read response: %w", err)
		}

		switch resp.Type {
		case RespTypeProgress:
			if progress != nil {
				progress(resp.Written, resp.Total)
			}
			// Continue reading
		case RespTypeResult, RespTypeError:
			return resp, nil
		default:
			// Unknown response type, treat as error
			return nil, fmt.Errorf("unknown response type: %s", resp.Type)
		}
	}
}

// writeRequest encodes and writes a request to the pipe.
func (c *pipeClient) writeRequest(req Request) error {
	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to encode request: %w", err)
	}
	data = append(data, '\n')

	var bytesWritten uint32
	err = windows.WriteFile(c.pipe, data, &bytesWritten, nil)
	if err != nil {
		return fmt.Errorf("WriteFile failed: %w", err)
	}

	if bytesWritten != uint32(len(data)) {
		return fmt.Errorf("incomplete write: wrote %d of %d bytes", bytesWritten, len(data))
	}

	// Flush to ensure data is sent immediately
	windows.FlushFileBuffers(c.pipe)

	return nil
}

// readResponse reads and decodes a response from the pipe.
func (c *pipeClient) readResponse() (*Response, error) {
	line, err := c.reader.ReadBytes('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read response line: %w", err)
	}

	var resp Response
	if err := json.Unmarshal(line, &resp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &resp, nil
}

// closePipeUnlocked closes the pipe handle.
// Caller must hold the mutex.
func (c *pipeClient) closePipeUnlocked() {
	if c.pipe != windows.InvalidHandle {
		windows.CloseHandle(c.pipe)
		c.pipe = windows.InvalidHandle
	}
	c.reader = nil
}
