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

@protocol BBPrivilegedHelper
- (void)unmountDisk:(NSString *)device reply:(void (^)(NSInteger status, NSString *out, NSString *err))reply;
- (void)ejectDisk:(NSString *)device reply:(void (^)(NSInteger status, NSString *out, NSString *err))reply;
- (void)writeLinuxISO:(NSString *)device
              isoPath:(NSString *)isoPath
                reply:(void (^)(BOOL success, NSString *err))reply;
@end

static NSXPCConnection *create_helper_connection(void) {
    NSString *label = [NSString stringWithUTF8String:kHelperLabel];
    NSXPCConnection *conn = [[NSXPCConnection alloc] initWithMachServiceName:label options:NSXPCConnectionPrivileged];
    NSXPCInterface *iface = [NSXPCInterface interfaceWithProtocol:@protocol(BBPrivilegedHelper)];
    conn.remoteObjectInterface = iface;
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

int helper_unmount_disk(const char *device, char **out, char **errmsg) {
    NSLog(@"[helper_unmount_disk] %@", [NSString stringWithUTF8String:device]);
    NSXPCConnection *conn = create_helper_connection();
    if (!conn) { setError(@"failed to create XPC connection", errmsg); return 1; }
    id<BBPrivilegedHelper> proxy = [conn remoteObjectProxy];
    __block NSInteger status = 1;
    __block NSString *sout = nil;
    __block NSString *serr = @"invalid reply";
    dispatch_semaphore_t sema = dispatch_semaphore_create(0);
    [proxy unmountDisk:[NSString stringWithUTF8String:device] reply:^(NSInteger st, NSString *o, NSString *e) {
        status = st;
        sout = o;
        serr = e;
        dispatch_semaphore_signal(sema);
    }];
    dispatch_semaphore_wait(sema, dispatch_time(DISPATCH_TIME_NOW, (int64_t)(60 * NSEC_PER_SEC)));
    [conn invalidate];
    NSLog(@"[helper_unmount_disk] status=%ld", (long)status);
    if (status == 0) {
        if (sout) { *out = strdup(sout.UTF8String); }
        return 0;
    }
    if (serr) setError(serr, errmsg);
    return (int)status ?: 1;
}

int helper_eject_disk(const char *device, char **out, char **errmsg) {
    NSLog(@"[helper_eject_disk] %@", [NSString stringWithUTF8String:device]);
    NSXPCConnection *conn = create_helper_connection();
    if (!conn) { setError(@"failed to create XPC connection", errmsg); return 1; }
    id<BBPrivilegedHelper> proxy = [conn remoteObjectProxy];
    __block NSInteger status = 1;
    __block NSString *sout = nil;
    __block NSString *serr = @"invalid reply";
    dispatch_semaphore_t sema = dispatch_semaphore_create(0);
    [proxy ejectDisk:[NSString stringWithUTF8String:device] reply:^(NSInteger st, NSString *o, NSString *e) {
        status = st;
        sout = o;
        serr = e;
        dispatch_semaphore_signal(sema);
    }];
    dispatch_semaphore_wait(sema, dispatch_time(DISPATCH_TIME_NOW, (int64_t)(60 * NSEC_PER_SEC)));
    [conn invalidate];
    NSLog(@"[helper_eject_disk] status=%ld", (long)status);
    if (status == 0) {
        if (sout) { *out = strdup(sout.UTF8String); }
        return 0;
    }
    if (serr) setError(serr, errmsg);
    return (int)status ?: 1;
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

    __block BOOL success = NO;
    __block NSString *serr = nil;
    dispatch_semaphore_t sema = dispatch_semaphore_create(0);

    id<BBPrivilegedHelper> proxy = [conn remoteObjectProxyWithErrorHandler:^(NSError *error) {
        NSLog(@"[helper_write_linux_iso] XPC error: %@", error);
        serr = [NSString stringWithFormat:@"XPC error: %@", error.localizedDescription];
        dispatch_semaphore_signal(sema);
    }];

    // Note: Progress is logged to system log by the helper.
    // XPC only allows one reply block per message, so we can't have a progress callback.
    [proxy writeLinuxISO:[NSString stringWithUTF8String:device]
                 isoPath:[NSString stringWithUTF8String:isoPath]
                   reply:^(BOOL s, NSString *e) {
        success = s;
        serr = e;
        dispatch_semaphore_signal(sema);
    }];

    // Very long timeout - direct I/O can take many minutes for large ISOs (4-6GB)
    long timeout = dispatch_semaphore_wait(sema, dispatch_time(DISPATCH_TIME_NOW, (int64_t)(3600 * NSEC_PER_SEC)));
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
