//
//  BBPrivilegedHelper.h
//  Boot Builder Privileged Helper
//
//  XPC protocol definition for privileged disk operations.
//
//  Design principle: Use Disk Arbitration framework to properly claim exclusive
//  disk access before performing raw I/O operations.
//

#import <Foundation/Foundation.h>

@protocol BBPrivilegedHelper

/// Unmount all volumes on a disk device
/// @param device Device path (e.g., "/dev/disk4")
/// @param reply Callback with (status, stdout, stderr)
- (void)unmountDisk:(NSString *)device
              reply:(void (^)(NSInteger status, NSString *output, NSString *error))reply;

/// Eject a disk device
/// @param device Device path (e.g., "/dev/disk4")
/// @param reply Callback with (status, stdout, stderr)
- (void)ejectDisk:(NSString *)device
            reply:(void (^)(NSInteger status, NSString *output, NSString *error))reply;

/// Write a Linux ISO to a disk using Disk Arbitration and direct I/O
/// Pipeline: DA claim → unmount → raw write → eject → unclaim
/// @param device Target device path (e.g., "/dev/disk4") - must be whole disk
/// @param isoPath Path to the source ISO file
/// @param reply Callback with (success, error message) - called once at end
/// Note: Progress is logged to system log. XPC only allows one reply block per message.
- (void)writeLinuxISO:(NSString *)device
              isoPath:(NSString *)isoPath
                reply:(void (^)(BOOL success, NSString *error))reply;

@end
