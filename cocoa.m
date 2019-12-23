// +build darwin
// +build 386 amd64
// +build !ios

#include "_cgo_export.h"
#include <pthread.h>
#include <stdio.h>

#import <Cocoa/Cocoa.h>
#import <Foundation/Foundation.h>

void makeCurrentContext(uintptr_t context) {
	id ctx = (NSOpenGLContext*)context;
	[ctx makeCurrentContext];
}

void flushContext(uintptr_t context) {
	id ctx = (NSOpenGLContext*)context;
	[ctx flushBuffer];
}

@interface ScreenGLView : NSOpenGLView<NSWindowDelegate>
{
}
@end

@implementation ScreenGLView
- (void)prepareOpenGL {
	[super prepareOpenGL];

	[self setWantsBestResolutionOpenGLSurface:YES];

	NSOpenGLContext *ctx = self.openGLContext;

	GLint swapInt = 0;
#pragma clang diagnostic push
#pragma clang diagnostic ignored "-Wdeprecated-declarations"
	[ctx setValues:&swapInt forParameter:NSOpenGLCPSwapInterval];
#pragma clang diagnostic pop

	preparedOpenGL((uintptr_t)self, (uintptr_t)ctx);
}

- (void)callResize {
	NSScreen *screen = self.window.screen;
	CGDirectDisplayID display = (CGDirectDisplayID)[[screen.deviceDescription valueForKey:@"NSScreenNumber"] intValue];
	CGSize screenSize = CGDisplayScreenSize(display);
	CGSize screenSizePx = screen.frame.size;
	double pxWidth = screenSize.width / screenSizePx.width;
	double pxHeight = screenSize.height / screenSizePx.height;

	CGSize sizePx = self.bounds.size;
	double width = pxWidth * sizePx.width;
	double height = pxHeight * sizePx.height;

	resize((GoUintptr)self, width, height, pxWidth, pxHeight);
}

- (void)reshape {
	[super reshape];
	[self callResize];
}

// - (void)drawRect:(NSRect)theRect {
// 	// Called during resize. Do an extra draw if we are visible.
// 	// This gets rid of flicker when resizing.
// 	drawgl((GoUintptr)self);
// }

- (void)mouseEventNS:(NSEvent *)theEvent {
	NSPoint p = theEvent.locationInWindow;
	mouseEvent((GoUintptr)self, p.x, p.y, theEvent.type, theEvent.buttonNumber, theEvent.modifierFlags);
}

- (void)mouseMoved:(NSEvent *)theEvent        { [self mouseEventNS:theEvent]; }
- (void)mouseDown:(NSEvent *)theEvent         { [self mouseEventNS:theEvent]; }
- (void)mouseDragged:(NSEvent *)theEvent      { [self mouseEventNS:theEvent]; }
- (void)mouseUp:(NSEvent *)theEvent           { [self mouseEventNS:theEvent]; }
- (void)rightMouseDown:(NSEvent *)theEvent    { [self mouseEventNS:theEvent]; }
- (void)rightMouseDragged:(NSEvent *)theEvent { [self mouseEventNS:theEvent]; }
- (void)rightMouseUp:(NSEvent *)theEvent      { [self mouseEventNS:theEvent]; }
- (void)otherMouseDown:(NSEvent *)theEvent    { [self mouseEventNS:theEvent]; }
- (void)otherMouseDragged:(NSEvent *)theEvent { [self mouseEventNS:theEvent]; }
- (void)otherMouseUp:(NSEvent *)theEvent      { [self mouseEventNS:theEvent]; }

// - (void)scrollWheel:(NSEvent *)theEvent       { [self mouseEventNS:theEvent]; }

// // raw modifier key presses
// - (void)flagsChanged:(NSEvent *)theEvent {
// 	flagEvent((GoUintptr)self, theEvent.modifierFlags);
// }

// // overrides special handling of escape and tab
// - (BOOL)performKeyEquivalent:(NSEvent *)theEvent {
// 	[self key:theEvent];
// 	return YES;
// }

// - (void)keyDown:(NSEvent *)theEvent { [self key:theEvent]; }
// - (void)keyUp:(NSEvent *)theEvent   { [self key:theEvent]; }

// - (void)key:(NSEvent *)theEvent {
// 	NSRange range = [theEvent.characters rangeOfComposedCharacterSequenceAtIndex:0];

// 	uint8_t buf[4] = {0, 0, 0, 0};
// 	if (![theEvent.characters getBytes:buf
// 			maxLength:4
// 			usedLength:nil
// 			encoding:NSUTF32LittleEndianStringEncoding
// 			options:NSStringEncodingConversionAllowLossy
// 			range:range
// 			remainingRange:nil]) {
// 		NSLog(@"failed to read key event %@", theEvent);
// 		return;
// 	}

// 	uint32_t rune = (uint32_t)buf[0]<<0 | (uint32_t)buf[1]<<8 | (uint32_t)buf[2]<<16 | (uint32_t)buf[3]<<24;

// 	uint8_t direction;
// 	if ([theEvent isARepeat]) {
// 		direction = 0;
// 	} else if (theEvent.type == NSEventTypeKeyDown) {
// 		direction = 1;
// 	} else {
// 		direction = 2;
// 	}
// 	keyEvent((GoUintptr)self, (int32_t)rune, direction, theEvent.keyCode, theEvent.modifierFlags);
// }

- (void)windowDidChangeScreenProfile:(NSNotification *)notification {
	[self callResize];
}

