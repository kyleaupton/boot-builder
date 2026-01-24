//
//  BBPrivilegedHelper.m
//  Boot Builder Privileged Helper
//
//  XPC service implementation for privileged disk operations.
//
//  Design principle: Use Disk Arbitration framework to properly claim exclusive
//  disk access before performing raw I/O operations. This is how tools like
//  balenaEtcher work.
//

#import <Foundation/Foundation.h>
#import <DiskArbitration/DiskArbitration.h>
#import <sys/stat.h>
#import "BBPrivilegedHelper.h"

// Build version for debugging
#define HELPER_BUILD_VERSION "2026-01-24-v18-cancellation"

// Buffer size for disk I/O (1 MB)
#define WRITE_BUFFER_SIZE (1024 * 1024)

// Progress report interval (every 10 MB)
#define PROGRESS_INTERVAL (10 * 1024 * 1024)

@interface BBPrivilegedHelperService : NSObject <NSXPCListenerDelegate, BBPrivilegedHelper> {
    id<NSObject> _activityToken;
}
/// Flag to signal cancellation to the running operation.
/// Atomic for thread-safety between the cancel call and the write loop.
@property (atomic) BOOL cancelRequested;
@end

@implementation BBPrivilegedHelperService

#pragma mark - NSXPCListenerDelegate

- (BOOL)listener:(NSXPCListener *)listener shouldAcceptNewConnection:(NSXPCConnection *)connection {
    NSLog(@"[Helper] Received connection from PID: %d (build: %s)", connection.processIdentifier, HELPER_BUILD_VERSION);

    // Configure the exported interface (what we expose to the client)
    NSXPCInterface *interface = [NSXPCInterface interfaceWithProtocol:@protocol(BBPrivilegedHelper)];

    // Configure the progress reporter proxy parameter for writeLinuxISO:isoPath:progressReporter:reply:
    // This tells XPC how to handle the proxy object at argument index 2 (0-indexed: device=0, isoPath=1, reporter=2)
    NSXPCInterface *progressInterface = [NSXPCInterface interfaceWithProtocol:@protocol(BBProgressReporter)];
    [interface setInterface:progressInterface
                forSelector:@selector(writeLinuxISO:isoPath:progressReporter:reply:)
              argumentIndex:2
                    ofReply:NO];

    connection.exportedInterface = interface;
    connection.exportedObject = self;

    // Capture weak reference to avoid retain cycle in blocks
    __weak BBPrivilegedHelperService *weakSelf = self;

    connection.invalidationHandler = ^{
        NSLog(@"[Helper] Connection invalidated - treating as cancel");
        weakSelf.cancelRequested = YES;
    };

    connection.interruptionHandler = ^{
        NSLog(@"[Helper] Connection interrupted - treating as cancel");
        weakSelf.cancelRequested = YES;
    };

    [connection resume];
    return YES;
}

#pragma mark - Helper Methods

/// Extract BSD name from device path (e.g., "/dev/disk4" or "/dev/rdisk4" → "disk4")
- (NSString *)bsdNameFromDevice:(NSString *)device {
    NSString *name = [device lastPathComponent]; // "disk4" or "rdisk4"
    if ([name hasPrefix:@"r"]) {
        name = [name substringFromIndex:1]; // "rdisk4" → "disk4"
    }
    return name;
}

/// Get raw device path from BSD name (e.g., "disk4" → "/dev/rdisk4")
- (NSString *)rawDeviceFromBSDName:(NSString *)bsdName {
    return [NSString stringWithFormat:@"/dev/r%@", bsdName];
}

