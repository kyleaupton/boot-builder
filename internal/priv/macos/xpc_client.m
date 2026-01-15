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
- (void)rawWrite:(NSFileHandle *)isoFileHandle device:(NSString *)device reply:(void (^)(NSInteger status, NSString *err))reply;
@end

static NSXPCConnection *create_helper_connection(void) {
    NSString *label = [NSString stringWithUTF8String:kHelperLabel];
    NSXPCConnection *conn = [[NSXPCConnection alloc] initWithMachServiceName:label options:NSXPCConnectionPrivileged];
    NSXPCInterface *iface = [NSXPCInterface interfaceWithProtocol:@protocol(BBPrivilegedHelper)];

    // Allow NSFileHandle to be sent over XPC for rawWrite
    NSSet *allowedClasses = [NSSet setWithObjects:[NSFileHandle class], nil];
    [iface setClasses:allowedClasses forSelector:@selector(rawWrite:device:reply:) argumentIndex:0 ofReply:NO];

    conn.remoteObjectInterface = iface;
    conn.invalidationHandler = ^{
        // handle invalidation if needed
    };
    conn.interruptionHandler = ^{
        // handle interruption if needed
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

typedef void (*progress_cb_t)(long long wrote, long long total);
int helper_raw_write(const char *iso_path, const char *raw_device, progress_cb_t cb, char **errmsg) {
    NSLog(@"[helper_raw_write] iso=%@ device=%@", [NSString stringWithUTF8String:iso_path], [NSString stringWithUTF8String:raw_device]);

    // Open ISO file as the user (this process has access to user files)
    NSString *isoPathStr = [NSString stringWithUTF8String:iso_path];
    NSFileHandle *isoHandle = [NSFileHandle fileHandleForReadingAtPath:isoPathStr];
    if (!isoHandle) {
        setError([NSString stringWithFormat:@"Cannot open ISO file: %@", isoPathStr], errmsg);
        return 1;
    }

    NSXPCConnection *conn = create_helper_connection();
    if (!conn) {
        [isoHandle closeFile];
        setError(@"failed to create XPC connection", errmsg);
        return 1;
    }

    id<BBPrivilegedHelper> proxy = [conn remoteObjectProxy];
    __block NSInteger status = 1;
    __block NSString *serr = @"invalid reply";
    dispatch_semaphore_t sema = dispatch_semaphore_create(0);

    // Pass the file handle to the helper (XPC will transfer it)
    [proxy rawWrite:isoHandle device:[NSString stringWithUTF8String:raw_device] reply:^(NSInteger st, NSString *e) {
        status = st;
        serr = e;
        dispatch_semaphore_signal(sema);
    }];

    dispatch_semaphore_wait(sema, dispatch_time(DISPATCH_TIME_NOW, (int64_t)(60 * NSEC_PER_SEC)));
    [conn invalidate];
    [isoHandle closeFile];

    NSLog(@"[helper_raw_write] status=%ld", (long)status);
    if (status == 0) { return 0; }
    if (serr) setError(serr, errmsg);
    return (int)status ?: 1;
}


