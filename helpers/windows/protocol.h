#ifndef FLASHIT_PROTOCOL_H
#define FLASHIT_PROTOCOL_H

#include <stdint.h>

/*
 * Protocol definitions for FlashIt privileged helper.
 * Communication occurs over a named pipe using newline-delimited JSON messages.
 */

/* Maximum buffer sizes */
#define MAX_REQUEST_ID      64
#define MAX_DEVICE_PATH     256
#define MAX_ISO_PATH        1024
#define MAX_FILESYSTEM      32
#define MAX_VOLUME_NAME     128
#define MAX_ERROR_MSG       512
#define MAX_JSON_BUFFER     4096

/* Command types */
typedef enum {
    CMD_UNKNOWN = 0,
    CMD_WRITE_ISO,
    CMD_FORMAT_DISK,
    CMD_CANCEL,
    CMD_SHUTDOWN,
    CMD_PING
} CommandType;

/* Request structure - parsed from incoming JSON */
typedef struct {
    char id[MAX_REQUEST_ID];
    CommandType command;
    char device[MAX_DEVICE_PATH];
    char iso_path[MAX_ISO_PATH];
    char filesystem[MAX_FILESYSTEM];
    char volume_name[MAX_VOLUME_NAME];
    char target_id[MAX_REQUEST_ID];  /* For cancel command */
} Request;

/* Response types */
typedef enum {
    RESP_PROGRESS = 0,
    RESP_RESULT,
    RESP_ERROR
} ResponseType;

/* Response structure - serialized to outgoing JSON */
typedef struct {
    char id[MAX_REQUEST_ID];
    ResponseType type;
    int success;            /* For RESP_RESULT */
    uint64_t written;       /* For RESP_PROGRESS */
    uint64_t total;         /* For RESP_PROGRESS */
    char error[MAX_ERROR_MSG];
} Response;

/*
 * Parse a JSON request string into a Request structure.
 * Returns 0 on success, -1 on parse error.
 */
int parse_request(const char *json, Request *req);

/*
 * Serialize a Response structure to JSON string.
 * The caller must provide a buffer of at least MAX_JSON_BUFFER bytes.
 * Returns 0 on success, -1 on serialization error.
 */
int serialize_response(const Response *resp, char *json_out, size_t json_out_size);

/*
 * Initialize a Response with default values.
 */
void init_response(Response *resp, const char *request_id);

/*
 * Helper to create a progress response.
 */
void make_progress_response(Response *resp, const char *request_id,
                            uint64_t written, uint64_t total);

/*
 * Helper to create a success result response.
 */
void make_success_response(Response *resp, const char *request_id);

/*
 * Helper to create an error response.
 */
void make_error_response(Response *resp, const char *request_id, const char *error_msg);

/*
 * Convert command string to CommandType enum.
 */
CommandType parse_command_type(const char *cmd_str);

#endif /* FLASHIT_PROTOCOL_H */
