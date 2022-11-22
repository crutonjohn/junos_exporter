package main

import (
	"github.com/czerwonk/junos_exporter/internal/config"
	"github.com/czerwonk/junos_exporter/pkg/collector"
	"github.com/czerwonk/junos_exporter/pkg/connector"
	"github.com/czerwonk/junos_exporter/pkg/interfacelabels"
	"github.com/czerwonk/junos_exporter/pkg/modules/accounting"
	"github.com/czerwonk/junos_exporter/pkg/modules/alarm"
	"github.com/czerwonk/junos_exporter/pkg/modules/bfd"
	"github.com/czerwonk/junos_exporter/pkg/modules/bgp"
	"github.com/czerwonk/junos_exporter/pkg/modules/environment"
	"github.com/czerwonk/junos_exporter/pkg/modules/firewall"
	"github.com/czerwonk/junos_exporter/pkg/modules/fpc"
	"github.com/czerwonk/junos_exporter/pkg/modules/interfacediagnostics"
	"github.com/czerwonk/junos_exporter/pkg/modules/interfacequeue"
	"github.com/czerwonk/junos_exporter/pkg/modules/interfaces"
	"github.com/czerwonk/junos_exporter/pkg/modules/ipsec"
	"github.com/czerwonk/junos_exporter/pkg/modules/isis"
	"github.com/czerwonk/junos_exporter/pkg/modules/l2circuit"
	"github.com/czerwonk/junos_exporter/pkg/modules/lacp"
	"github.com/czerwonk/junos_exporter/pkg/modules/ldp"
	"github.com/czerwonk/junos_exporter/pkg/modules/mac"
	"github.com/czerwonk/junos_exporter/pkg/modules/mplslsp"
	"github.com/czerwonk/junos_exporter/pkg/modules/nat"
	"github.com/czerwonk/junos_exporter/pkg/modules/nat2"
	"github.com/czerwonk/junos_exporter/pkg/modules/ospf"
	"github.com/czerwonk/junos_exporter/pkg/modules/power"
	"github.com/czerwonk/junos_exporter/pkg/modules/route"
	"github.com/czerwonk/junos_exporter/pkg/modules/routingengine"
	"github.com/czerwonk/junos_exporter/pkg/modules/rpki"
	"github.com/czerwonk/junos_exporter/pkg/modules/rpm"
	"github.com/czerwonk/junos_exporter/pkg/modules/security"
	"github.com/czerwonk/junos_exporter/pkg/modules/storage"
	"github.com/czerwonk/junos_exporter/pkg/modules/system"
	"github.com/czerwonk/junos_exporter/pkg/modules/vpws"
	"github.com/czerwonk/junos_exporter/pkg/modules/vrrp"
)

type collectors struct {
	logicalSystem string
	dynamicLabels *interfacelabels.DynamicLabels
	collectors    map[string]collector.RPCCollector
	devices       map[string][]collector.RPCCollector
	cfg           *config.Config
}

func collectorsForDevices(devices []*connector.Device, cfg *config.Config, logicalSystem string, dynamicLabels *interfacelabels.DynamicLabels) *collectors {
	c := &collectors{
		logicalSystem: logicalSystem,
		dynamicLabels: dynamicLabels,
		collectors:    make(map[string]collector.RPCCollector),
		devices:       make(map[string][]collector.RPCCollector),
		cfg:           cfg,
	}

	for _, d := range devices {
		c.initCollectorsForDevices(d)
	}

	return c
}

