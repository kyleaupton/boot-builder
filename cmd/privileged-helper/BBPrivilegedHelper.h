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
/// @param isoFileHandle File handle for ISO file (opened by client)
/// @param device Device path (e.g., "/dev/disk4" - will be converted to /dev/rdisk4)
/// @param reply Callback with (status, error message)
- (void)rawWrite:(NSFileHandle *)isoFileHandle
          device:(NSString *)device
           reply:(void (^)(NSInteger status, NSString *error))reply;

@end
