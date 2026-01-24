//
//  xpc_client.m
//  Boot Builder XPC Client
//
//  Client-side XPC calls to the privileged helper.
//

#import <Foundation/Foundation.h>
#import <CoreFoundation/CoreFoundation.h>
#import <ServiceManagement/ServiceManagement.h>
#import <Security/Security.h>
#import <dispatch/dispatch.h>

static const char *kHelperLabel = "dev.kyleupton.boot-builder.helper";
static const char *kHelperBundleID = "dev.kyleupton.boot-builder.helper";

// Active XPC connection for the current operation.
// Used to send cancel requests over the same connection as the operation.
static NSXPCConnection *_activeConnection = nil;

// CGO callback function exported from Go
extern void GoProgressCallback(uint64_t written, uint64_t total);

/// Protocol for receiving progress updates during write operations.
/// The client implements this and exports it via the XPC connection.
@protocol BBProgressReporter <NSObject>
- (void)updateProgress:(uint64_t)bytesWritten totalBytes:(uint64_t)totalBytes;
@end

@protocol BBPrivilegedHelper
- (void)writeLinuxISO:(NSString *)device
              isoPath:(NSString *)isoPath
     progressReporter:(id<BBProgressReporter>)reporter
                reply:(void (^)(BOOL success, NSString *err))reply;
- (void)formatDisk:(NSString *)device
        filesystem:(NSString *)filesystem
        volumeName:(NSString *)volumeName
             reply:(void (^)(BOOL success, NSString *err))reply;
- (void)cancelCurrentOperation;
@end

/// Progress reporter implementation that bridges to Go via CGO callback.
@interface WriteProgressReporter : NSObject <BBProgressReporter>
@end

@implementation WriteProgressReporter
- (void)updateProgress:(uint64_t)bytesWritten totalBytes:(uint64_t)totalBytes {
    GoProgressCallback(bytesWritten, totalBytes);
}
@end

static NSXPCConnection *create_helper_connection(void) {
    NSString *label = [NSString stringWithUTF8String:kHelperLabel];
    NSXPCConnection *conn = [[NSXPCConnection alloc] initWithMachServiceName:label options:NSXPCConnectionPrivileged];

    // Configure the remote interface (what the helper exports)
    NSXPCInterface *remoteIface = [NSXPCInterface interfaceWithProtocol:@protocol(BBPrivilegedHelper)];

    // Tell XPC about the progress reporter proxy parameter
    NSXPCInterface *progressIface = [NSXPCInterface interfaceWithProtocol:@protocol(BBProgressReporter)];
    [remoteIface setInterface:progressIface
                  forSelector:@selector(writeLinuxISO:isoPath:progressReporter:reply:)
                argumentIndex:2
                      ofReply:NO];

    conn.remoteObjectInterface = remoteIface;

    // Configure our exported interface (what we export for the helper to call back)
    NSXPCInterface *exportedIface = [NSXPCInterface interfaceWithProtocol:@protocol(BBProgressReporter)];
    conn.exportedInterface = exportedIface;

    conn.invalidationHandler = ^{
        NSLog(@"[XPC Client] Connection invalidated");
    };
    conn.interruptionHandler = ^{
        NSLog(@"[XPC Client] Connection interrupted");
    };
    [conn resume];
    return conn;
}

static void setError(NSString *msg, char **errmsg) {
    if (!errmsg) return;
    if (*errmsg) { free(*errmsg); *errmsg = NULL; }
    *errmsg = strdup(msg.UTF8String);
}

int helper_ensure_ready(char **errmsg) {
    NSLog(@"[helper_ensure_ready] start");
    // If already present, consider it ready.
    CFStringRef label = CFStringCreateWithCString(NULL, kHelperLabel, kCFStringEncodingUTF8);
    NSDictionary *job = (__bridge_transfer NSDictionary *)SMJobCopyDictionary(kSMDomainSystemLaunchd, label);
    if (label) CFRelease(label);
    if (job) {
        NSLog(@"[helper_ensure_ready] job present: %@", job);
        return 0;
    }
    NSLog(@"[helper_ensure_ready] job not present; requesting authorization and blessing");

    // Request authorization to bless the helper
    AuthorizationItem right = {kSMRightBlessPrivilegedHelper, 0, NULL, 0};
    AuthorizationRights rights = {1, &right};
    AuthorizationRef authRef = NULL;
    AuthorizationFlags flags = kAuthorizationFlagDefaults |
                               kAuthorizationFlagInteractionAllowed |
                               kAuthorizationFlagExtendRights |
                               kAuthorizationFlagPreAuthorize;
    OSStatus status = AuthorizationCreate(&rights, kAuthorizationEmptyEnvironment, flags, &authRef);
    if (status != errAuthorizationSuccess) {
        NSLog(@"[helper_ensure_ready] AuthorizationCreate failed: %d", (int)status);
        setError([NSString stringWithFormat:@"AuthorizationCreate failed: %d", (int)status], errmsg);
        if (authRef) AuthorizationFree(authRef, kAuthorizationFlagDefaults);
        return 1;
    }

    CFErrorRef cferr = NULL;
    CFStringRef bundleID = CFStringCreateWithCString(NULL, kHelperBundleID, kCFStringEncodingUTF8);
    NSLog(@"[helper_ensure_ready] calling SMJobBless for %@", (__bridge NSString *)bundleID);
    Boolean ok = SMJobBless(kSMDomainSystemLaunchd, bundleID, authRef, &cferr);
    if (bundleID) CFRelease(bundleID);
    AuthorizationFree(authRef, kAuthorizationFlagDefaults);
    if (!ok) {
        NSString *msg = @"SMJobBless failed";
        if (cferr) {
            msg = [NSString stringWithFormat:@"SMJobBless failed: %@", (__bridge NSError *)cferr];
            CFRelease(cferr);
        }
        NSLog(@"[helper_ensure_ready] %@", msg);
        setError(msg, errmsg);
        return 1;
    }
    NSLog(@"[helper_ensure_ready] SMJobBless ok");
    return 0;
}

