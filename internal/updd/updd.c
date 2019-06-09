#include "_cgo_export.h"

#include "upddapi.h"

void TBCALL touchCallback(unsigned long context, _PointerEvent* ev) {
	switch (ev->type) {
	case _EventTypeUnload:
		//[upddapi] driver is being uninstalled - all api programs must terminate
		TBApiClose();
		// exit(0);
		break;
	case _EventTypeDigitiserEvent:
		callback(ev->hStylus, ev->pe.digitiserEvent.screenx, ev->pe.digitiserEvent.screeny, ev->pe.digitiserEvent.de.touchEvent.touchingLeft);
		break;
	}
}

void TBCALL connectCallback(unsigned long context, _PointerEvent* ev) {
	if (ev->pe.config.configEventType == CONFIG_EVENT_CONNECT) {
		TBApiRegisterEvent(0,  // specify 0 to receive callbacks for all active devices or a handle for a specific device
			0,    // this value gets passed to the callback function
			_EventTypeUnload |        // notification that we must unload (during driver uninstall)
			_EventTypeDigitiserEvent, // digitiser events give information related to touches
			touchCallback);
	}
}
