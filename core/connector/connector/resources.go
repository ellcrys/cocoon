package connector

import "github.com/ellcrys/cocoon/core/connector/monitor"

// updateResourceUsage updates the resource usage data of the running cocoon code
func (cn *Connector) updateResourceUsage(report monitor.Report) {
	cn.resourceUsage.NetRx = cn.resourceUsage.NetRx + (report.NetRx - cn.resourceUsage.NetRx)
	cn.resourceUsage.NetTx = cn.resourceUsage.NetTx + (report.NetTx - cn.resourceUsage.NetTx)
}