func (c *collectors) initCollectorsForDevices(device *connector.Device) {
	f := c.cfg.FeaturesForDevice(device.Host)

	c.devices[device.Host] = make([]collector.RPCCollector, 0)

	c.addCollectorIfEnabledForDevice(device, "routingengine", f.RoutingEngine, routingengine.NewCollector)
	c.addCollectorIfEnabledForDevice(device, "accounting", f.Accounting, accounting.NewCollector)
	c.addCollectorIfEnabledForDevice(device, "alarm", f.Alarm, func() collector.RPCCollector {
		return alarm.NewCollector(*alarmFilter)
	})
	c.addCollectorIfEnabledForDevice(device, "bfd", f.BFD, bfd.NewCollector)
	c.addCollectorIfEnabledForDevice(device, "bgp", f.BGP, func() collector.RPCCollector {
		return bgp.NewCollector(c.logicalSystem)
	})
	c.addCollectorIfEnabledForDevice(device, "env", f.Environment, environment.NewCollector)
	c.addCollectorIfEnabledForDevice(device, "firewall", f.Firewall, firewall.NewCollector)
	c.addCollectorIfEnabledForDevice(device, "fpc", f.FPC, fpc.NewCollector)
	c.addCollectorIfEnabledForDevice(device, "ifacediag", f.InterfaceDiagnostic, func() collector.RPCCollector {
		return interfacediagnostics.NewCollector(c.dynamicLabels)
	})
	c.addCollectorIfEnabledForDevice(device, "ifacequeue", f.InterfaceQueue, func() collector.RPCCollector {
		return interfacequeue.NewCollector(c.dynamicLabels)
	})
	c.addCollectorIfEnabledForDevice(device, "iface", f.Interfaces, func() collector.RPCCollector {
		return interfaces.NewCollector(c.dynamicLabels)
	})
	c.addCollectorIfEnabledForDevice(device, "ipsec", f.IPSec, ipsec.NewCollector)
	c.addCollectorIfEnabledForDevice(device, "isis", f.ISIS, isis.NewCollector)
	c.addCollectorIfEnabledForDevice(device, "l2c", f.L2Circuit, l2circuit.NewCollector)
	c.addCollectorIfEnabledForDevice(device, "lacp", f.LACP, lacp.NewCollector)
	c.addCollectorIfEnabledForDevice(device, "ldp", f.LDP, ldp.NewCollector)
	c.addCollectorIfEnabledForDevice(device, "nat", f.NAT, nat.NewCollector)
	c.addCollectorIfEnabledForDevice(device, "nat2", f.NAT2, nat2.NewCollector)
	c.addCollectorIfEnabledForDevice(device, "ospf", f.OSPF, func() collector.RPCCollector {
		return ospf.NewCollector(c.logicalSystem)
	})
	c.addCollectorIfEnabledForDevice(device, "routes", f.Routes, route.NewCollector)
	c.addCollectorIfEnabledForDevice(device, "rpki", f.RPKI, rpki.NewCollector)
	c.addCollectorIfEnabledForDevice(device, "rpm", f.RPM, rpm.NewCollector)
	c.addCollectorIfEnabledForDevice(device, "security", f.Security, security.NewCollector)
	c.addCollectorIfEnabledForDevice(device, "storage", f.Storage, storage.NewCollector)
	c.addCollectorIfEnabledForDevice(device, "system", f.System, system.NewCollector)
	c.addCollectorIfEnabledForDevice(device, "power", f.Power, power.NewCollector)
	c.addCollectorIfEnabledForDevice(device, "mac", f.MAC, mac.NewCollector)
	c.addCollectorIfEnabledForDevice(device, "vrrp", f.VRRP, vrrp.NewCollector)
	c.addCollectorIfEnabledForDevice(device, "vpws", f.VPWS, vpws.NewCollector)
	c.addCollectorIfEnabledForDevice(device, "mpls_lsp", f.MPLS_LSP, mplslsp.NewCollector)
}

func (c *collectors) addCollectorIfEnabledForDevice(device *connector.Device, key string, enabled bool, newCollector func() collector.RPCCollector) {
	if !enabled {
		return
	}

	col, found := c.collectors[key]
	if !found {
		col = newCollector()
		c.collectors[key] = col
	}

	c.devices[device.Host] = append(c.devices[device.Host], col)
}

func (c *collectors) allEnabledCollectors() []collector.RPCCollector {
	collectors := make([]collector.RPCCollector, len(c.collectors))

	i := 0
	for _, collector := range c.collectors {
		collectors[i] = collector
		i++
	}

	return collectors
}

func (c *collectors) collectorsForDevice(device *connector.Device) []collector.RPCCollector {
	cols, found := c.devices[device.Host]
	if !found {
		return []collector.RPCCollector{}
	}

	return cols
}
