package state

const (
	Invalid = iota
	New
	Paused
	Running
	Stopped
	Restarting
	Terminating
	Terminated
)

var stateNames = map[int]string{
	Invalid:    "Invalid",
	New:        "New",
	Paused:     "Paused",
	Running:    "Running",
	Stopped:    "Stopped",
	Restarting: "Restarting",
	Terminating: "Terminating",
	Terminated: "Terminated",
}