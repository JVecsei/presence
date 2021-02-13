package presence

import (
	"context"
	"testing"
)

type MockBTScanner struct {
	config map[string]struct {
		found bool
		err   error
	}
}

func (m MockBTScanner) IsPresent(context context.Context, mac string) (bool, error) {
	return m.config[mac].found, m.config[mac].err
}

func TestRegister(t *testing.T) {
	p := New(&MockBTScanner{})

	p.RegisterAction("mymac", ActionFunc(func(mac string, present bool) {
		//nothing
	}))
	p.RegisterAction("mymac", ActionFunc(func(mac string, present bool) {
		//nothing
	}))
	p.RegisterAction("mymac2", ActionFunc(func(mac string, present bool) {
		//nothing
	}))
}

func TestScan(t *testing.T) {
	p := New(&MockBTScanner{
		config: map[string]struct {
			found bool
			err   error
		}{
			"mymac":  {true, nil},
			"mymac2": {true, nil},
			"mymac3": {false, nil},
		},
	})
	foundMac := 0
	foundMac2 := 0
	foundMac3 := 0

	p.RegisterAction("mymac", ActionFunc(func(mac string, present bool) {
		if !present {
			t.Errorf("mac should not be present")
		}
		foundMac++
	}))
	p.RegisterAction("mymac", ActionFunc(func(mac string, present bool) {
		if !present {
			t.Errorf("mac2 should not be present")
		}
		foundMac++
	}))

	p.RegisterAction("mymac2", ActionFunc(func(mac string, present bool) {
		if !present {
			t.Errorf("mac2 should not be present")
		}
		foundMac2++
	}))

	p.RegisterAction("mymac3", ActionFunc(func(mac string, present bool) {
		if present {
			t.Errorf("mac3 should not be present")
		}
		foundMac3++
	}))
	p.Scan(context.Background())

	if foundMac != 2 {
		t.Errorf("Should have called two action funcs for mac but called %d", foundMac)
	}

	if foundMac2 != 1 {
		t.Errorf("Should have called one action func for mac2 but called %d", foundMac2)
	}

	if foundMac3 != 1 {
		t.Errorf("Should have called one action func for mac3 but called %d", foundMac2)
	}
}