/// Run a command and return its output
- (BOOL)runCommand:(NSString *)launchPath
         arguments:(NSArray<NSString *> *)arguments
            output:(NSString **)output
             error:(NSString **)error
            status:(NSInteger *)status {

    NSLog(@"[Helper] Running: %@ %@", launchPath, [arguments componentsJoinedByString:@" "]);

    NSTask *task = [[NSTask alloc] init];
    task.launchPath = launchPath;
    task.arguments = arguments;

    NSPipe *outPipe = [NSPipe pipe];
    NSPipe *errPipe = [NSPipe pipe];
    task.standardOutput = outPipe;
    task.standardError = errPipe;

    NSError *launchError = nil;
    [task launchAndReturnError:&launchError];

    if (launchError) {
        NSLog(@"[Helper] Failed to launch %@: %@", launchPath, launchError);
        if (error) *error = [NSString stringWithFormat:@"Failed to launch %@: %@", launchPath, launchError.localizedDescription];
        if (status) *status = -1;
        return NO;
    }

    [task waitUntilExit];

    NSData *outData = [[outPipe fileHandleForReading] readDataToEndOfFile];
    NSData *errData = [[errPipe fileHandleForReading] readDataToEndOfFile];

    NSString *outStr = [[NSString alloc] initWithData:outData encoding:NSUTF8StringEncoding] ?: @"";
    NSString *errStr = [[NSString alloc] initWithData:errData encoding:NSUTF8StringEncoding] ?: @"";

    NSInteger exitStatus = task.terminationStatus;
    NSLog(@"[Helper] %@ exited with status %ld", [launchPath lastPathComponent], (long)exitStatus);
    if (outStr.length > 0) NSLog(@"[Helper] stdout: %@", outStr);
    if (errStr.length > 0) NSLog(@"[Helper] stderr: %@", errStr);

    if (output) *output = outStr;
    if (error) *error = errStr;
    if (status) *status = exitStatus;

    return (exitStatus == 0);
}

/// Unmount a disk and any APFS containers that reference it.
/// APFS creates synthesized container disks (e.g., disk5 for physical disk4).
/// These must be unmounted before the physical disk or we get EBUSY.
- (BOOL)unmountDiskAndContainers:(NSString *)bsdName error:(NSString **)errorMsg {
    NSLog(@"[Helper] Unmounting disk and any APFS containers for %@...", bsdName);

    NSString *output = nil;
    NSString *error = nil;
    NSInteger status = 0;

    // Get disk list to find any APFS containers referencing this physical disk
    if (![self runCommand:@"/usr/sbin/diskutil"
                arguments:@[@"list"]
                   output:&output
                    error:&error
                   status:&status]) {
        NSLog(@"[Helper] Warning: diskutil list failed, continuing anyway");
    }

    if (output) {
        // Parse diskutil list output to find synthesized containers
        // Example output:
        // /dev/disk5 (synthesized):
        //    #:                       TYPE NAME                    SIZE       IDENTIFIER
        //    0:      APFS Container Scheme -                      +30.8 GB    disk5
        //                                  Physical Store disk4s2
        //    1:                APFS Volume Untitled                16.4 KB    disk5s1

        NSArray *lines = [output componentsSeparatedByString:@"\n"];
        NSString *currentDisk = nil;
        BOOL isSynthesized = NO;

        for (NSString *line in lines) {
            // Check for disk header line like "/dev/disk5 (synthesized):"
            if ([line hasPrefix:@"/dev/disk"]) {
                // Extract disk name
                NSRange parenRange = [line rangeOfString:@" ("];
                if (parenRange.location != NSNotFound) {
                    currentDisk = [[line substringToIndex:parenRange.location] lastPathComponent];
                    isSynthesized = [line containsString:@"(synthesized)"];
                    NSLog(@"[Helper] Found disk: %@ (synthesized: %@)", currentDisk, isSynthesized ? @"YES" : @"NO");
                } else {
                    currentDisk = nil;
                    isSynthesized = NO;
                }
                continue;
            }

            // If we're in a synthesized disk section, look for "Physical Store <bsdName>s"
            if (isSynthesized && currentDisk) {
                // Look for "Physical Store disk4s" pattern (with partition suffix)
                NSString *physicalStorePattern = [NSString stringWithFormat:@"Physical Store %@s", bsdName];
                if ([line containsString:physicalStorePattern]) {
                    NSLog(@"[Helper] Found APFS container %@ backed by %@, unmounting...", currentDisk, bsdName);

                    // Unmount this synthesized container first
                    NSString *containerDevice = [NSString stringWithFormat:@"/dev/%@", currentDisk];
                    [self runCommand:@"/usr/sbin/diskutil"
                           arguments:@[@"unmountDisk", @"force", containerDevice]
                              output:nil
                               error:nil
                              status:nil];
                    // Don't fail if unmount fails - it might already be unmounted
                }
            }
        }
    }

    // Now unmount the physical disk itself
    NSLog(@"[Helper] Unmounting physical disk %@...", bsdName);
    NSString *devicePath = [NSString stringWithFormat:@"/dev/%@", bsdName];
    if (![self runCommand:@"/usr/sbin/diskutil"
                arguments:@[@"unmountDisk", @"force", devicePath]
                   output:&output
                    error:&error
                   status:&status]) {
        // Check if it's just already unmounted
        NSString *combined = [NSString stringWithFormat:@"%@ %@", output ?: @"", error ?: @""];
        if (![combined.lowercaseString containsString:@"not mounted"] &&
            ![combined.lowercaseString containsString:@"already unmounted"]) {
            if (errorMsg) *errorMsg = [NSString stringWithFormat:@"Failed to unmount %@: %@", bsdName, error];
            return NO;
        }
        NSLog(@"[Helper] Disk already unmounted, continuing...");
    }

    // Wait for system to settle after unmounting
    NSLog(@"[Helper] Waiting 500ms for system to settle...");
    usleep(500000);

    return YES;
}

