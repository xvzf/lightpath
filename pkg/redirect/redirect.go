package redirect

import "net"

type Redirect interface {
	Prereqs() error
	AddIP(comment string, ip net.IP) error
	RemoveIP(comment string, ip net.IP) error
	Cleanup() error
}
