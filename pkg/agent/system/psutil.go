package system

import (
	"github.com/sentrycloud/sentry/pkg/agent/reporter"
	"github.com/sentrycloud/sentry/pkg/newlog"
	"github.com/sentrycloud/sentry/pkg/protocol"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
	"strings"
	"time"
)

const CollectInterval = 10

var (
	prevReadWriteCounterMap = make(map[string]uint64)
	prevReadWriteTimeMap    = make(map[string]uint64)
	prevTotalIOTimeMap      = make(map[string]uint64)

	prevIOBytesRead, prevIOBytesWrite uint64

	prevPacketsSent, prevBytesSent, prevPacketsRecv, prevBytesRecv uint64
)

func CollectSystemMetric(r *reporter.Reporter) {
	for {
		var systemMetrics []protocol.MetricValue
		startTime := time.Now()

		ts := uint64(startTime.Unix() / CollectInterval * CollectInterval)

		collectCpuPercent(&systemMetrics, ts)
		collectLoadAverage(&systemMetrics, ts)
		collectMemoryUsage(&systemMetrics, ts)
		collectDiskUsage(&systemMetrics, ts)
		collectDiskIOStats(&systemMetrics, ts)
		collectNetIOStats(&systemMetrics, ts)
		collectTCPStatus(&systemMetrics, ts)
		collectProcessNumber(&systemMetrics, ts)

		if r != nil {
			r.Report(systemMetrics)
		}

		collectTime := time.Now().Sub(startTime)
		sleepTime := CollectInterval*time.Second - collectTime
		time.Sleep(sleepTime)
	}
}

func collectCpuPercent(systemMetrics *[]protocol.MetricValue, ts uint64) {
	cpuPercent, err := cpu.Percent(0*time.Second, false)
	if err == nil {
		addMetric(systemMetrics, "sentry_sys_cpu_usage", nil, ts, cpuPercent[0])
	}
}

func collectLoadAverage(systemMetrics *[]protocol.MetricValue, ts uint64) {
	loadAvg, err := load.Avg()
	if err == nil {
		addMetric(systemMetrics, "sentry_sys_load_average", nil, ts, loadAvg.Load1)
	}
}

func collectMemoryUsage(systemMetrics *[]protocol.MetricValue, ts uint64) {
	memStat, err := mem.VirtualMemory()
	if err == nil {
		addMetric(systemMetrics, "sentry_sys_mem_usage", nil, ts, memStat.UsedPercent)
	}
}

func collectDiskUsage(systemMetrics *[]protocol.MetricValue, ts uint64) {
	diskStats, err := disk.Partitions(true)
	if err == nil {
		for _, diskStat := range diskStats {
			if strings.HasPrefix(diskStat.Device, "/dev") {
				diskUsage, e := disk.Usage(diskStat.Mountpoint)
				if e == nil {
					tags := map[string]string{"device": diskStat.Device}
					addMetric(systemMetrics, "sentry_sys_disk_usage", tags, ts, diskUsage.UsedPercent)
				}
			}
		}
	}
}

