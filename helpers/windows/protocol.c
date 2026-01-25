#include "protocol.h"
#include "cJSON.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>

/*
 * Protocol implementation for FlashIt privileged helper.
 * Uses cJSON for JSON parsing and serialization.
 */

CommandType parse_command_type(const char *cmd_str) {
    if (cmd_str == NULL) {
        return CMD_UNKNOWN;
    }
    if (strcmp(cmd_str, "write_iso") == 0) {
        return CMD_WRITE_ISO;
    }
    if (strcmp(cmd_str, "format_disk") == 0) {
        return CMD_FORMAT_DISK;
    }
    if (strcmp(cmd_str, "cancel") == 0) {
        return CMD_CANCEL;
    }
    if (strcmp(cmd_str, "shutdown") == 0) {
        return CMD_SHUTDOWN;
    }
    if (strcmp(cmd_str, "ping") == 0) {
        return CMD_PING;
    }
    return CMD_UNKNOWN;
}

static void safe_strcpy(char *dest, size_t dest_size, const char *src) {
    if (src == NULL) {
        dest[0] = '\0';
        return;
    }
    strncpy(dest, src, dest_size - 1);
    dest[dest_size - 1] = '\0';
}

int parse_request(const char *json, Request *req) {
    if (json == NULL || req == NULL) {
        return -1;
    }

    /* Initialize request with defaults */
    memset(req, 0, sizeof(Request));

    cJSON *root = cJSON_Parse(json);
    if (root == NULL) {
        return -1;
    }

    /* Parse required "id" field */
    cJSON *id = cJSON_GetObjectItemCaseSensitive(root, "id");
    if (cJSON_IsString(id) && id->valuestring != NULL) {
        safe_strcpy(req->id, sizeof(req->id), id->valuestring);
    }

    /* Parse required "command" field */
    cJSON *command = cJSON_GetObjectItemCaseSensitive(root, "command");
    if (cJSON_IsString(command) && command->valuestring != NULL) {
        req->command = parse_command_type(command->valuestring);
    }

    /* Parse optional "device" field */
    cJSON *device = cJSON_GetObjectItemCaseSensitive(root, "device");
    if (cJSON_IsString(device) && device->valuestring != NULL) {
        safe_strcpy(req->device, sizeof(req->device), device->valuestring);
    }

    /* Parse optional "iso_path" field */
    cJSON *iso_path = cJSON_GetObjectItemCaseSensitive(root, "iso_path");
    if (cJSON_IsString(iso_path) && iso_path->valuestring != NULL) {
        safe_strcpy(req->iso_path, sizeof(req->iso_path), iso_path->valuestring);
    }

    /* Parse optional "filesystem" field */
    cJSON *filesystem = cJSON_GetObjectItemCaseSensitive(root, "filesystem");
    if (cJSON_IsString(filesystem) && filesystem->valuestring != NULL) {
        safe_strcpy(req->filesystem, sizeof(req->filesystem), filesystem->valuestring);
    }

    /* Parse optional "volume_name" field */
    cJSON *volume_name = cJSON_GetObjectItemCaseSensitive(root, "volume_name");
    if (cJSON_IsString(volume_name) && volume_name->valuestring != NULL) {
        safe_strcpy(req->volume_name, sizeof(req->volume_name), volume_name->valuestring);
    }

    /* Parse optional "target_id" field (for cancel command) */
    cJSON *target_id = cJSON_GetObjectItemCaseSensitive(root, "target_id");
    if (cJSON_IsString(target_id) && target_id->valuestring != NULL) {
        safe_strcpy(req->target_id, sizeof(req->target_id), target_id->valuestring);
    }

    cJSON_Delete(root);
    return 0;
}

static const char* response_type_to_string(ResponseType type) {
    switch (type) {
        case RESP_PROGRESS: return "progress";
        case RESP_RESULT:   return "result";
        case RESP_ERROR:    return "error";
        default:            return "unknown";
    }
}

int serialize_response(const Response *resp, char *json_out, size_t json_out_size) {
    if (resp == NULL || json_out == NULL || json_out_size < 64) {
        return -1;
    }

    cJSON *root = cJSON_CreateObject();
    if (root == NULL) {
        return -1;
    }

    cJSON_AddStringToObject(root, "id", resp->id);
    cJSON_AddStringToObject(root, "type", response_type_to_string(resp->type));

    switch (resp->type) {
        case RESP_PROGRESS:
            cJSON_AddNumberToObject(root, "written", (double)resp->written);
            cJSON_AddNumberToObject(root, "total", (double)resp->total);
            break;
        case RESP_RESULT:
            cJSON_AddBoolToObject(root, "success", resp->success);
            if (!resp->success && resp->error[0] != '\0') {
                cJSON_AddStringToObject(root, "error", resp->error);
            }
            break;
        case RESP_ERROR:
            cJSON_AddBoolToObject(root, "success", 0);
            cJSON_AddStringToObject(root, "error", resp->error);
            break;
    }

    char *json_str = cJSON_PrintUnformatted(root);
    cJSON_Delete(root);

    if (json_str == NULL) {
        return -1;
    }

    size_t len = strlen(json_str);
    if (len >= json_out_size) {
        free(json_str);
        return -1;
    }

    strcpy(json_out, json_str);
    free(json_str);
    return 0;
}

void init_response(Response *resp, const char *request_id) {
    if (resp == NULL) {
        return;
    }
    memset(resp, 0, sizeof(Response));
    if (request_id != NULL) {
        safe_strcpy(resp->id, sizeof(resp->id), request_id);
    }
}

void make_progress_response(Response *resp, const char *request_id,
                            uint64_t written, uint64_t total) {
    init_response(resp, request_id);
    resp->type = RESP_PROGRESS;
    resp->written = written;
    resp->total = total;
}

void make_success_response(Response *resp, const char *request_id) {
    init_response(resp, request_id);
    resp->type = RESP_RESULT;
    resp->success = 1;
}

void make_error_response(Response *resp, const char *request_id, const char *error_msg) {
    init_response(resp, request_id);
    resp->type = RESP_ERROR;
    resp->success = 0;
    if (error_msg != NULL) {
        safe_strcpy(resp->error, sizeof(resp->error), error_msg);
    }
}