#pragma mark - Disk Arbitration Callbacks

static void unmountCallback(DADiskRef disk, DADissenterRef dissenter, void *context) {
    dispatch_semaphore_t sem = (__bridge dispatch_semaphore_t)context;
    if (dissenter) {
        DAReturn status = DADissenterGetStatus(dissenter);
        NSLog(@"[Helper] Unmount dissenter: status=0x%x", status);
    } else {
        NSLog(@"[Helper] Unmount succeeded");
    }
    dispatch_semaphore_signal(sem);
}

static void claimCallback(DADiskRef disk, DADissenterRef dissenter, void *context) {
    dispatch_semaphore_t sem = (__bridge dispatch_semaphore_t)context;
    if (dissenter) {
        DAReturn status = DADissenterGetStatus(dissenter);
        NSLog(@"[Helper] Claim dissenter: status=0x%x", status);
    } else {
        NSLog(@"[Helper] Claim succeeded");
    }
    dispatch_semaphore_signal(sem);
}

static void ejectCallback(DADiskRef disk, DADissenterRef dissenter, void *context) {
    dispatch_semaphore_t sem = (__bridge dispatch_semaphore_t)context;
    if (dissenter) {
        DAReturn status = DADissenterGetStatus(dissenter);
        NSLog(@"[Helper] Eject dissenter: status=0x%x", status);
    } else {
        NSLog(@"[Helper] Eject succeeded");
    }
    dispatch_semaphore_signal(sem);
}

// Claim release callback - called when something tries to steal our claim
static DADissenterRef claimReleaseCallback(DADiskRef disk, void *context) {
    NSLog(@"[Helper] Something is trying to release our disk claim - denying");
    // Return a dissenter to deny the release request
    return DADissenterCreate(kCFAllocatorDefault, kDAReturnBusy, CFSTR("Disk imaging in progress"));
}

#pragma mark - BBPrivilegedHelper Protocol

- (void)cancelCurrentOperation {
    NSLog(@"[Helper] Cancel requested (build: %s)", HELPER_BUILD_VERSION);
    self.cancelRequested = YES;
}

