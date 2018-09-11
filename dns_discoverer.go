package veneur

import (
	"fmt"
	"net"
)

type DnsDiscoverer struct {
	Port int
}

func NewDnsDiscoverer(port int) (*DnsDiscoverer, error) {
	return &DnsDiscoverer{
		Port: port,
	}, nil
}

func (d *DnsDiscoverer) GetDestinationsForService(serviceName string) ([]string, error) {
	ipAddresses, err := net.LookupIP(serviceName)
	if err != nil {
		return nil, err
	}
	numOfIps := len(ipAddresses)
	hostAddresses := make([]string, numOfIps)
	for index, ip := range ipAddresses {
		hostAddresses[index] = fmt.Sprintf("%s:%d", ip.String(), d.Port)
	}

	return hostAddresses, nil
}
