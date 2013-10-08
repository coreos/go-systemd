package dbus

import (
	"github.com/guelfey/go.dbus"
)

// From the systemd docs:
//
// The properties array of StartTransientUnit() may take many of the settings
// that may also be configured in unit files. Not all parameters are currently
// accepted though, but we plan to cover more properties with future release.
// Currently you may set the Description, Slice and all dependency types of
// units, as well as RemainAfterExit, ExecStart for service units,
// TimeoutStopUSec and PIDs for scope units, and CPUAccounting, CPUShares,
// BlockIOAccounting, BlockIOWeight, BlockIOReadBandwidth,
// BlockIOWriteBandwidth, BlockIODeviceWeight, MemoryAccounting, MemoryLimit,
// DevicePolicy, DeviceAllow for services/scopes/slices. These fields map
// directly to their counterparts in unit files and as normal D-Bus object
// properties. The exception here is the PIDs field of scope units which is
// used for construction of the scope only and specifies the initial PIDs to
// add to the scope object.

type Property property

type property struct {
	Name  string
	Value dbus.Variant
}

type execStart struct {
	Path             string   // the binary path to execute
	Args             []string // an array with all arguments to pass to the executed command, starting with argument 0
	UncleanIsFailure bool     // a boolean whether it should be considered a failure if the process exits uncleanly
}

// PropExecStart sets the ExecStart service property.  The first argument is a
// slice with the binary path to execute followed by the arguments to pass to
// the executed command. See
// http://www.freedesktop.org/software/systemd/man/systemd.service.html#ExecStart=
func PropExecStart(command []string, uncleanIsFailure bool) Property {
	return Property(
		property{
			Name: "ExecStart",
			Value: dbus.MakeVariant(
				[]execStart{
					execStart{
						Path:             command[0],
						Args:             command,
						UncleanIsFailure: uncleanIsFailure,
					}})})
}

// PropRemainAfterExit sets the RemainAfterExit service property. See
// http://www.freedesktop.org/software/systemd/man/systemd.service.html#RemainAfterExit=
func PropRemainAfterExit(b bool) Property {
	return Property(
		property{
			Name:  "RemainAfterExit",
			Value: dbus.MakeVariant(b),
		})
}

// PropDescription sets the Description unit property. See
// http://www.freedesktop.org/software/systemd/man/systemd.unit#Description=
func PropDescription(desc string) Property {
	return Property(
		property{
			Name:  "Description",
			Value: dbus.MakeVariant(desc),
		})
}

func propDependency(name string, units []string) Property {
	return Property(
		property{
			Name:  name,
			Value: dbus.MakeVariant(units),
		})
}

// PropRequires sets the Requires unit property.  See
// http://www.freedesktop.org/software/systemd/main/systemd.unit.html#Requires=
func PropRequires(units ...string) Property {
	return propDependency("Requires", units)
}

// PropRequiresOverridable sets the RequiresOverridable unit property.  See
// http://www.freedesktop.org/software/systemd/main/systemd.unit.html#RequiresOverridable=
func PropRequiresOverridable(units ...string) Property {
	return propDependency("RequiresOverridable", units)
}

// PropRequisite sets the Requisite unit property.  See
// http://www.freedesktop.org/software/systemd/main/systemd.unit.html#Requisite=
func PropRequisite(units ...string) Property {
	return propDependency("Requisite", units)
}

// PropRequisiteOverridable sets the RequisiteOverridable unit property.  See
// http://www.freedesktop.org/software/systemd/main/systemd.unit.html#RequisiteOverridable=
func PropRequisiteOverridable(units ...string) Property {
	return propDependency("RequisiteOverridable", units)
}

// PropWants sets the Wants unit property.  See
// http://www.freedesktop.org/software/systemd/main/systemd.unit.html#Wants=
func PropWants(units ...string) Property {
	return propDependency("Wants", units)
}

// PropBindsTo sets the BindsTo unit property.  See
// http://www.freedesktop.org/software/systemd/main/systemd.unit.html#BindsTo=
func PropBindsTo(units ...string) Property {
	return propDependency("BindsTo", units)
}

// PropRequiredBy sets the RequiredBy unit property.  See
// http://www.freedesktop.org/software/systemd/main/systemd.unit.html#RequiredBy=
func PropRequiredBy(units ...string) Property {
	return propDependency("RequiredBy", units)
}

// PropRequiredByOverridable sets the RequiredByOverridable unit property.  See
// http://www.freedesktop.org/software/systemd/main/systemd.unit.html#RequiredByOverridable=
func PropRequiredByOverridable(units ...string) Property {
	return propDependency("RequiredByOverridable", units)
}

// PropWantedBy sets the WantedBy unit property.  See
// http://www.freedesktop.org/software/systemd/main/systemd.unit.html#WantedBy=
func PropWantedBy(units ...string) Property {
	return propDependency("WantedBy", units)
}

// PropBoundBy sets the BoundBy unit property.  See
// http://www.freedesktop.org/software/systemd/main/systemd.unit.html#BoundBy=
func PropBoundBy(units ...string) Property {
	return propDependency("BoundBy", units)
}

// PropConflicts sets the Conflicts unit property.  See
// http://www.freedesktop.org/software/systemd/main/systemd.unit.html#Conflicts=
func PropConflicts(units ...string) Property {
	return propDependency("Conflicts", units)
}

// PropConflictedBy sets the ConflictedBy unit property.  See
// http://www.freedesktop.org/software/systemd/main/systemd.unit.html#ConflictedBy=
func PropConflictedBy(units ...string) Property {
	return propDependency("ConflictedBy", units)
}

// PropBefore sets the Before unit property.  See
// http://www.freedesktop.org/software/systemd/main/systemd.unit.html#Before=
func PropBefore(units ...string) Property {
	return propDependency("Before", units)
}

// PropAfter sets the After unit property.  See
// http://www.freedesktop.org/software/systemd/main/systemd.unit.html#After=
func PropAfter(units ...string) Property {
	return propDependency("After", units)
}

// PropOnFailure sets the OnFailure unit property.  See
// http://www.freedesktop.org/software/systemd/main/systemd.unit.html#OnFailure=
func PropOnFailure(units ...string) Property {
	return propDependency("OnFailure", units)
}

// PropTriggers sets the Triggers unit property.  See
// http://www.freedesktop.org/software/systemd/main/systemd.unit.html#Triggers=
func PropTriggers(units ...string) Property {
	return propDependency("Triggers", units)
}

// PropTriggeredBy sets the TriggeredBy unit property.  See
// http://www.freedesktop.org/software/systemd/main/systemd.unit.html#TriggeredBy=
func PropTriggeredBy(units ...string) Property {
	return propDependency("TriggeredBy", units)
}

// PropPropagatesReloadTo sets the PropagatesReloadTo unit property.  See
// http://www.freedesktop.org/software/systemd/main/systemd.unit.html#PropagatesReloadTo=
func PropPropagatesReloadTo(units ...string) Property {
	return propDependency("PropagatesReloadTo", units)
}

// PropRequiresMountsFor sets the RequiresMountsFor unit property.  See
// http://www.freedesktop.org/software/systemd/main/systemd.unit.html#RequiresMountsFor=
func PropRequiresMountsFor(units ...string) Property {
	return propDependency("RequiresMountsFor", units)
}
