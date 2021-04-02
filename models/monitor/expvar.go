package monitor

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
)

type Monitor struct {
	MonitorService interface{}
	MonitorSystem  interface{}
}

func GetMonitor(pretty ...string) (m Monitor) {
	return Monitor{
		MonitorService: GetMonitorService(pretty...),
		MonitorSystem:  GetMonitorSystem(pretty...),
	}
}

var (
	start = time.Now()
)

type MonitorServicePretty struct {
	DateTime   time.Time
	Cmdline    []string
	RunTime    string
	GoVersion  string
	CPUs       int
	OS         string
	Goroutines int
	Cgos       int64
	Memstats   MemStatsPretty
}

type MemStatsPretty struct {
	Sys          string
	Alloc        string
	HeapSys      string
	HeapAlloc    string
	HeapIdle     string
	HeapInuse    string
	HeapReleased string
	StackSys     string
	StackInuse   string
	GCSys        string
	OtherSys     string
	NextGC       string
}

type MonitorService struct {
	DateTime   time.Time
	Cmdline    []string
	RunTime    float64
	GoVersion  string
	CPUs       int
	OS         string
	Goroutines int
	Cgos       int64
	Memstats   MemStats
}

type MemStats struct {
	Sys          uint64
	Alloc        uint64
	HeapSys      uint64
	HeapAlloc    uint64
	HeapIdle     uint64
	HeapInuse    uint64
	HeapReleased uint64
	StackSys     uint64
	StackInuse   uint64
	GCSys        uint64
	OtherSys     uint64
	NextGC       uint64
}

// calculateUptime 计算运行时间
func calculateUptimePretty() string {
	return time.Since(start).String()
}

// calculateUptime 计算运行时间
func calculateUptime() float64 {
	return time.Since(start).Seconds()
}

// currentGoVersion 当前 Golang 版本
func currentGoVersion() string {
	return runtime.Version()
}

// getNumCPUs 获取 CPU 核心数量
func getNumCPUs() int {
	return runtime.NumCPU()
}

// getGoOS 当前系统类型
func getGoOS() string {
	return runtime.GOOS
}

// getNumGoroutins 当前 goroutine 数量
func getNumGoroutins() int {
	return runtime.NumGoroutine()
}

// getNumCgoCall CGo 调用次数
func getNumCgoCall() int64 {
	return runtime.NumCgoCall()
}

// memstats
func memstats(t ...string) interface{} {
	stats := new(runtime.MemStats)
	runtime.ReadMemStats(stats)

	if len(t) > 0 && t[0] == "pretty" {
		return MemStatsPretty{
			Sys:          formatSize(stats.Sys),
			Alloc:        formatSize(stats.Alloc),
			HeapSys:      formatSize(stats.HeapSys),
			HeapAlloc:    formatSize(stats.HeapAlloc),
			HeapIdle:     formatSize(stats.HeapIdle),
			HeapInuse:    formatSize(stats.HeapInuse),
			HeapReleased: formatSize(stats.HeapReleased),
			StackSys:     formatSize(stats.StackSys),
			StackInuse:   formatSize(stats.StackInuse),
			GCSys:        formatSize(stats.GCSys),
			OtherSys:     formatSize(stats.OtherSys),
			NextGC:       formatSize(stats.NextGC),
		}

	}

	return MemStats{
		Sys:          stats.Sys,
		Alloc:        stats.Alloc,
		HeapSys:      stats.HeapSys,
		HeapAlloc:    stats.HeapAlloc,
		HeapIdle:     stats.HeapIdle,
		HeapInuse:    stats.HeapInuse,
		HeapReleased: stats.HeapReleased,
		StackSys:     stats.StackSys,
		StackInuse:   stats.StackInuse,
		GCSys:        stats.GCSys,
		OtherSys:     stats.OtherSys,
		NextGC:       stats.NextGC,
	}

}

func cmdline() []string {
	return os.Args
}

func GetMonitorService(t ...string) interface{} {
	if len(t) > 0 && t[0] == "pretty" {
		return MonitorServicePretty{
			DateTime:   time.Now(),
			OS:         getGoOS(),
			GoVersion:  currentGoVersion(),
			Cmdline:    cmdline(),
			Memstats:   memstats(t...).(MemStatsPretty),
			RunTime:    calculateUptimePretty(),
			CPUs:       getNumCPUs(),
			Goroutines: getNumGoroutins(),
			Cgos:       getNumCgoCall(),
		}
	}

	return MonitorService{
		DateTime:   time.Now(),
		OS:         getGoOS(),
		GoVersion:  currentGoVersion(),
		Cmdline:    cmdline(),
		Memstats:   memstats(t...).(MemStats),
		RunTime:    calculateUptime(),
		CPUs:       getNumCPUs(),
		Goroutines: getNumGoroutins(),
		Cgos:       getNumCgoCall(),
	}
}

