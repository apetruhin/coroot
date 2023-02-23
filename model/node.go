package model

import (
	"github.com/coroot/coroot/timeseries"
)

type DiskStats struct {
	IOUtilizationPercent *timeseries.TimeSeries
	ReadOps              *timeseries.TimeSeries
	WriteOps             *timeseries.TimeSeries
	WrittenBytes         *timeseries.TimeSeries
	ReadBytes            *timeseries.TimeSeries
	ReadTime             *timeseries.TimeSeries
	WriteTime            *timeseries.TimeSeries
	Wait                 *timeseries.TimeSeries
	Await                *timeseries.TimeSeries
}

type InterfaceStats struct {
	Name      string
	Addresses []string
	Up        *timeseries.TimeSeries
	RxBytes   *timeseries.TimeSeries
	TxBytes   *timeseries.TimeSeries
}

type NodePriceBreakdown struct {
	CPUPerCore    float32
	MemoryPerByte float32
	Costs
}

type Node struct {
	AgentVersion LabelLastValue

	Name      LabelLastValue
	MachineID string
	Uptime    *timeseries.TimeSeries

	CpuCapacity     *timeseries.TimeSeries
	CpuUsagePercent *timeseries.TimeSeries
	CpuUsageByMode  map[string]*timeseries.TimeSeries

	MemoryTotalBytes     *timeseries.TimeSeries
	MemoryFreeBytes      *timeseries.TimeSeries
	MemoryAvailableBytes *timeseries.TimeSeries
	MemoryCachedBytes    *timeseries.TimeSeries

	Disks         map[string]*DiskStats
	NetInterfaces []*InterfaceStats

	Instances []*Instance `json:"-"`

	CloudProvider     LabelLastValue
	Region            LabelLastValue
	AvailabilityZone  LabelLastValue
	InstanceType      LabelLastValue
	InstanceLifeCycle LabelLastValue

	PricePerHour float32
}

func NewNode(machineId string) *Node {
	return &Node{
		MachineID:      machineId,
		Disks:          map[string]*DiskStats{},
		CpuUsageByMode: map[string]*timeseries.TimeSeries{},
	}
}

func (node *Node) IsUp() bool {
	return !DataIsMissing(node.CpuUsagePercent)
}

func (node *Node) GetPriceBreakdown() *NodePriceBreakdown {
	if node.PricePerHour == 0 {
		return nil
	}
	cores := node.CpuCapacity.Last()
	ram := node.MemoryTotalBytes.Last()
	if timeseries.IsNaN(cores) || timeseries.IsNaN(ram) {
		return nil
	}
	ramGb := ram / (1000 * 1000 * 1000)
	perUnit := node.PricePerHour / (cores + ramGb) // assume that 1Gb of memory costs the same as 1 vCPU
	return &NodePriceBreakdown{
		CPUPerCore:    perUnit,
		MemoryPerByte: perUnit / (1000 * 1000 * 1000),
		Costs: Costs{
			CPUUsagePerHour:    perUnit * cores,
			MemoryUsagePerHour: perUnit * ramGb,
		},
	}
}
