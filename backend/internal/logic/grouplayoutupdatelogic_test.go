package logic

import (
	"errors"
	"net/http"
	"testing"

	"class-management-system/backend/internal/httperr"
	"class-management-system/backend/internal/model"
	"class-management-system/backend/internal/types"
)

func TestValidateGroupLayout(t *testing.T) {
	groups := []model.Group{
		{BaseModel: model.BaseModel{ID: 11}},
		{BaseModel: model.BaseModel{ID: 22}},
		{BaseModel: model.BaseModel{ID: 33}},
	}

	valid := &types.GroupLayoutUpdateReq{Items: []types.GroupLayoutItem{
		{Id: 22, Position: 1},
		{Id: 33, Position: 2},
		{Id: 11, Position: 3},
	}}
	positions, err := validateGroupLayout(valid, groups)
	if err != nil {
		t.Fatalf("validateGroupLayout() rejected valid layout: %v", err)
	}
	if len(positions) != 3 || positions[0].ID != 22 || positions[0].Position != 1 {
		t.Fatalf("validateGroupLayout() positions = %#v", positions)
	}

	tests := []struct {
		name string
		req  *types.GroupLayoutUpdateReq
	}{
		{name: "nil request"},
		{name: "missing group", req: &types.GroupLayoutUpdateReq{Items: valid.Items[:2]}},
		{name: "unknown group", req: &types.GroupLayoutUpdateReq{Items: []types.GroupLayoutItem{{Id: 11, Position: 1}, {Id: 22, Position: 2}, {Id: 44, Position: 3}}}},
		{name: "duplicate group", req: &types.GroupLayoutUpdateReq{Items: []types.GroupLayoutItem{{Id: 11, Position: 1}, {Id: 11, Position: 2}, {Id: 33, Position: 3}}}},
		{name: "duplicate position", req: &types.GroupLayoutUpdateReq{Items: []types.GroupLayoutItem{{Id: 11, Position: 1}, {Id: 22, Position: 1}, {Id: 33, Position: 3}}}},
		{name: "position too low", req: &types.GroupLayoutUpdateReq{Items: []types.GroupLayoutItem{{Id: 11, Position: 0}, {Id: 22, Position: 2}, {Id: 33, Position: 3}}}},
		{name: "position too high", req: &types.GroupLayoutUpdateReq{Items: []types.GroupLayoutItem{{Id: 11, Position: 1}, {Id: 22, Position: 2}, {Id: 33, Position: 4}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validateGroupLayout(tt.req, groups)
			var httpErr *httperr.Error
			if !errors.As(err, &httpErr) || httpErr.Code != http.StatusBadRequest {
				t.Fatalf("validateGroupLayout() error = %v, want HTTP 400", err)
			}
		})
	}
}
