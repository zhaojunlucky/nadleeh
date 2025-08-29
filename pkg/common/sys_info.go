package common

import (
	"fmt"
	"runtime"

	"github.com/shirou/gopsutil/v4/host"
	log "github.com/sirupsen/logrus"
	"github.com/zhaojunlucky/golib/pkg/env"
)

var Sys = newSysInfo()

type SysInfo struct {
	sysInfo env.Env
}

func (s *SysInfo) init() {
	s.sysInfo = env.NewEmptyRWEnv()

	info, err := host.Info()
	if err != nil {
		log.Fatalf("Error getting host info: %v", err)
	}

	s.sysInfo.Set("CPU_COUNT", fmt.Sprintf("%d", runtime.NumCPU()))

	s.sysInfo.Set("OS", info.OS)
	s.sysInfo.Set("OS_ARCH", info.KernelArch)
	s.sysInfo.Set("PLATFORM", info.Platform)
	s.sysInfo.Set("PLATFORM_FAMILY", info.PlatformFamily)
	s.sysInfo.Set("PLATFORM_VERSION", info.PlatformVersion)
	s.sysInfo.Set("KERNEL_VERSION", info.KernelVersion)
	s.sysInfo.Set("HOSTNAME", info.Hostname)
}

func (s *SysInfo) GetInfo() env.Env {
	return s.sysInfo
}

func newSysInfo() *SysInfo {
	sys := &SysInfo{}
	sys.init()
	return sys
}
