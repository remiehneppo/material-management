package handler

import (
	"errors"
	"net/http"
	"testing"

	"github.com/remiehneppo/material-management/internal/domain/materialrequest"
	"github.com/remiehneppo/material-management/types"
)

func TestIssueErrorResponse(t *testing.T) {
	tests := []struct {
		err        error
		wantStatus int
		wantDetail bool
	}{
		{materialrequest.ErrRequesterMismatch, http.StatusForbidden, true},
		{types.ErrMaterialRequestNotFound, http.StatusNotFound, true},
		{materialrequest.ErrDraftRequired, http.StatusConflict, true},
		{types.ErrUnauthorized, http.StatusUnauthorized, true},
		{errors.New("database unavailable"), http.StatusInternalServerError, false},
	}
	for _, test := range tests {
		status, message := issueErrorResponse(test.err)
		if status != test.wantStatus {
			t.Fatalf("status = %d, want %d", status, test.wantStatus)
		}
		if (message == test.err.Error()) != test.wantDetail {
			t.Fatalf("message %q detail exposure = %v, want %v", message, message == test.err.Error(), test.wantDetail)
		}
	}
}
