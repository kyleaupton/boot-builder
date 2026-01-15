//
//  BBPrivilegedHelper.m
//  Boot Builder Privileged Helper
//
//  XPC service implementation for privileged disk operations.
//

#import <Foundation/Foundation.h>
#import "BBPrivilegedHelper.h"

@interface BBPrivilegedHelperService : NSObject <NSXPCListenerDelegate, BBPrivilegedHelper>
@end

@implementation BBPrivilegedHelperService

#pragma mark - NSXPCListenerDelegate

- (BOOL)listener:(NSXPCListener *)listener shouldAcceptNewConnection:(NSXPCConnection *)connection {
    NSLog(@"[Helper] Received connection from PID: %d", connection.processIdentifier);

    // Configure the connection
    NSXPCInterface *interface = [NSXPCInterface interfaceWithProtocol:@protocol(BBPrivilegedHelper)];

    // Allow NSFileHandle to be received over XPC for rawWrite
    NSSet *allowedClasses = [NSSet setWithObjects:[NSFileHandle class], nil];
    [interface setClasses:allowedClasses forSelector:@selector(rawWrite:device:reply:) argumentIndex:0 ofReply:NO];

    connection.exportedInterface = interface;
    connection.exportedObject = self;

    // XPC will verify code signature based on SMAuthorizedClients in Info.plist
    // No additional verification needed (SMJobBless handles this)

    connection.invalidationHandler = ^{
        NSLog(@"[Helper] Connection invalidated");
    };

    connection.interruptionHandler = ^{
        NSLog(@"[Helper] Connection interrupted");
    };

    [connection resume];
    return YES;
}

#pragma mark - BBPrivilegedHelper Protocol

- (void)unmountDisk:(NSString *)device reply:(void (^)(NSInteger, NSString *, NSString *))reply {
    NSLog(@"[Helper] unmountDisk: %@", device);

    // Validate input
    if (!device || device.length == 0) {
        reply(1, @"", @"Device not specified");
        return;
    }

    if (![device hasPrefix:@"/dev/"]) {
        reply(1, @"", @"Invalid device path - must start with /dev/");
        return;
    }

    // Execute diskutil unmountDisk force <device>
    NSTask *task = [[NSTask alloc] init];
    task.launchPath = @"/usr/sbin/diskutil";
    task.arguments = @[@"unmountDisk", @"force", device];

    NSPipe *outPipe = [NSPipe pipe];
    NSPipe *errPipe = [NSPipe pipe];
    task.standardOutput = outPipe;
    task.standardError = errPipe;

    NSError *error = nil;
    [task launchAndReturnError:&error];

    if (error) {
        NSLog(@"[Helper] Failed to launch diskutil: %@", error);
        reply(1, @"", [NSString stringWithFormat:@"Failed to launch diskutil: %@", error.localizedDescription]);
        return;
    }

    [task waitUntilExit];

    NSData *outData = [[outPipe fileHandleForReading] readDataToEndOfFile];
    NSData *errData = [[errPipe fileHandleForReading] readDataToEndOfFile];

    NSString *outStr = [[NSString alloc] initWithData:outData encoding:NSUTF8StringEncoding] ?: @"";
    NSString *errStr = [[NSString alloc] initWithData:errData encoding:NSUTF8StringEncoding] ?: @"";

    NSInteger status = task.terminationStatus;
    NSLog(@"[Helper] unmountDisk completed with status: %ld", (long)status);

    reply(status, outStr, errStr);
}

