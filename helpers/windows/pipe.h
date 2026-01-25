#ifndef FLASHIT_PIPE_H
#define FLASHIT_PIPE_H

#include <windows.h>
#include "protocol.h"

/*
 * Named pipe helpers for FlashIt privileged helper.
 * The helper creates a named pipe server that the Go client connects to.
 */

/* Pipe buffer size */
#define PIPE_BUFFER_SIZE 65536

/* Timeout for pipe operations in milliseconds */
#define PIPE_TIMEOUT_MS 30000

/*
 * Create a named pipe server with the specified name.
 * The pipe name should be in the format: \\.\pipe\flashit-helper-{uuid}
 *
 * Returns a valid HANDLE on success, INVALID_HANDLE_VALUE on error.
 */
HANDLE create_pipe_server(const char *pipe_name);

/*
 * Wait for a client to connect to the pipe.
 * This is a blocking call.
 *
 * Returns 0 on success, -1 on error.
 */
int wait_for_client(HANDLE pipe);

/*
 * Read a complete JSON request from the pipe.
 * Reads until a newline character is encountered.
 *
 * pipe: The pipe handle to read from
 * buffer: Output buffer for the JSON string (must be at least MAX_JSON_BUFFER bytes)
 * buffer_size: Size of the output buffer
 *
 * Returns the number of bytes read on success, -1 on error, 0 on pipe closed.
 */
int read_request_json(HANDLE pipe, char *buffer, size_t buffer_size);

/*
 * Read a request from the pipe and parse it into a Request structure.
 *
 * pipe: The pipe handle to read from
 * req: Output Request structure
 *
 * Returns 0 on success, -1 on error, 1 on pipe closed.
 */
int read_request(HANDLE pipe, Request *req);

/*
 * Send a response to the pipe as JSON.
 * Appends a newline character to the JSON.
 *
 * pipe: The pipe handle to write to
 * resp: The response to send
 *
 * Returns 0 on success, -1 on error.
 */
int send_response(HANDLE pipe, const Response *resp);

/*
 * Send raw JSON data to the pipe.
 * The data should already include the newline delimiter.
 *
 * pipe: The pipe handle to write to
 * json: The JSON string to send
 * len: Length of the JSON string
 *
 * Returns 0 on success, -1 on error.
 */
int write_pipe(HANDLE pipe, const char *data, size_t len);

/*
 * Close the pipe handle.
 */
void close_pipe(HANDLE pipe);

/*
 * Disconnect a client from the pipe server.
 * Call this after a client disconnects to prepare for a new client.
 */
void disconnect_client(HANDLE pipe);

#endif /* FLASHIT_PIPE_H */
