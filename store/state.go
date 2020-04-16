package store

type DealState int

const (
	Unknown = DealState(iota)
	Success
	Dealing
	Expired
	Failed
)

func TransferState(state string) int {
	switch state {
	case "unknow":
		return 0
	case "complete":
		return 1
	case "started", "staged", "accepted", "sealing":
		return 2
	case "failed", "error", "rejected":
		return 4
	default:
		return 0
	}
}
