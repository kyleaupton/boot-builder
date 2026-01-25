#include <windows.h>
#include <stdio.h>
#include <string.h>

#include "protocol.h"
#include "pipe.h"
#include "diskops.h"

/*
 * FlashIt Privileged Helper for Windows
 *
 * This helper runs with administrator privileges and performs
 * disk operations that require elevation (raw disk writes, formatting).
 *
 * Communication is via a named pipe using JSON messages.
 * The pipe name is passed as the first command-line argument.
 *
 * Flow:
 * 1. Parse pipe name from command line
 * 2. Create named pipe server
 * 3. Wait for client connection
 * 4. Enter message loop:
 *    a. Read request from pipe
 *    b. Dispatch to appropriate handler
 *    c. Send response(s) back
 * 5. On shutdown command or client disconnect, exit
 */

/* Global state */
static volatile int g_running = 1;
static HANDLE g_pipe = INVALID_HANDLE_VALUE;

/* Forward declarations */
static void handle_request(HANDLE pipe, const Request *req);
static void handle_write_iso(HANDLE pipe, const Request *req);
static void handle_format_disk(HANDLE pipe, const Request *req);
static void handle_cancel(HANDLE pipe, const Request *req);
static void handle_shutdown(HANDLE pipe, const Request *req);
static void handle_ping(HANDLE pipe, const Request *req);
static void send_error(HANDLE pipe, const char *request_id, const char *error);
static void send_success(HANDLE pipe, const char *request_id);

int main(int argc, char *argv[]) {
    /*
     * Expect pipe name as first argument.
     * Format: \\.\pipe\flashit-helper-{uuid}
     */

    if (argc < 2) {
        fprintf(stderr, "Usage: %s <pipe-name>\n", argv[0]);
        fprintf(stderr, "Example: %s \\\\.\\pipe\\flashit-helper-abc123\n", argv[0]);
        return 1;
    }

    const char *pipe_name = argv[1];

    /* Validate pipe name format */
    if (strncmp(pipe_name, "\\\\.\\pipe\\", 9) != 0) {
        fprintf(stderr, "Error: Invalid pipe name format. Must start with \\\\.\\pipe\\\n");
        return 1;
    }

    /* Create the named pipe server */
    g_pipe = create_pipe_server(pipe_name);
    if (g_pipe == INVALID_HANDLE_VALUE) {
        DWORD err = GetLastError();
        fprintf(stderr, "Error: Failed to create pipe server (error %lu)\n", err);
        return 1;
    }

    fprintf(stderr, "FlashIt Helper: Waiting for client on %s\n", pipe_name);

    /* Wait for client connection */
    if (wait_for_client(g_pipe) != 0) {
        DWORD err = GetLastError();
        fprintf(stderr, "Error: Failed to connect client (error %lu)\n", err);
        close_pipe(g_pipe);
        return 1;
    }

    fprintf(stderr, "FlashIt Helper: Client connected\n");

    /* Main message loop */
    while (g_running) {
        Request req;
        int result = read_request(g_pipe, &req);

        if (result < 0) {
            /* Read error */
            fprintf(stderr, "Error: Failed to read request\n");
            break;
        }

        if (result > 0) {
            /* Pipe closed by client */
            fprintf(stderr, "FlashIt Helper: Client disconnected\n");
            break;
        }

        /* Dispatch request to handler */
        handle_request(g_pipe, &req);
    }

    /* Cleanup */
    close_pipe(g_pipe);
    fprintf(stderr, "FlashIt Helper: Shutting down\n");

    return 0;
}

static void handle_request(HANDLE pipe, const Request *req) {
    if (req == NULL) {
        return;
    }

    switch (req->command) {
        case CMD_WRITE_ISO:
            handle_write_iso(pipe, req);
            break;
        case CMD_FORMAT_DISK:
            handle_format_disk(pipe, req);
            break;
        case CMD_CANCEL:
            handle_cancel(pipe, req);
            break;
        case CMD_SHUTDOWN:
            handle_shutdown(pipe, req);
            break;
        case CMD_PING:
            handle_ping(pipe, req);
            break;
        default:
            send_error(pipe, req->id, "unknown command");
            break;
    }
}

static void handle_write_iso(HANDLE pipe, const Request *req) {
    fprintf(stderr, "FlashIt Helper: WriteISO device=%s iso=%s\n",
            req->device, req->iso_path);

    OperationContext ctx = {
        .request_id = req->id,
        .pipe = pipe,
        .cancel_flag = NULL  /* TODO: Wire up cancellation */
    };

    int result = write_iso(req->device, req->iso_path, &ctx);

    if (result == 0) {
        send_success(pipe, req->id);
    } else {
        if (is_cancelled()) {
            send_error(pipe, req->id, "cancelled: operation was cancelled");
        } else {
            send_error(pipe, req->id, get_last_error_message());
        }
    }
}

static void handle_format_disk(HANDLE pipe, const Request *req) {
    fprintf(stderr, "FlashIt Helper: FormatDisk device=%s fs=%s label=%s\n",
            req->device, req->filesystem, req->volume_name);

    int result = format_disk(req->device, req->filesystem, req->volume_name);

    if (result == 0) {
        send_success(pipe, req->id);
    } else {
        send_error(pipe, req->id, get_last_error_message());
    }
}

static void handle_cancel(HANDLE pipe, const Request *req) {
    fprintf(stderr, "FlashIt Helper: Cancel target=%s\n", req->target_id);

    cancel_current_operation();
    send_success(pipe, req->id);
}

static void handle_shutdown(HANDLE pipe, const Request *req) {
    fprintf(stderr, "FlashIt Helper: Shutdown requested\n");

    send_success(pipe, req->id);
    g_running = 0;
}

static void handle_ping(HANDLE pipe, const Request *req) {
    /* Simple ping/pong for connection verification */
    send_success(pipe, req->id);
}

static void send_error(HANDLE pipe, const char *request_id, const char *error) {
    Response resp;
    make_error_response(&resp, request_id, error);
    send_response(pipe, &resp);
}

static void send_success(HANDLE pipe, const char *request_id) {
    Response resp;
    make_success_response(&resp, request_id);
    send_response(pipe, &resp);
}
