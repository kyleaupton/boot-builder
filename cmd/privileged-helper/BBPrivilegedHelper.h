//
//  BBPrivilegedHelper.h
//  Boot Builder Privileged Helper
//
//  XPC protocol definition for privileged disk operations.
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

/// Write ISO file to raw block device
/// @param isoPath Path to ISO file
/// @param device Device path (e.g., "/dev/disk4" - will be converted to /dev/rdisk4)
/// @param reply Callback with (status, error message)
- (void)rawWrite:(NSString *)isoPath
          device:(NSString *)device
           reply:(void (^)(NSInteger status, NSString *error))reply;

@end
