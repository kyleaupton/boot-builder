//
//  main.m
//  FlashIt Privileged Helper
//
//  XPC service main entry point.
//

#import <Foundation/Foundation.h>
#import "BBPrivilegedHelper.h"

// Forward declaration of the service class
@interface BBPrivilegedHelperService : NSObject <NSXPCListenerDelegate, BBPrivilegedHelper>
@end

// Build version - update this when making changes to verify deployment
#define HELPER_BUILD_VERSION "2025-01-17-v16-apfs-container-fix"

int main(int argc, const char *argv[]) {
    @autoreleasepool {
        NSLog(@"[Helper] Starting privileged helper service (build: %s)", HELPER_BUILD_VERSION);

        // Create XPC listener with our Mach service name
        NSXPCListener *listener = [[NSXPCListener alloc]
            initWithMachServiceName:@"dev.kyleupton.flashit.helper"];

        // Create and set our service delegate
        BBPrivilegedHelperService *delegate = [[BBPrivilegedHelperService alloc] init];
        listener.delegate = delegate;

        // Start listening for connections
        [listener resume];

        NSLog(@"[Helper] Listening for XPC connections...");

        // Run the main run loop indefinitely
        [[NSRunLoop currentRunLoop] run];
    }

    return 0;
}
