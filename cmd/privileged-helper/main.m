//
//  main.m
//  Boot Builder Privileged Helper
//
//  XPC service main entry point.
//

#import <Foundation/Foundation.h>
#import "BBPrivilegedHelper.h"

// Forward declaration of the service class
@interface BBPrivilegedHelperService : NSObject <NSXPCListenerDelegate, BBPrivilegedHelper>
@end

int main(int argc, const char *argv[]) {
    @autoreleasepool {
        NSLog(@"[Helper] Starting privileged helper service");

        // Create XPC listener with our Mach service name
        NSXPCListener *listener = [[NSXPCListener alloc]
            initWithMachServiceName:@"dev.kyleupton.boot-builder.helper"];

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
