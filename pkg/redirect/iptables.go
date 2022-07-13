package redirect

import (
	"fmt"
	"net"
	"sync"

	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/util/iptables"
	"k8s.io/utils/exec"
)

//
// 1. create a chain for k8s svc (in our case we want to have postrouting): iptables -N LIGHTPATH-SVC-REDIRECT
// 2. Before routing, redirect: -A PREROUTING -m comment --comment "lightpath prerouting rules" -j LIGHTPATH-SVC-REDIRECT
// [3. For each service, add a REDIRECT target to envoy: -A LIGHTPATH-SVC-REDIRECT -s <src-ip> -p tcp -j REDIRECT --to-port <envoy proxy port; 1666>] -m comment --comment "lightpath:<namespace>/<svc-name>"
//
// Step 3 could likely change -> lightpath needs to maintain the current and desired state; only to apply the diff
//

const (
	iptablesLightpathChainName = "LIGHTPATH-SVC-REDIRECT"
)

type IptablesRedirect struct {
	m sync.Mutex

	iptables  iptables.Interface // IPv4
	ip6tables iptables.Interface // IPv6
	envoyPort int
}

func NewIptablesRedirect(envoyPort int) *IptablesRedirect {
	return &IptablesRedirect{
		iptables:  iptables.New(exec.New(), iptables.ProtocolIPv4),
		ip6tables: iptables.New(exec.New(), iptables.ProtocolIPv6),
		envoyPort: envoyPort,
	}
}

func (ir *IptablesRedirect) Prereqs() error {
	var err error

	_, err = ir.iptables.EnsureChain(iptables.TableNAT, iptablesLightpathChainName)
	if err != nil {
		klog.ErrorS(err, "Failed to ensure iptables chain (IPv4)")
		return err
	}

	_, err = ir.ip6tables.EnsureChain(iptables.TableNAT, iptablesLightpathChainName)
	if err != nil {
		klog.ErrorS(err, "Failed to ensure iptables chain (IPv6)")
		return err
	}

	return nil
}

func (ir *IptablesRedirect) AddIP(comment string, ip net.IP) error {
	ir.m.Lock()
	defer ir.m.Unlock()

	redirectArgs := []string{"-m", "comment", "--comment", comment, "-s", ip.String(), "-j", "REDIRECT", "--to-port", fmt.Sprint(ir.envoyPort)}

	var err error

	if ip.To4() != nil {
		// IPv4
		_, err = ir.iptables.EnsureRule(iptables.Append, iptables.TableNAT, iptablesLightpathChainName, redirectArgs...)
	} else {
		// IPv6
		_, err = ir.ip6tables.EnsureRule(iptables.Append, iptables.TableNAT, iptablesLightpathChainName, redirectArgs...)
	}

	return err
}

func (ir *IptablesRedirect) RemoveIP(comment string, ip net.IP) error {
	ir.m.Lock()
	defer ir.m.Unlock()

	redirectArgs := []string{"-m", "comment", "--comment", comment, "-s", ip.String(), "-j", "REDIRECT", "--to-port", fmt.Sprint(ir.envoyPort)}

	var err error

	if ip.To4() != nil {
		// IPv4
		err = ir.iptables.DeleteRule(iptables.TableNAT, iptablesLightpathChainName, redirectArgs...)
	} else {
		// IPv6
		err = ir.ip6tables.DeleteRule(iptables.TableNAT, iptablesLightpathChainName, redirectArgs...)
	}

	// Catch not found error
	if iptables.IsNotFoundError(err) {
		return nil
	}

	return err
}

func (ir *IptablesRedirect) Cleanup() error {
	ir.m.Lock()
	defer ir.m.Unlock()

	preroutingArgs := []string{"-m", "comment", "--comment", "lightpath prerouting rules", "-j", iptablesLightpathChainName}

	// We want to catch all errors but continue the cleanup
	var lastErr error = nil

	// Delete prerouting rules
	if err := ir.iptables.DeleteRule(iptables.TableNAT, iptables.ChainPrerouting, preroutingArgs...); err != nil && iptables.IsNotFoundError(err) {
		klog.ErrorS(err, "Failed to delete jump to chain rule (IPv4)")
		lastErr = err
	}
	if err := ir.ip6tables.DeleteRule(iptables.TableNAT, iptables.ChainPrerouting, preroutingArgs...); err != nil && iptables.IsNotFoundError(err) {
		klog.ErrorS(err, "Failed to delete jump to chain rule (IPv6)")
		lastErr = err
	}

	// Delete chain
	if err := ir.iptables.DeleteChain(iptables.TableNAT, iptablesLightpathChainName); err != nil && iptables.IsNotFoundError(err) {
		klog.ErrorS(err, "Failed to delete jump to chain rule (IPv4)")
		lastErr = err
	}
	if err := ir.ip6tables.DeleteChain(iptables.TableNAT, iptablesLightpathChainName); err != nil && iptables.IsNotFoundError(err) {
		klog.ErrorS(err, "Failed to delete jump to chain rule (IPv6")
		lastErr = err
	}

	return lastErr
}