- (void)ejectDisk:(NSString *)device reply:(void (^)(NSInteger, NSString *, NSString *))reply {
    NSLog(@"[Helper] ejectDisk: %@", device);

    // Validate input
    if (!device || device.length == 0) {
        reply(1, @"", @"Device not specified");
        return;
    }

    if (![device hasPrefix:@"/dev/"]) {
        reply(1, @"", @"Invalid device path - must start with /dev/");
        return;
    }

    // Execute diskutil eject <device>
    NSTask *task = [[NSTask alloc] init];
    task.launchPath = @"/usr/sbin/diskutil";
    task.arguments = @[@"eject", device];

    NSPipe *outPipe = [NSPipe pipe];
    NSPipe *errPipe = [NSPipe pipe];
    task.standardOutput = outPipe;
    task.standardError = errPipe;

    NSError *error = nil;
    [task launchAndReturnError:&error];

    if (error) {
        NSLog(@"[Helper] Failed to launch diskutil: %@", error);
        reply(1, @"", [NSString stringWithFormat:@"Failed to launch diskutil: %@", error.localizedDescription]);
        return;
    }

    [task waitUntilExit];

    NSData *outData = [[outPipe fileHandleForReading] readDataToEndOfFile];
    NSData *errData = [[errPipe fileHandleForReading] readDataToEndOfFile];

    NSString *outStr = [[NSString alloc] initWithData:outData encoding:NSUTF8StringEncoding] ?: @"";
    NSString *errStr = [[NSString alloc] initWithData:errData encoding:NSUTF8StringEncoding] ?: @"";

    NSInteger status = task.terminationStatus;
    NSLog(@"[Helper] ejectDisk completed with status: %ld", (long)status);

    reply(status, outStr, errStr);
}

- (void)rawWrite:(NSFileHandle *)isoFileHandle device:(NSString *)device reply:(void (^)(NSInteger, NSString *))reply {
    NSLog(@"[Helper] rawWrite: device=%@", device);

    // Validate inputs
    if (!isoFileHandle) {
        reply(1, @"ISO file handle not provided");
        return;
    }

    if (!device || device.length == 0) {
        reply(1, @"Device not specified");
        return;
    }

    // Safety check: prevent writing to disk0
    if ([device isEqualToString:@"/dev/disk0"] || [device hasSuffix:@"disk0"]) {
        NSLog(@"[Helper] BLOCKED: Refusing to write to disk0");
        reply(1, @"Refusing to write to /dev/disk0 (system disk)");
        return;
    }

    if (![device hasPrefix:@"/dev/"]) {
        reply(1, @"Invalid device path - must start with /dev/");
        return;
    }

    // Convert /dev/diskN to /dev/rdiskN for better performance
    NSString *rawDevice = device;
    if ([device hasPrefix:@"/dev/disk"]) {
        rawDevice = [device stringByReplacingOccurrencesOfString:@"/dev/disk" withString:@"/dev/rdisk"];
        NSLog(@"[Helper] Using raw device: %@", rawDevice);
    }

    // Open raw device for writing (requires root privileges)
    NSFileHandle *devHandle = [NSFileHandle fileHandleForWritingAtPath:rawDevice];
    if (!devHandle) {
        reply(1, [NSString stringWithFormat:@"Cannot open device for writing: %@", rawDevice]);
        return;
    }

    NSLog(@"[Helper] Starting raw write...");

    // Copy data in 4MB chunks
    const NSUInteger bufferSize = 4 * 1024 * 1024; // 4MB
    NSUInteger totalWritten = 0;
    NSData *chunk;

    @try {
        while (YES) {
            @autoreleasepool {
                chunk = [isoFileHandle readDataOfLength:bufferSize];
                if (chunk.length == 0) {
                    break; // EOF
                }

                [devHandle writeData:chunk];
                totalWritten += chunk.length;

                // Log progress every 100MB
                if (totalWritten % (100 * 1024 * 1024) == 0 || totalWritten < bufferSize) {
                    NSLog(@"[Helper] Wrote %lu MB", (unsigned long)(totalWritten / 1024 / 1024));
                }
            }
        }

        // Ensure all data is written to disk
        [devHandle synchronizeFile];

        NSLog(@"[Helper] Raw write complete: %lu bytes", (unsigned long)totalWritten);

    } @catch (NSException *exception) {
        NSLog(@"[Helper] Write exception: %@", exception);
        [devHandle closeFile];
        reply(1, [NSString stringWithFormat:@"Write error: %@", exception.reason]);
        return;
    }

    [devHandle closeFile];

    reply(0, @"");
}

@end
