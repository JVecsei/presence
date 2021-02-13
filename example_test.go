package presence

import (
	"context"
	"fmt"
)

func ExamplePresence_Scan() {
	mockBTScanner := &MockBTScanner{
		config: map[string]struct {
			found bool
			err   error
		}{
			"mymac":  {true, nil},
			"mymac2": {false, nil},
		},
	}
	p := New(mockBTScanner)

	p.RegisterAction("mymac", ActionFunc(func(mac string, present bool) {
		fmt.Printf("%s found: %t\n", mac, present)
	}))

	p.RegisterAction("mymac", ActionFunc(func(mac string, present bool) {
		fmt.Printf("%s found: %t\n", mac, present)
	}))

	p.RegisterAction("mymac2", ActionFunc(func(mac string, present bool) {
		fmt.Printf("%s found: %t\n", mac, present)
	}))

	p.Scan(context.Background())
	//Output:
	//mymac2 found: false
	//mymac found: true
	//mymac found: true
}
