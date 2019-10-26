package digitizer

import (
	"fmt"

	"github.com/flynn/hid"
)

func Run(callback func(id uint8, pressed bool, x, y uint16)) {
	devices, err := hid.Devices()
	if err != nil {
		fmt.Println("error getting devices:", err)
		return
	}
	var dev hid.Device
	for _, d := range devices {
		if d.ProductID == 0x109 {
			dev, err = d.Open()
			if err != nil {
				fmt.Println("error opening device :", err)
				return
			}
		}
	}
	if dev == nil {
		// fmt.Println("touchscreen device not found")
		return
	}

	type touch struct {
		down bool
		x, y uint16
	}
	var touches [10]touch

	for report := range dev.ReadCh() {
		for i := 1; i < 51; i += 5 {
			if report[i] == 0 {
				break
			}

			id := report[i]>>3 - 1
			t := touch{
				down: report[i]%2 == 1,
				x:    uint16(report[i+1]) + uint16(report[i+2])<<8,
				y:    uint16(report[i+3]) + uint16(report[i+4])<<8,
			}
			if touches[id] != t {
				touches[id] = t
				callback(id, t.down, t.x, t.y)
			}
		}
	}
}