// // TODO: catch windowDidMiniaturize?

// - (void)windowDidExpose:(NSNotification *)notification {
// 	lifecycleVisible((GoUintptr)self, true);
// }

// - (void)windowDidBecomeKey:(NSNotification *)notification {
// 	lifecycleFocused((GoUintptr)self, true);
// }

// - (void)windowDidResignKey:(NSNotification *)notification {
// 	lifecycleFocused((GoUintptr)self, false);
// 	if ([NSApp isHidden]) {
// 		lifecycleVisible((GoUintptr)self, false);
// 	}
// }

// - (void)windowWillClose:(NSNotification *)notification {
// 	// TODO: is this right? Closing a window via the top-left red button
// 	// seems to return early without ever calling windowClosing.
// 	if (self.window.nextResponder == NULL) {
// 		return; // already called close
// 	}

// 	windowClosing((GoUintptr)self);
// 	[self.window.nextResponder release];
// 	self.window.nextResponder = NULL;
// }
@end

uintptr_t newWindow(double width, double height) {
	NSScreen *screen = NSScreen.mainScreen;
	CGDirectDisplayID display = (CGDirectDisplayID)[[screen.deviceDescription valueForKey:@"NSScreenNumber"] intValue];
	CGSize screenSize = CGDisplayScreenSize(display);
	CGSize screenSizePx = screen.frame.size;
	double pxWidth = screenSize.width / screenSizePx.width;
	double pxHeight = screenSize.height / screenSizePx.height;
	double w = width / pxWidth;
	double h = height / pxHeight;

	__block ScreenGLView* view = NULL;
	dispatch_sync(dispatch_get_main_queue(), ^{
		id menuBar = [NSMenu new];
		id menuItem = [NSMenuItem new];
		[menuBar addItem:menuItem];
		[NSApp setMainMenu:menuBar];

		id menu = [NSMenu new];
		id hideMenuItem = [[NSMenuItem alloc] initWithTitle:@"Hide" action:@selector(hide:) keyEquivalent:@"h"];
		id quitMenuItem = [[NSMenuItem alloc] initWithTitle:@"Quit" action:@selector(terminate:) keyEquivalent:@"q"];
		[menu addItem:hideMenuItem];
		[menu addItem:quitMenuItem];
		[menuItem setSubmenu:menu];

		NSRect rect = NSMakeRect(0, 0, w, h);

		NSWindow* window = [[NSWindow alloc] initWithContentRect:rect
				styleMask:NSWindowStyleMaskTitled
				backing:NSBackingStoreBuffered
				defer:NO];
		window.styleMask |= NSWindowStyleMaskResizable;
		window.styleMask |= NSWindowStyleMaskMiniaturizable;
		// window.styleMask |= NSWindowStyleMaskClosable;
		window.displaysWhenScreenProfileChanges = YES;
		[window cascadeTopLeftFromPoint:NSMakePoint(20,20)];
		// [window setCollectionBehavior:NSWindowCollectionBehaviorFullScreenPrimary];
		[window setAcceptsMouseMovedEvents:YES];

		NSOpenGLPixelFormatAttribute attr[] = {
			NSOpenGLPFAOpenGLProfile, NSOpenGLProfileVersion3_2Core,
			NSOpenGLPFAColorSize,     24,
			NSOpenGLPFAAlphaSize,     8,
			NSOpenGLPFADepthSize,     16,
			NSOpenGLPFADoubleBuffer,
			NSOpenGLPFAAllowOfflineRenderers,
			0
		};
		id pixFormat = [[NSOpenGLPixelFormat alloc] initWithAttributes:attr];
		view = [[ScreenGLView alloc] initWithFrame:rect pixelFormat:pixFormat];
		[window setContentView:view];
		[window setDelegate:view];
		[window makeFirstResponder:view];

		// [window toggleFullScreen:window];
		[window makeKeyAndOrderFront:window];
	});

	return (uintptr_t)view;
}

@interface AppDelegate : NSObject<NSApplicationDelegate>
{
}
@end

@implementation AppDelegate
- (void)applicationDidFinishLaunching:(NSNotification *)aNotification {
	applicationDidFinishLaunching();
	[[NSRunningApplication currentApplication] activateWithOptions:(NSApplicationActivateAllWindows | NSApplicationActivateIgnoringOtherApps)];
}

// - (void)applicationWillTerminate:(NSNotification *)aNotification {
// 	lifecycleDeadAll();
// }

// - (void)applicationWillHide:(NSNotification *)aNotification {
// 	lifecycleHideAll();
// }
@end

void runApp() {
	[NSAutoreleasePool new];
	[NSApplication sharedApplication];
	[NSApp setActivationPolicy:NSApplicationActivationPolicyRegular];
	// [NSApp setPresentationOptions:NSApplicationPresentationFullScreen];
	AppDelegate* delegate = [[AppDelegate alloc] init];
	[NSApp setDelegate:delegate];
	[NSApp run];
}

NSPoint mapFromScreen(uintptr_t window, NSPoint pt) {
	ScreenGLView *v = (ScreenGLView*)window;
	pt.y = v.window.screen.frame.size.height - pt.y; // TODO: This adjustment doesn't belong here.
	return [v.window convertPointFromScreen: pt];
}

uint64 threadID() {
	uint64 id;
	if (pthread_threadid_np(pthread_self(), &id)) {
		abort();
	}
	return id;
}