func GetMonitorForProm() []byte {
	stats := new(runtime.MemStats)
	runtime.ReadMemStats(stats)

	buf := new(bytes.Buffer)
	buf.WriteString("RunTime " + parseFloat64(calculateUptime()) + "\n")
	buf.WriteString("CPUs " + parseInt(getNumCPUs()) + "\n")
	buf.WriteString("Goroutines " + parseInt(getNumGoroutins()) + "\n")
	buf.WriteString("Cgos " + parseInt64(getNumCgoCall()) + "\n")
	buf.WriteString("Sys " + parseUint64(stats.Sys) + "\n")
	buf.WriteString("Alloc " + parseUint64(stats.Alloc) + "\n")
	buf.WriteString("HeapSys " + parseUint64(stats.HeapSys) + "\n")
	buf.WriteString("HeapAlloc " + parseUint64(stats.HeapAlloc) + "\n")
	buf.WriteString("HeapIdle " + parseUint64(stats.HeapIdle) + "\n")
	buf.WriteString("HeapInuse " + parseUint64(stats.HeapInuse) + "\n")
	buf.WriteString("HeapReleased " + parseUint64(stats.HeapReleased) + "\n")
	buf.WriteString("StackSys " + parseUint64(stats.StackSys) + "\n")
	buf.WriteString("StackInuse " + parseUint64(stats.StackInuse) + "\n")
	buf.WriteString("GCSys " + parseUint64(stats.GCSys) + "\n")
	buf.WriteString("OtherSys " + parseUint64(stats.OtherSys) + "\n")
	buf.WriteString("NextGC " + parseUint64(stats.NextGC) + "\n")
	return buf.Bytes()
}

type MonitorSystemPretty struct {
	DateTime    time.Time
	DiskStats   []DiskStatsPretty
	SysMemStats SysMemStatsPretty
	CpuStats    CpuStats
}
type MonitorSystem struct {
	DateTime    time.Time
	DiskStats   []DiskStats
	SysMemStats SysMemStats
	CpuStats    CpuStats
}

type DiskStatsPretty struct {
	Path string
	All  string
	Used string
	Free string
}
type DiskStats struct {
	Path string
	All  uint64
	Used uint64
	Free uint64
}

type SysMemStatsPretty struct {
	All   string
	Used  string
	Avail string
	Free  string
}
type SysMemStats struct {
	All   uint64
	Used  uint64
	Avail uint64
	Free  uint64
}

type CpuStats struct {
	Count   int
	Percent float64
}

func diskUsagePretty(paths []string) (dss []DiskStatsPretty) {
	dss = []DiskStatsPretty{}
	for _, path := range paths {
		di, err := disk.Usage(path)
		if err != nil {
			continue
		}

		ds := DiskStatsPretty{
			Path: di.Path,
			All:  formatSize(di.Total),
			Used: formatSize(di.Used),
			Free: formatSize(di.Free),
		}
		dss = append(dss, ds)

	}
	return
}
func diskUsage(paths []string) (dss []DiskStats) {
	dss = []DiskStats{}
	for _, path := range paths {
		di, err := disk.Usage(path)
		if err != nil {
			continue
		}

		ds := DiskStats{
			Path: di.Path,
			All:  di.Total,
			Used: di.Used,
			Free: di.Free,
		}
		dss = append(dss, ds)

	}
	return
}

func sysMemUsage(pretty ...string) interface{} {
	v, err := mem.VirtualMemory()
	if err != nil {
		return nil
	}
	if len(pretty) > 0 && pretty[0] == "pretty" {
		sm := SysMemStatsPretty{}
		sm.All = formatSize(v.Total)
		sm.Avail = formatSize(v.Available)
		sm.Used = formatSize(v.Used)
		sm.Free = formatSize(v.Free)
		return sm
	}
	sm := SysMemStats{}
	sm.All = v.Total
	sm.Avail = v.Available
	sm.Used = v.Used
	sm.Free = v.Free
	return sm
}

func cpuUsage() (cs CpuStats) {
	cs = CpuStats{}
	cnt, _ := cpu.Counts(true)
	cs.Count = cnt
	percent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return
	}
	// cs.Percent, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", percent[0]), 64)
	cs.Percent = percent[0]
	return
}

func GetMonitorSystem(t ...string) interface{} {
	parts, err := disk.Partitions(true)
	if err != nil {
		return MonitorSystem{time.Now(), []DiskStats{}, SysMemStats{}, CpuStats{}}
	}
	paths := make([]string, len(parts))
	for _, part := range parts {
		paths = append(paths, part.Mountpoint)
	}

	if len(t) > 0 && t[0] == "pretty" {
		return MonitorSystemPretty{
			DateTime:    time.Now(),
			DiskStats:   diskUsagePretty(paths),
			SysMemStats: sysMemUsage(t...).(SysMemStatsPretty),
			CpuStats:    cpuUsage(),
		}
	}

	return MonitorSystem{
		DateTime:    time.Now(),
		DiskStats:   diskUsage(paths),
		SysMemStats: sysMemUsage(t...).(SysMemStats),
		CpuStats:    cpuUsage(),
	}
}
func formatSize(size uint64) string {
	if size > 0 && size < 1024 {
		return fmt.Sprintf("%.2fB", float64(size)/float64(1))
	} else if size < (1024 * 1024) {
		return fmt.Sprintf("%.2fKB", float64(size)/float64(1024))
	} else if size < (1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fMB", float64(size)/float64(1024*1024))
	} else if size < (1024 * 1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fGB", float64(size)/float64(1024*1024*1024))
	} else if size < (1024 * 1024 * 1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fTB", float64(size)/float64(1024*1024*1024*1024))
	} else {
		return fmt.Sprintf("%.2fEB", float64(size)/float64(1024*1024*1024*1024*1024))
	}
}

func parseFloat64(v float64) string {
	return strconv.FormatFloat(v, 'E', -1, 64)
}
func parseInt(v int) string {
	return strconv.FormatInt(int64(v), 10)
}
func parseInt64(v int64) string {
	return strconv.FormatInt(v, 10)
}
func parseUint64(v uint64) string {
	return strconv.FormatUint(v, 10)
}