- (void)writeLinuxISO:(NSString *)device
              isoPath:(NSString *)isoPath
     progressReporter:(id<BBProgressReporter>)reporter
                reply:(void (^)(BOOL, NSString *))reply {

    NSLog(@"[Helper] writeLinuxISO: device=%@ iso=%@ reporter=%@ (build: %s)",
          device, isoPath, reporter ? @"present" : @"nil", HELPER_BUILD_VERSION);
    NSLog(@"[Helper] Running as UID: %d, EUID: %d", getuid(), geteuid());

    // Reset cancellation flag for new operation
    self.cancelRequested = NO;

    // Validate inputs
    if (!device || device.length == 0) {
        reply(NO, @"Device not specified");
        return;
    }

    if (!isoPath || isoPath.length == 0) {
        reply(NO, @"ISO path not specified");
        return;
    }

    // Safety check: prevent writing to disk0
    NSString *bsdName = [self bsdNameFromDevice:device];
    if ([bsdName isEqualToString:@"disk0"]) {
        NSLog(@"[Helper] BLOCKED: Refusing to write to disk0");
        reply(NO, @"Refusing to write to /dev/disk0 (system disk)");
        return;
    }

    // Verify ISO file exists and get its size
    struct stat st;
    if (stat([isoPath fileSystemRepresentation], &st) != 0) {
        int err = errno;
        reply(NO, [NSString stringWithFormat:@"Cannot stat ISO file: %s (errno %d)", strerror(err), err]);
        return;
    }
    uint64_t totalBytes = st.st_size;
    NSLog(@"[Helper] ISO size: %llu bytes (%.1f GB)", totalBytes, totalBytes / 1e9);

    // Begin activity to prevent helper termination during long operation
    _activityToken = [[NSProcessInfo processInfo] beginActivityWithOptions:NSActivityUserInitiated | NSActivityIdleSystemSleepDisabled
                                                                    reason:@"Writing Linux ISO to USB drive"];
    NSLog(@"[Helper] Started activity assertion");

    // Variables for cleanup
    DASessionRef session = NULL;
    DADiskRef disk = NULL;
    int srcFd = -1;
    int dstFd = -1;
    void *buffer = NULL;
    BOOL success = NO;
    NSString *errorMessage = nil;
    BOOL claimed = NO;

    // Step 1: Unmount disk and any APFS containers (must be done via diskutil before DA)
    // APFS creates synthesized container disks that hold the physical disk busy
    NSLog(@"[Helper] Step 1: Unmounting disk and APFS containers...");
    {
        NSString *unmountError = nil;
        if (![self unmountDiskAndContainers:bsdName error:&unmountError]) {
            errorMessage = unmountError ?: @"Failed to unmount disk";
            goto cleanup;
        }
    }

    // Step 2: Create Disk Arbitration session
    NSLog(@"[Helper] Step 2: Creating Disk Arbitration session...");
    session = DASessionCreate(kCFAllocatorDefault);
    if (!session) {
        errorMessage = @"Failed to create Disk Arbitration session";
        goto cleanup;
    }

    // Schedule session with current run loop
    DASessionScheduleWithRunLoop(session, CFRunLoopGetCurrent(), kCFRunLoopDefaultMode);

    // Step 3: Get disk reference from BSD name
    NSLog(@"[Helper] Step 3: Getting disk reference for %@...", bsdName);
    disk = DADiskCreateFromBSDName(kCFAllocatorDefault, session, [bsdName UTF8String]);
    if (!disk) {
        errorMessage = [NSString stringWithFormat:@"Failed to get disk reference for %@", bsdName];
        goto cleanup;
    }

    // Step 4: Claim exclusive access to the disk
    NSLog(@"[Helper] Step 4: Claiming exclusive access to disk...");
    {
        dispatch_semaphore_t sem = dispatch_semaphore_create(0);
        DADiskClaim(disk, kDADiskClaimOptionDefault, claimReleaseCallback, NULL, claimCallback, (__bridge void *)sem);

        // Run the run loop briefly to process the callback
        while (dispatch_semaphore_wait(sem, DISPATCH_TIME_NOW) != 0) {
            CFRunLoopRunInMode(kCFRunLoopDefaultMode, 0.1, true);
        }
        claimed = YES;
    }

    // Brief delay for system to settle after claim
    NSLog(@"[Helper] Waiting for system to settle...");
    usleep(300000); // 300ms

    // Step 5: Open source file
    NSLog(@"[Helper] Step 5: Opening source ISO...");
    srcFd = open([isoPath fileSystemRepresentation], O_RDONLY);
    if (srcFd < 0) {
        int err = errno;
        NSLog(@"[Helper] Failed to open ISO: %s (errno %d)", strerror(err), err);
        errorMessage = [NSString stringWithFormat:@"Failed to open ISO: %s (errno %d)", strerror(err), err];
        goto cleanup;
    }

    // Step 6: Open raw destination device
    NSLog(@"[Helper] Step 6: Opening raw device...");
    {
        NSString *rawPath = [self rawDeviceFromBSDName:bsdName];
        NSLog(@"[Helper] Using raw device: %@", rawPath);

        dstFd = open([rawPath fileSystemRepresentation], O_WRONLY);
        if (dstFd < 0) {
            int err = errno;
            NSLog(@"[Helper] Failed to open %@: %s (errno %d)", rawPath, strerror(err), err);
            errorMessage = [NSString stringWithFormat:@"Failed to open %@: %s (errno %d)", rawPath, strerror(err), err];
            goto cleanup;
        }
    }

    // Step 7: Allocate buffer and perform write
    NSLog(@"[Helper] Step 7: Writing ISO to disk (this may take several minutes)...");
    buffer = malloc(WRITE_BUFFER_SIZE);
    if (!buffer) {
        errorMessage = @"Failed to allocate write buffer";
        goto cleanup;
    }

    {
        uint64_t bytesWritten = 0;
        uint64_t lastProgressReport = 0;
        ssize_t bytesRead;

        while ((bytesRead = read(srcFd, buffer, WRITE_BUFFER_SIZE)) > 0) {
            // Check for cancellation before each chunk
            if (self.cancelRequested) {
                NSLog(@"[Helper] Cancellation detected at %llu bytes - stopping write", bytesWritten);
                errorMessage = @"cancelled:Operation cancelled by user";
                goto cleanup;
            }

            ssize_t totalWritten = 0;
            while (totalWritten < bytesRead) {
                ssize_t written = write(dstFd, (char *)buffer + totalWritten, bytesRead - totalWritten);
                if (written < 0) {
                    int err = errno;
                    NSLog(@"[Helper] Write failed at offset %llu: %s (errno %d)", bytesWritten, strerror(err), err);
                    errorMessage = [NSString stringWithFormat:@"Write failed: %s (errno %d)", strerror(err), err];
                    goto cleanup;
                }
                totalWritten += written;
            }

            bytesWritten += bytesRead;

            // Report progress periodically via the proxy
            if (bytesWritten - lastProgressReport >= PROGRESS_INTERVAL) {
                lastProgressReport = bytesWritten;
                NSLog(@"[Helper] Progress: %llu / %llu bytes (%.1f%%)",
                      bytesWritten, totalBytes, (double)bytesWritten / totalBytes * 100);

                // Call the progress reporter proxy (if provided)
                if (reporter) {
                    @try {
                        [reporter updateProgress:bytesWritten totalBytes:totalBytes];
                    } @catch (NSException *e) {
                        NSLog(@"[Helper] Progress reporter error: %@ (continuing write)", e.reason);
                    }
                }
            }
        }

        if (bytesRead < 0) {
            int err = errno;
            NSLog(@"[Helper] Read failed: %s (errno %d)", strerror(err), err);
            errorMessage = [NSString stringWithFormat:@"Read failed: %s (errno %d)", strerror(err), err];
            goto cleanup;
        }

        NSLog(@"[Helper] Write complete: %llu bytes written", bytesWritten);
    }

    // Sync to ensure data is flushed
    NSLog(@"[Helper] Syncing...");
    if (fsync(dstFd) != 0) {
        NSLog(@"[Helper] fsync warning: %s (continuing anyway)", strerror(errno));
    }

    success = YES;
    NSLog(@"[Helper] writeLinuxISO completed successfully");

cleanup:
    // Close file descriptors
    if (srcFd >= 0) {
        close(srcFd);
    }
    if (dstFd >= 0) {
        close(dstFd);
    }

    // Free buffer
    if (buffer) {
        free(buffer);
    }

    // Unclaim and eject disk
    if (disk) {
        if (claimed) {
            NSLog(@"[Helper] Unclaiming disk...");
            DADiskUnclaim(disk);
        }

        // Eject the disk
        NSLog(@"[Helper] Ejecting disk...");
        dispatch_semaphore_t sem = dispatch_semaphore_create(0);
        DADiskEject(disk, kDADiskEjectOptionDefault, ejectCallback, (__bridge void *)sem);

        // Run the run loop briefly to process the callback
        while (dispatch_semaphore_wait(sem, DISPATCH_TIME_NOW) != 0) {
            CFRunLoopRunInMode(kCFRunLoopDefaultMode, 0.1, true);
        }

        CFRelease(disk);
    }

    // Cleanup DA session
    if (session) {
        DASessionUnscheduleFromRunLoop(session, CFRunLoopGetCurrent(), kCFRunLoopDefaultMode);
        CFRelease(session);
    }

    // End activity assertion
    if (_activityToken) {
        [[NSProcessInfo processInfo] endActivity:_activityToken];
        _activityToken = nil;
        NSLog(@"[Helper] Ended activity assertion");
    }

    reply(success, errorMessage);
}

