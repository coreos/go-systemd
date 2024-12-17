//go:generate stringer -type=ScheduleShutdownType -linecomment

package login1

type ScheduleShutdownType int

const (
	ScheduleTypePowerOff    ScheduleShutdownType = iota // poweroff
	ScheduleTypeDryPowerOff                             // dry-poweroff
	ScheduleTypeReboot                                  // reboot
	ScheduleTypeDryReboot                               // dry-reboot
	ScheduleTypeHalt                                    // halt
	SceduleTypeDryHalt                                  // dry-halt
)

const (
	SD_LOGIND_ROOT_CHECK_INHIBITORS          = uint64(1) << 0
	SD_LOGIND_KEXEC_REBOOT                   = uint64(1) << 1
	SD_LOGIND_SOFT_REBOOT                    = uint64(1) << 2
	SD_LOGIND_SOFT_REBOOT_IF_NEXTROOT_SET_UP = uint64(1) << 3
	SD_LOGIND_SKIP_INHIBITORS                = uint64(1) << 4
)
