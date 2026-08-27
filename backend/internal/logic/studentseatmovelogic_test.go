package logic

import (
	"errors"
	"net/http"
	"testing"

	"class-management-system/backend/internal/httperr"
	"class-management-system/backend/internal/types"
)

func TestValidateStudentSeatMove(t *testing.T) {
	tests := []struct {
		name string
		req  *types.StudentSeatMoveReq
		want string
	}{
		{name: "valid swap", req: &types.StudentSeatMoveReq{StudentId: 11, GroupId: 2, TargetStudentId: 12}},
		{name: "valid move to unassigned", req: &types.StudentSeatMoveReq{StudentId: 11, GroupId: 0}},
		{name: "nil request", want: "invalid studentId"},
		{name: "missing student", req: &types.StudentSeatMoveReq{GroupId: 2}, want: "invalid studentId"},
		{name: "negative group", req: &types.StudentSeatMoveReq{StudentId: 11, GroupId: -1}, want: "invalid groupId"},
		{name: "negative target", req: &types.StudentSeatMoveReq{StudentId: 11, GroupId: 2, TargetStudentId: -1}, want: "invalid targetStudentId"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateStudentSeatMove(tt.req)
			if tt.want == "" {
				if err != nil {
					t.Fatalf("validateStudentSeatMove() error = %v", err)
				}
				return
			}
			var httpErr *httperr.Error
			if !errors.As(err, &httpErr) {
				t.Fatalf("validateStudentSeatMove() error = %T, want *httperr.Error", err)
			}
			if httpErr.Code != http.StatusBadRequest || httpErr.Msg != tt.want {
				t.Fatalf("validateStudentSeatMove() error = %#v, want code %d message %q", httpErr, http.StatusBadRequest, tt.want)
			}
		})
	}
}
