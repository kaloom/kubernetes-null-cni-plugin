/*
Copyright (c) 2018 Kaloom Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

This is a "Multi-plugin" (a fork off Intel's Multus plugin) that
delegates work to other CNI plugins. The delegation's concept is
refered to from the CNI project; it reads other plugin netconf, and
then invoke them, e.g. flannel, knf or sriov plugin.
*/

package main

import (
	"fmt"
	"net"

	kc "github.com/kaloom/kubernetes-common"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/containernetworking/cni/pkg/version"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/containernetworking/plugins/pkg/utils/sysctl"
	"github.com/vishvananda/netlink"
)

const (
	nullIPv6AddressPrefixLen  = 64
	disableIPv6SysctlTemplate = "net.ipv6.conf.%s.disable_ipv6"
)

func delLinkLocalIPv6Addr(ifName string) (net.IP, error) {
	nl, err := netlink.LinkByName(ifName)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup link %q: %v", ifName, err)
	}

	// Make sure sysctl "disable_ipv6" is 0 if we are about to add
	// an IPv6 address to the interface

	// Enabled IPv6 for loopback "lo" and the interface
	// being configured
	for _, iface := range [2]string{"lo", ifName} {
		ipv6SysctlValueName := fmt.Sprintf(disableIPv6SysctlTemplate, iface)

		// Read current sysctl value
		value, err := sysctl.Sysctl(ipv6SysctlValueName)
		if err != nil || value == "0" {
			// FIXME: log warning if unable to read sysctl value
			continue
		}

		// Write sysctl to enable IPv6
		_, err = sysctl.Sysctl(ipv6SysctlValueName, "0")
		if err != nil {
			return nil, fmt.Errorf("failed to enable IPv6 for interface %q (%s=%s): %v", iface, ipv6SysctlValueName, value, err)
		}
	}

	err = netlink.LinkSetUp(nl)
	if err != nil {
		return nil, fmt.Errorf("failed to bring link %q up: %v", ifName, err)
	}

	addr, err := netlink.AddrList(nl, netlink.FAMILY_V6)
	if err != nil {
		return nil, fmt.Errorf("failed to get the list of ipv6 addresses on link %q", ifName)
	}
	for _, nla := range addr {
		if nla.IP.IsLinkLocalUnicast() {
			kc.LogDebug("resetLinkLocalIPv6Addr: will call netlink.AddrDel on %s for link %s", nla.IP.String(), ifName)
			if err := netlink.AddrDel(nl, &nla); err != nil {
				return nil, fmt.Errorf("failed to delete link local ipv6 address %q on intreface %q: %v",
					nla.IP.String(), ifName, err)
			}
			return nla.IP, nil
		}
	}
	return nil, fmt.Errorf("failed to find a link local ipv6 address on link %s", ifName)
}

func cmdAdd(args *skel.CmdArgs) error {
	var linkLocalIPv6Addr net.IP
	kc.LogDebug("cmdAdd: args: %v\n", string(args.StdinData[:]))

	versionDecoder := &version.ConfigDecoder{}
	confVersion, err := versionDecoder.Decode(args.StdinData)
	if err != nil {
		kc.LogError("cmdAdd: versionDecoder.Decode failed: %v\n", err)
		return err
	}
	netns, err := ns.GetNS(args.Netns)
	if err != nil {
		err = fmt.Errorf("failed to open netns %q: %v", args.Netns, err)
		kc.LogError("cmdAdd: %v\n", err)
		return err
	}
	defer netns.Close()

	// to avoid having the caller error because that the returned
	// link local ipv6 address already exist, we delete the
	// current one
	err = netns.Do(func(_ ns.NetNS) error {
		var err error
		linkLocalIPv6Addr, err = delLinkLocalIPv6Addr(args.IfName)
		if err != nil {
			kc.LogError("cmdAdd: delLinkLocalIPv6Addr failed: %v\n", err)
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	// as an IPAM it's expected to return in the result an ip address,
	// for the null plugin we return a link local ipv6 address
	nullInterfaceIP := current.IPConfig{
		Version: "6",
		Address: net.IPNet{
			IP:   linkLocalIPv6Addr,
			Mask: net.CIDRMask(nullIPv6AddressPrefixLen, 128),
		},
	}
	result := current.Result{
		CNIVersion: confVersion,
		IPs: []*current.IPConfig{
			&nullInterfaceIP,
		},
	}
	kc.LogInfo("cmdAdd: returned ip address %s for sandbox %s\n", linkLocalIPv6Addr, netns.Path())

	return result.Print()
}

func cmdDel(args *skel.CmdArgs) error {
	kc.LogDebug("cmdDel: args: %v\n", string(args.StdinData[:]))
	kc.LogInfo("cmdDel: nothing to do for IfName %s\n", args.IfName)
	return nil
}

func main() {
	logParams := kc.LoggingParams{
		Prefix: "NULL ",
	}
	// will get a file object if _CNI_LOGGING_LEVEL environment variable is
	// set to a value >= 1, otherwise no logging goes to /dev/null
	lf := kc.OpenLogFile(&logParams)
	defer kc.CloseLogFile(lf)

	// TODO: implement plugin version
	skel.PluginMain(cmdAdd, cmdGet, cmdDel, version.All, "TODO")
}

func cmdGet(args *skel.CmdArgs) error {
	// TODO: implement
	return fmt.Errorf("not implemented")
}
