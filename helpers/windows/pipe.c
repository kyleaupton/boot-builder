#include "pipe.h"
#include "protocol.h"
#include <stdio.h>
#include <string.h>

/*
 * Named pipe implementation for FlashIt privileged helper.
 * STUB IMPLEMENTATION - Basic structure without full I/O.
 */

HANDLE create_pipe_server(const char *pipe_name) {
    /*
     * Create named pipe with:
     * - PIPE_ACCESS_DUPLEX for bidirectional communication
     * - PIPE_TYPE_BYTE | PIPE_READMODE_BYTE for byte-mode I/O
     * - PIPE_WAIT for blocking operations
     * - Security attributes that allow non-elevated clients to connect
     */

    if (pipe_name == NULL || pipe_name[0] == '\0') {
        SetLastError(ERROR_INVALID_PARAMETER);
        return INVALID_HANDLE_VALUE;
    }

    /* Create a security descriptor that allows Everyone to connect */
    SECURITY_DESCRIPTOR sd;
    if (!InitializeSecurityDescriptor(&sd, SECURITY_DESCRIPTOR_REVISION)) {
        return INVALID_HANDLE_VALUE;
    }

    /* Set a NULL DACL = allow all access (needed for non-elevated client) */
    if (!SetSecurityDescriptorDacl(&sd, TRUE, NULL, FALSE)) {
        return INVALID_HANDLE_VALUE;
    }

    SECURITY_ATTRIBUTES sa;
    sa.nLength = sizeof(SECURITY_ATTRIBUTES);
    sa.lpSecurityDescriptor = &sd;
    sa.bInheritHandle = FALSE;

    HANDLE pipe = CreateNamedPipeA(
        pipe_name,
        PIPE_ACCESS_DUPLEX,
        PIPE_TYPE_BYTE | PIPE_READMODE_BYTE | PIPE_WAIT,
        1,                      /* Max instances */
        PIPE_BUFFER_SIZE,       /* Out buffer size */
        PIPE_BUFFER_SIZE,       /* In buffer size */
        PIPE_TIMEOUT_MS,        /* Default timeout */
        &sa                     /* Security allowing non-elevated access */
    );

    return pipe;
}

int wait_for_client(HANDLE pipe) {
    if (pipe == NULL || pipe == INVALID_HANDLE_VALUE) {
        return -1;
    }

    BOOL connected = ConnectNamedPipe(pipe, NULL);
    if (!connected) {
        DWORD err = GetLastError();
        /* ERROR_PIPE_CONNECTED means client already connected before we called */
        if (err != ERROR_PIPE_CONNECTED) {
            return -1;
        }
    }

    return 0;
}

int read_request_json(HANDLE pipe, char *buffer, size_t buffer_size) {
    if (pipe == NULL || pipe == INVALID_HANDLE_VALUE ||
        buffer == NULL || buffer_size == 0) {
        return -1;
    }

    /*
     * Read until newline delimiter.
     * This is a simple implementation that reads byte by byte.
     * TODO: Optimize with buffered reading.
     */

    size_t pos = 0;
    char ch;
    DWORD bytes_read;

    while (pos < buffer_size - 1) {
        BOOL success = ReadFile(pipe, &ch, 1, &bytes_read, NULL);
        if (!success) {
            DWORD err = GetLastError();
            if (err == ERROR_BROKEN_PIPE || err == ERROR_NO_DATA) {
                /* Pipe closed by client */
                return 0;
            }
            return -1;
        }

        if (bytes_read == 0) {
            /* Pipe closed */
            return 0;
        }

        if (ch == '\n') {
            buffer[pos] = '\0';
            return (int)pos;
        }

        buffer[pos++] = ch;
    }

    /* Buffer full without finding newline */
    buffer[buffer_size - 1] = '\0';
    return (int)pos;
}

int read_request(HANDLE pipe, Request *req) {
    if (req == NULL) {
        return -1;
    }

    char buffer[MAX_JSON_BUFFER];
    int len = read_request_json(pipe, buffer, sizeof(buffer));

    if (len < 0) {
        return -1;  /* Error */
    }
    if (len == 0) {
        return 1;   /* Pipe closed */
    }

    return parse_request(buffer, req);
}

int send_response(HANDLE pipe, const Response *resp) {
    if (pipe == NULL || pipe == INVALID_HANDLE_VALUE || resp == NULL) {
        return -1;
    }

    char buffer[MAX_JSON_BUFFER];
    if (serialize_response(resp, buffer, sizeof(buffer) - 1) != 0) {
        return -1;
    }

    /* Append newline delimiter */
    size_t len = strlen(buffer);
    buffer[len] = '\n';
    buffer[len + 1] = '\0';

    return write_pipe(pipe, buffer, len + 1);
}

int write_pipe(HANDLE pipe, const char *data, size_t len) {
    if (pipe == NULL || pipe == INVALID_HANDLE_VALUE ||
        data == NULL || len == 0) {
        return -1;
    }

    DWORD bytes_written;
    BOOL success = WriteFile(pipe, data, (DWORD)len, &bytes_written, NULL);

    if (!success || bytes_written != (DWORD)len) {
        return -1;
    }

    /* Flush to ensure data is sent immediately */
    FlushFileBuffers(pipe);

    return 0;
}

void close_pipe(HANDLE pipe) {
    if (pipe != NULL && pipe != INVALID_HANDLE_VALUE) {
        FlushFileBuffers(pipe);
        DisconnectNamedPipe(pipe);
        CloseHandle(pipe);
    }
}

void disconnect_client(HANDLE pipe) {
    if (pipe != NULL && pipe != INVALID_HANDLE_VALUE) {
        FlushFileBuffers(pipe);
        DisconnectNamedPipe(pipe);
    }
}
