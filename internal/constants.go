package internal

type ViewingStatus string
type ViewingAction string

const (
	StatusScheduled ViewingStatus = "SCHEDULED"
	StatusCompleted ViewingStatus = "COMPLETED"
	StatusCancelled ViewingStatus = "CANCELLED"
	StatusMissed    ViewingStatus = "MISSED"

	ActionCancel      ViewingAction = "CANCEL"
	ActionComplete    ViewingAction = "COMPLETE"
	ActionUpdateNotes ViewingAction = "UPDATE_NOTES"
)

const (
	DEFAULT_LIMIT = 20
	MAX_LIMIT     = 100
)