// Write a Linux ISO to a disk using Disk Arbitration and direct I/O
// This is a long-running operation (several minutes for large ISOs)
// Returns: 0 on success, 1 on error
int helper_write_linux_iso(const char *device, const char *isoPath, char **errmsg) {
    NSLog(@"[helper_write_linux_iso] device=%@ isoPath=%@",
          [NSString stringWithUTF8String:device],
          [NSString stringWithUTF8String:isoPath]);

    NSXPCConnection *conn = create_helper_connection();
    if (!conn) {
        setError(@"failed to create XPC connection", errmsg);
        return 1;
    }

    // Store connection globally so cancel can use it
    _activeConnection = conn;

    // Create progress reporter and export it on the connection
    WriteProgressReporter *reporter = [[WriteProgressReporter alloc] init];
    conn.exportedObject = reporter;

    __block BOOL success = NO;
    __block NSString *serr = nil;
    dispatch_semaphore_t sema = dispatch_semaphore_create(0);

    id<BBPrivilegedHelper> proxy = [conn remoteObjectProxyWithErrorHandler:^(NSError *error) {
        NSLog(@"[helper_write_linux_iso] XPC error: %@", error);
        serr = [NSString stringWithFormat:@"XPC error: %@", error.localizedDescription];
        dispatch_semaphore_signal(sema);
    }];

    // Pass the progress reporter proxy to the helper
    // The helper will call updateProgress: on it, which bridges back to Go via GoProgressCallback
    [proxy writeLinuxISO:[NSString stringWithUTF8String:device]
                 isoPath:[NSString stringWithUTF8String:isoPath]
        progressReporter:reporter
                   reply:^(BOOL s, NSString *e) {
        success = s;
        serr = e;
        dispatch_semaphore_signal(sema);
    }];

    // Very long timeout - direct I/O can take many minutes for large ISOs (4-6GB)
    long timeout = dispatch_semaphore_wait(sema, dispatch_time(DISPATCH_TIME_NOW, (int64_t)(3600 * NSEC_PER_SEC)));

    // Clear active connection and invalidate
    _activeConnection = nil;
    [conn invalidate];

    if (timeout != 0) {
        setError(@"timeout waiting for helper (>1 hour)", errmsg);
        return 1;
    }

    if (!success && serr) {
        setError(serr, errmsg);
        return 1;
    }

    NSLog(@"[helper_write_linux_iso] success=%d", success);
    return success ? 0 : 1;
}

// Format a disk with the specified filesystem and volume name
// Returns: 0 on success, 1 on error
int helper_format_disk(const char *device, const char *filesystem, const char *volumeName, char **errmsg) {
    NSLog(@"[helper_format_disk] device=%@ filesystem=%@ volumeName=%@",
          [NSString stringWithUTF8String:device],
          [NSString stringWithUTF8String:filesystem],
          [NSString stringWithUTF8String:volumeName]);

    NSXPCConnection *conn = create_helper_connection();
    if (!conn) {
        setError(@"failed to create XPC connection", errmsg);
        return 1;
    }

    __block BOOL success = NO;
    __block NSString *serr = nil;
    dispatch_semaphore_t sema = dispatch_semaphore_create(0);

    id<BBPrivilegedHelper> proxy = [conn remoteObjectProxyWithErrorHandler:^(NSError *error) {
        NSLog(@"[helper_format_disk] XPC error: %@", error);
        serr = [NSString stringWithFormat:@"XPC error: %@", error.localizedDescription];
        dispatch_semaphore_signal(sema);
    }];

    [proxy formatDisk:[NSString stringWithUTF8String:device]
           filesystem:[NSString stringWithUTF8String:filesystem]
           volumeName:[NSString stringWithUTF8String:volumeName]
                reply:^(BOOL s, NSString *e) {
        success = s;
        serr = e;
        dispatch_semaphore_signal(sema);
    }];

    // Formatting typically takes less than a minute
    long timeout = dispatch_semaphore_wait(sema, dispatch_time(DISPATCH_TIME_NOW, (int64_t)(300 * NSEC_PER_SEC)));
    [conn invalidate];

    if (timeout != 0) {
        setError(@"timeout waiting for helper (>5 minutes)", errmsg);
        return 1;
    }

    if (!success && serr) {
        setError(serr, errmsg);
        return 1;
    }

    NSLog(@"[helper_format_disk] success=%d", success);
    return success ? 0 : 1;
}

// Cancel the currently running operation (if any).
// Sends the cancel request over the same connection as the active operation.
// Safe to call even if no operation is running (no-op).
void helper_cancel_current_operation(void) {
    NSXPCConnection *conn = _activeConnection;
    if (!conn) {
        NSLog(@"[helper_cancel_current_operation] No active connection - ignoring");
        return;
    }

    NSLog(@"[helper_cancel_current_operation] Sending cancel request");

    id<BBPrivilegedHelper> proxy = [conn remoteObjectProxyWithErrorHandler:^(NSError *error) {
        NSLog(@"[helper_cancel_current_operation] XPC error: %@", error);
    }];

    // Fire and forget - the operation's reply block will be called with the cancel error
    [proxy cancelCurrentOperation];
}