func collectDiskIOStats(systemMetrics *[]protocol.MetricValue, ts uint64) {
	diskIoCounters, err := disk.IOCounters()
	if err != nil {
		newlog.Error("get disk io counters failed: %v", err)
		return
	}

	// report only max ioWait and ioUtil
	maxIOWait := 0.0
	maxIOUtil := 0.0
	var currentBytesRead uint64 = 0
	var currentBytesWrite uint64 = 0
	for name, ioCounters := range diskIoCounters {
		currentBytesRead += ioCounters.ReadBytes
		currentBytesWrite += ioCounters.WriteBytes
		currentIOCount := ioCounters.ReadCount + ioCounters.WriteCount
		currentIOTime := ioCounters.ReadTime + ioCounters.WriteTime
		if _, exist := prevReadWriteCounterMap[name]; exist {
			ioCount := currentIOCount - prevReadWriteCounterMap[name]
			ioTime := currentIOTime - prevReadWriteTimeMap[name]
			ioTotalTime := ioCounters.IoTime - prevTotalIOTimeMap[name]

			ioWait := 0.0
			if ioCount != 0 {
				ioWait = float64(ioTime) / float64(ioCount)
			}
			ioUtilPercent := 100 * float64(ioTotalTime) / float64(CollectInterval*1000)

			if maxIOWait < ioWait {
				maxIOWait = ioWait
			}

			if maxIOUtil < ioUtilPercent {
				maxIOUtil = ioUtilPercent
			}
		}

		prevReadWriteCounterMap[name] = currentIOCount
		prevReadWriteTimeMap[name] = currentIOTime
		prevTotalIOTimeMap[name] = ioCounters.IoTime
	}

	if prevIOBytesRead != 0 {
		bytesReadPerSecond := (currentBytesRead - prevIOBytesRead) / CollectInterval
		bytesWritePerSecond := (currentBytesWrite - prevIOBytesWrite) / CollectInterval

		addMetric(systemMetrics, "sentry_sys_io_read", nil, ts, float64(bytesReadPerSecond))
		addMetric(systemMetrics, "sentry_sys_io_write", nil, ts, float64(bytesWritePerSecond))

		addMetric(systemMetrics, "sentry_sys_io_wait", nil, ts, maxIOWait)
		addMetric(systemMetrics, "sentry_sys_io_util", nil, ts, maxIOUtil)
	}

	prevIOBytesRead = currentBytesRead
	prevIOBytesWrite = currentBytesWrite
}

func collectNetIOStats(systemMetrics *[]protocol.MetricValue, ts uint64) {
	netIoCounters, err := net.IOCounters(false)
	if err != nil {
		newlog.Error("get net io counters failed: %v", err)
		return
	}

	if prevBytesSent != 0 {
		bytesSentPerSecond := float64(netIoCounters[0].BytesSent-prevBytesSent) / CollectInterval
		packetsSentPerSecond := float64(netIoCounters[0].PacketsSent-prevPacketsSent) / CollectInterval
		bytesRecvPerSecond := float64(netIoCounters[0].BytesRecv-prevBytesRecv) / CollectInterval
		packetsRecvPerSecond := float64(netIoCounters[0].PacketsRecv-prevPacketsRecv) / CollectInterval

		addMetric(systemMetrics, "sentry_sys_net_bytes_sent", nil, ts, bytesSentPerSecond)
		addMetric(systemMetrics, "sentry_sys_net_packets_sent", nil, ts, packetsSentPerSecond)
		addMetric(systemMetrics, "sentry_sys_net_bytes_recv", nil, ts, bytesRecvPerSecond)
		addMetric(systemMetrics, "sentry_sys_net_packets_recv", nil, ts, packetsRecvPerSecond)
	}

	prevBytesSent = netIoCounters[0].BytesSent
	prevPacketsSent = netIoCounters[0].PacketsSent
	prevBytesRecv = netIoCounters[0].BytesRecv
	prevPacketsRecv = netIoCounters[0].PacketsRecv
}

func collectTCPStatus(systemMetrics *[]protocol.MetricValue, ts uint64) {
	tcpConnections, err := net.Connections("tcp")
	if err != nil {
		newlog.Error("get net connections failed: %v", err)
		return
	}

	tcpStatus := make(map[string]int)
	for _, tcpConn := range tcpConnections {
		tcpStatus[tcpConn.Status] = tcpStatus[tcpConn.Status] + 1
	}

	for status, count := range tcpStatus {
		tags := map[string]string{"status": status}
		addMetric(systemMetrics, "sentry_sys_tcp_status", tags, ts, float64(count))
	}
}

func collectProcessNumber(systemMetrics *[]protocol.MetricValue, ts uint64) {
	pids, err := process.Pids()
	if err == nil {
		addMetric(systemMetrics, "sentry_sys_process_number", nil, ts, float64(len(pids)))
	}
}

func addMetric(systemMetrics *[]protocol.MetricValue, metricName string, tags map[string]string, ts uint64, value float64) {
	metric := protocol.MetricValue{
		Metric:    metricName,
		Tags:      tags,
		Timestamp: ts,
		Value:     value,
	}

	*systemMetrics = append(*systemMetrics, metric)
}
