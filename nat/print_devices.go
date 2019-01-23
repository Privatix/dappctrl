package nat

import "fmt"

// PrintAvailableDevices print available devices if found.
func PrintAvailableDevices() {
	config := NewConfig()
	upnp := discoverUPnP(config)
	if upnp != nil {
		fmt.Printf("found UPnP device: %+v\n", upnp)
	}
	pmp := discoverPMP()
	if pmp != nil {
		fmt.Printf("found NAT-PMP device: %+v\n", pmp)
	}
	if upnp == nil && pmp == nil {
		fmt.Println("no devices found")
	}
}
