// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"class-management-system/backend/internal/httperr"
	"class-management-system/backend/internal/model"
	"class-management-system/backend/internal/repository"
	"class-management-system/backend/internal/svc"
	"class-management-system/backend/internal/types"

	"github.com/xuri/excelize/v2"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type StudentImportLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewStudentImportLogic(ctx context.Context, svcCtx *svc.ServiceContext) *StudentImportLogic {
	return &StudentImportLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *StudentImportLogic) StudentImport(r *http.Request) (resp *types.Empty, err error) {
	if r == nil {
		return nil, badRequest("invalid request")
	}
	ct := r.Header.Get("Content-Type")
	mediaType, _, _ := mime.ParseMediaType(ct)
	if !strings.HasPrefix(mediaType, "multipart/") {
		return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "Content-Type must be multipart/form-data"}
	}
	const maxUploadSize = 10 << 20
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		return nil, err
	}
	f, hdr, err := r.FormFile("file")
	if err != nil {
		return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "missing file"}
	}
	defer func() { _ = f.Close() }()
	if ext := strings.ToLower(filepath.Ext(hdr.Filename)); ext != ".xlsx" {
		return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "only .xlsx supported"}
	}
	if hdr.Size <= 0 {
		return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "empty file"}
	}
	if hdr.Size > maxUploadSize {
		return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "file too large (max 10MB)"}
	}
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	wb, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer func() { _ = wb.Close() }()
	sheet := wb.GetSheetName(0)
	if sheet == "" {
		return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "empty workbook"}
	}
	rows, err := wb.GetRows(sheet)
	if err != nil {
		return nil, err
	}
	if len(rows) <= 1 {
		return &types.Empty{}, nil
	}
	header := rows[0]
	studentNoH := strings.ToLower(cell(header, 0))
	nameH := strings.ToLower(cell(header, 1))
	if studentNoH != "studentno" && studentNoH != "学号" {
		return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "invalid header: column A must be StudentNo/学号"}
	}
	if nameH != "name" && nameH != "姓名" {
		return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: "invalid header: column B must be Name/姓名"}
	}

	students := make([]model.Student, 0, len(rows)-1)
	rowErrors := make([]string, 0)
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		studentNo := cell(row, 0)
		name := cell(row, 1)
		gender := cell(row, 2)
		phone := cell(row, 3)
		position := cell(row, 4)
		if studentNo == "" || name == "" {
			rowErrors = append(rowErrors, fmt.Sprintf("row %d: missing studentNo or name", i+1))
			continue
		}
		students = append(students, model.Student{
			StudentNo:  studentNo,
			Name:       name,
			Gender:     gender,
			Phone:      phone,
			Position:   position,
			GroupID:    0,
			TotalScore: 0,
		})
	}
	if len(rowErrors) > 0 {
		return nil, &httperr.Error{Code: http.StatusBadRequest, Msg: strings.Join(rowErrors, "; ")}
	}
	if err := l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		repo := repository.NewStudentRepo(tx)
		return repo.BatchUpsertByStudentNo(l.ctx, students)
	}); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

func cell(row []string, idx int) string {
	if idx < 0 || idx >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[idx])
}
