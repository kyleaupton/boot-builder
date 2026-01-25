//
//  BBPrivilegedHelper.h
//  FlashIt Privileged Helper
//
//  XPC protocol definition for privileged disk operations.
//
//  Design principle: Use Disk Arbitration framework to properly claim exclusive
//  disk access before performing raw I/O operations.
//

#import <Foundation/Foundation.h>

/// Protocol for receiving progress updates during write operations.
/// The client implements this and passes a proxy to the helper.
@protocol BBProgressReporter <NSObject>
- (void)updateProgress:(uint64_t)bytesWritten totalBytes:(uint64_t)totalBytes;
@end

@protocol BBPrivilegedHelper

/// Write a Linux ISO to a disk using Disk Arbitration and direct I/O
/// Pipeline: DA claim → unmount → raw write → eject → unclaim
/// @param device Target device path (e.g., "/dev/disk4") - must be whole disk
/// @param isoPath Path to the source ISO file
/// @param progressReporter Proxy object for receiving progress updates (can be nil)
/// @param reply Callback with (success, error message) - called once at end
- (void)writeLinuxISO:(NSString *)device
              isoPath:(NSString *)isoPath
     progressReporter:(id<BBProgressReporter>)reporter
                reply:(void (^)(BOOL success, NSString *error))reply;

/// Format a disk with the specified filesystem
/// @param device Target device path (e.g., "/dev/disk4") - must be whole disk
/// @param filesystem Filesystem type (e.g., "FAT32", "ExFAT", "APFS", "HFS+")
/// @param volumeName Label for the new volume
/// @param reply Callback with (success, error message) - called once at end
- (void)formatDisk:(NSString *)device
        filesystem:(NSString *)filesystem
        volumeName:(NSString *)volumeName
             reply:(void (^)(BOOL success, NSString *error))reply;

/// Cancel the currently running operation (if any).
/// Idempotent - safe to call multiple times or when no operation is running.
/// The cancelled operation's reply block will be called with success=NO and
/// an error message indicating cancellation.
- (void)cancelCurrentOperation;

@end
