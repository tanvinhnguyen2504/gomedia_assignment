package internal

import (
	"time"
)

func (req CreateViewingRequest) Validate() error {
	if req.AgentID == 0 || req.LeadID == 0 || req.PropertyAddress == "" || req.ScheduledAt.IsZero() {
		return ErrMissingField
	}
	if !req.ScheduledAt.After(time.Now()) {
		return ErrPastDate
	}
	return nil
}

func isValidAction(action ViewingAction) bool {
	return action == ActionComplete || action == ActionCancel || action == ActionUpdateNotes
}

func (req BulkUpdateRequest) Validate() error {
	if len(req.IDs) == 0 {
		return ErrMissingField
	}

	if !isValidAction(req.Action) {
		return ErrInvalidAction
	}

	if req.Action == ActionUpdateNotes && req.Notes == nil {
		return ErrMissingField
	}
	return nil
}
