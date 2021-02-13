package presence

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"
)

//Action Status function is called with the processed MAC address and info whether the device was found or not
type Action interface {
	Status(mac string, found bool)
}

//ActionFunc is called with the processed MAC address and info whether the device was found or not
type ActionFunc func(mac string, found bool)

//Status is called with the status of the given MAC address
func (a ActionFunc) Status(mac string, found bool) {
	a(mac, found)
}

//Presence allows scanning for nearby bluetooth devices based on their MAC address
type Presence struct {
	btScanner BluetoothScanner
	actions   map[string][]Action
}

//BluetoothScanner provides required methods to scan for bluetooth devices
type BluetoothScanner interface {
	IsPresent(context context.Context, mac string) (bool, error)
}

//Ensure HCITool implements the BluetoothScanner interface
var _ BluetoothScanner = (*HCITool)(nil)

//HCITool is the default BluetoothScanner implementation
type HCITool struct {
	path string
}

//NewHCITool returns a new instance of HCITool
func NewHCITool() (*HCITool, error) {
	hci, err := exec.LookPath("hcitool")
	if err != nil {
		return nil, err
	}

	return &HCITool{
		path: hci,
	}, nil
}

//IsPresent checks if bluetooth device with given MAC address can be found
func (h *HCITool) IsPresent(ctx context.Context, mac string) (bool, error) {
	cmd := exec.CommandContext(ctx, h.path, "name", mac)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("Error while scannng for device %s, %w", mac, err)
	}
	if len(output) <= 0 {
		return false, nil
	}
	return true, nil
}

//New returns new Presence instance
func New(btScanner BluetoothScanner) *Presence {
	p := &Presence{
		actions:   make(map[string][]Action),
		btScanner: btScanner,
	}
	return p
}

//RegisterAction registers one ore more actions for a certain MAC address.
//
//Note: Given actions are added to already existing registered actions for given MAC address.
func (p *Presence) RegisterAction(mac string, actions ...Action) {
	for _, a := range actions {
		p.actions[mac] = append(p.actions[mac], a)
	}
}

//UnregisterActions removes all registered actions for a given MAC address
//
//If no actions are registered, this method is a no-op.
func (p *Presence) UnregisterActions(mac string) {
	delete(p.actions, mac)
}

//Scan executes a single presence check for all registered MAC addresses and calls
//action functions with the scan result.
func (p *Presence) Scan(ctx context.Context) {
	var wg sync.WaitGroup
	done := make(chan struct{})
	go func() {
		for mac, actions := range p.actions {
			wg.Add(1)
			go func(mac string, actions []Action) {
				defer wg.Done()
				isPresent, err := p.btScanner.IsPresent(ctx, mac)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Could not scan successfully")
					isPresent = false
				}
				for _, a := range actions {
					a.Status(mac, isPresent)
				}
			}(mac, actions)
		}
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
	}
}

//ScanPeriodically will call Scan periodically with a given interval after
//a run cycle has finished
func (p *Presence) ScanPeriodically(ctx context.Context, interval time.Duration) {
	//Run once immediately
	p.Scan(ctx)

	//Run periodically afterwards
	for {
		select {
		case <-time.Tick(interval):
			p.Scan(ctx)
		case <-ctx.Done():
			return
		}
	}
}
