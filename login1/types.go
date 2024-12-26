//go:generate go run golang.org/x/tools/cmd/stringer@latest -linecomment -type=ScheduleShutdownType

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
	// Request for weak inhibitors to apply to privileged users too.
	SD_LOGIND_ROOT_CHECK_INHIBITORS = uint64(1) << 0
	// Perform kexec reboot if a kexec kernel is loaded.
	SD_LOGIND_KEXEC_REBOOT = uint64(1) << 1
	// When SD_LOGIND_SOFT_REBOOT (0x04) is set, or SD_LOGIND_SOFT_REBOOT_IF_NEXTROOT_SET_UP
	// (0x08) is set and a new root file system has been set up on "/run/nextroot/",
	// then RebootWithFlags() performs a userspace reboot only.
	SD_LOGIND_SOFT_REBOOT                    = uint64(1) << 2
	SD_LOGIND_SOFT_REBOOT_IF_NEXTROOT_SET_UP = uint64(1) << 3
	// Skip all inhibitors.
	SD_LOGIND_SKIP_INHIBITORS = uint64(1) << 4
)