- (void)formatDisk:(NSString *)device
        filesystem:(NSString *)filesystem
        volumeName:(NSString *)volumeName
             reply:(void (^)(BOOL, NSString *))reply {

    NSLog(@"[Helper] formatDisk: device=%@ filesystem=%@ volumeName=%@ (build: %s)",
          device, filesystem, volumeName, HELPER_BUILD_VERSION);

    // Validate inputs
    if (!device || device.length == 0) {
        reply(NO, @"Device not specified");
        return;
    }

    if (!filesystem || filesystem.length == 0) {
        reply(NO, @"Filesystem not specified");
        return;
    }

    if (!volumeName || volumeName.length == 0) {
        reply(NO, @"Volume name not specified");
        return;
    }

    if (![device hasPrefix:@"/dev/"]) {
        reply(NO, @"Invalid device path - must start with /dev/");
        return;
    }

    // Safety check: prevent formatting disk0
    NSString *bsdName = [self bsdNameFromDevice:device];
    if ([bsdName isEqualToString:@"disk0"]) {
        NSLog(@"[Helper] BLOCKED: Refusing to format disk0");
        reply(NO, @"Refusing to format /dev/disk0 (system disk)");
        return;
    }

    // Validate filesystem type
    NSArray *validFilesystems = @[@"FAT32", @"ExFAT", @"APFS", @"HFS+", @"JHFS+", @"MS-DOS"];
    NSString *fsUpper = [filesystem uppercaseString];

    // Map common names to diskutil names
    NSString *diskutilFS = filesystem;
    if ([fsUpper isEqualToString:@"FAT32"]) {
        diskutilFS = @"FAT32";
    } else if ([fsUpper isEqualToString:@"EXFAT"]) {
        diskutilFS = @"ExFAT";
    } else if ([fsUpper isEqualToString:@"HFS+"]) {
        diskutilFS = @"HFS+";
    }

    BOOL validFS = NO;
    for (NSString *valid in validFilesystems) {
        if ([fsUpper isEqualToString:[valid uppercaseString]]) {
            validFS = YES;
            break;
        }
    }

    if (!validFS) {
        reply(NO, [NSString stringWithFormat:@"Unsupported filesystem: %@. Supported: FAT32, ExFAT, APFS, HFS+", filesystem]);
        return;
    }

    NSLog(@"[Helper] Formatting %@ as %@ with label '%@'...", device, diskutilFS, volumeName);

    NSString *output = nil;
    NSString *error = nil;
    NSInteger status = 0;

    // diskutil eraseDisk <filesystem> <volumeName> <device>
    // Use MBR partition scheme for FAT32 (better compatibility with Windows)
    // Use GPT for other filesystems
    NSString *partitionScheme = ([fsUpper isEqualToString:@"FAT32"] || [fsUpper isEqualToString:@"EXFAT"]) ? @"MBR" : @"GPT";

    BOOL success = [self runCommand:@"/usr/sbin/diskutil"
                          arguments:@[@"eraseDisk", diskutilFS, volumeName, partitionScheme, device]
                             output:&output
                              error:&error
                             status:&status];

    if (!success) {
        NSString *errMsg = [NSString stringWithFormat:@"diskutil eraseDisk failed (status %ld): %@ %@",
                           (long)status, output ?: @"", error ?: @""];
        NSLog(@"[Helper] %@", errMsg);
        reply(NO, errMsg);
        return;
    }

    NSLog(@"[Helper] formatDisk completed successfully");
    reply(YES, nil);
}

@end
