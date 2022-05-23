package internal

import (
	"fmt"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type Environment struct {
	Vendor    string `json:"vendor"`
	Name      string `json:"name"`
	Perf      string `json:"perf"`
	Processor string `json:"processor"`
	Graphics  string `json:"graphics,omitempty"`
	Platform  string `json:"platform"`
	Kernel    string `json:"kernel"`
	Init      string `json:"init,omitempty"`
}

func SystemInfo() *Environment {
	it := &Environment{}
	it.vendor().kernel().release().cpuTitle()
	switch runtime.GOOS {

	case "windows":
		it.Platform = runtime.GOOS
		it.Kernel = "NT " + strings.Fields(it.Kernel)[0]
		it.Graphics = strings.Join(graphics(), ", ")

	case "linux":
		it.Kernel = "Linux " + strings.Split(it.Kernel, "-")[0]
		if fp, _ := exec.LookPath("systemctl"); fp != "" {
			it.Init = "systemd"
		} else if fp, _ = exec.LookPath("service"); fp != "" {
			it.Init = "upstart" // sysvinit
		} else {
			it.Init = "no init"
		}
		if "root" == os.Getenv("USER") {
			it.Graphics = strings.Join(graphics(), ", ")
		}
		if it.Platform == "" {
			it.android()
		}
	}

	it.Platform += "/" + runtime.GOARCH
	return it
}

func (it *Environment) cpuTitle() *Environment {
	it.Perf = processorSpeed()
	stat, _ := cpu.Info()
	if len(stat) == 0 {
		return it
	}
	switch {
	case strings.HasPrefix(stat[0].ModelName, "AMD"):
		it.Processor = strings.TrimSpace(strings.Split(stat[0].ModelName, "with")[0])
	case strings.HasPrefix(stat[0].ModelName, "Intel"):
		it.Processor = strings.TrimSpace(strings.Split(stat[0].ModelName, "@")[0])
	default:
		it.Processor = stat[0].ModelName
		if it.Processor == "" {
			fp := "/proc/cpuinfo"
			for _, elem := range strings.Split(String(&fp), "\n") {
				if strings.HasPrefix(elem, "Hardware") {
					it.Processor = strings.TrimSpace(strings.Split(elem, ":")[1])
				}
				if strings.HasPrefix(elem, "Model") {
					it.Vendor = strings.TrimSpace(strings.Split(elem, ":")[1])
				}
			}
		}
	}
	pei := carefullySelectedCPUs[it.Processor]
	if pei.Power != "" {
		it.Perf += fmt.Sprintf("Hertz=Max:%s.T%d Power:%s", pei.HertzMax, runtime.NumCPU(), pei.Power)
	} else {
		it.Perf += fmt.Sprintf("Hertz=%.2fG.T%d", stat[0].Mhz/1000, runtime.NumCPU())
	}
	info, _ := mem.VirtualMemory()
	if info != nil {
		it.Perf += fmt.Sprintf(" - Memory(%s)", SizeFormat(float64(info.Total)))
	}
	return it
}

var carefullySelectedCPUs = map[string]ProcessorExtensionInformation{

	// 锐龙™ 线程撕裂者™

	"AMD Ryzen Threadripper PRO 5995WX": {"4.50G", "280W"},
	"AMD Ryzen Threadripper PRO 5975WX": {"4.50G", "280W"},
	"AMD Ryzen Threadripper PRO 5965WX": {"4.50G", "280W"},
	"AMD Ryzen Threadripper PRO 5955WX": {"4.50G", "280W"},
	"AMD Ryzen Threadripper PRO 5945WX": {"4.50G", "280W"},
	"AMD Ryzen Threadripper PRO 3995WX": {"4.20G", "280W"},
	"AMD Ryzen Threadripper PRO 3975WX": {"4.20G", "280W"},
	"AMD Ryzen Threadripper PRO 3955WX": {"4.30G", "280W"},
	"AMD Ryzen Threadripper 3990X":      {"4.30G", "280W"},
	"AMD Ryzen Threadripper 3970X":      {"4.50G", "280W"},
	"AMD Ryzen Threadripper 3960X":      {"4.50G", "280W"},

	// 锐龙™ 6000

	"AMD Ryzen 9 6980HX": {"5.00G", "45W"},
	"AMD Ryzen 9 6980HS": {"5.00G", "35W"},
	"AMD Ryzen 9 6900HX": {"4.90G", "45W"},
	"AMD Ryzen 9 6900HS": {"4.90G", "35W"},
	"AMD Ryzen 7 6800H":  {"4.70G", "45W"},
	"AMD Ryzen 7 6800HS": {"4.70G", "35W"},
	"AMD Ryzen 5 6600H":  {"4.50G", "45W"},
	"AMD Ryzen 5 6600HS": {"4.50G", "35W"},

	// 锐龙™ 5000

	"AMD Ryzen 9 5980HX": {"4.80G", "45W"},
	"AMD Ryzen 9 5980HS": {"4.80G", "35W"},
	"AMD Ryzen 9 5900HX": {"4.60G", "45W"},
	"AMD Ryzen 9 5900HS": {"4.60G", "35W"},
	"AMD Ryzen 7 5800H":  {"4.40G", "45W"},
	"AMD Ryzen 7 5800HS": {"4.40G", "35W"},
	"AMD Ryzen 5 5600H":  {"4.20G", "45W"},
	"AMD Ryzen 5 5600HS": {"4.20G", "35W"},

	// 酷睿™ 12 Gen

	"Intel(R) Core(TM) i9-12900HK CPU": {"5.00G", "45W"},
	"Intel(R) Core(TM) i9-12900H CPU":  {"5.00G", "45W"},
	"Intel(R) Core(TM) i7-12800H CPU":  {"4.80G", "45W"},
	"Intel(R) Core(TM) i7-12700H CPU":  {"4.70G", "45W"},
	"Intel(R) Core(TM) i7-12650H CPU":  {"4.70G", "45W"},
	"Intel(R) Core(TM) i5-12600H CPU":  {"4.50G", "45W"},
	"Intel(R) Core(TM) i5-12500H CPU":  {"4.50G", "45W"},
	"Intel(R) Core(TM) i5-12450H CPU":  {"4.40G", "45W"},
}

type ProcessorExtensionInformation struct {
	HertzMax string
	Power    string
}