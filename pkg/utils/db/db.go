package db

import (
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/qxbao/asfpc/db"
	"github.com/qxbao/asfpc/infras"
)

type GroupWithCategories struct {
	db.GetGroupsByAccountIdRow
	Categories json.RawMessage `json:"categories"`
}

type ProfileAnalysisRow struct {
	ID           int32           `json:"id"`
	FacebookID   string          `json:"facebook_id"`
	Name         sql.NullString  `json:"name"`
	IsAnalyzed   sql.NullBool    `json:"is_analyzed"`
	Categories   json.RawMessage `json:"categories"`
	NonNullCount int32           `json:"non_null_count"`
}

func ToNullString(ptr *string) sql.NullString {
	if ptr != nil {
		return sql.NullString{String: *ptr, Valid: true}
	}
	return sql.NullString{Valid: false}
}

func ToNullInt32(ptr *int32) sql.NullInt32 {
	if ptr != nil {
		return sql.NullInt32{Int32: *ptr, Valid: true}
	}
	return sql.NullInt32{Valid: false}
}

func ExtractEntityName(entity *infras.EntityNameID) sql.NullString {
	if entity != nil && entity.Name != nil {
		return sql.NullString{String: *entity.Name, Valid: true}
	}
	return sql.NullString{Valid: false}
}

func JoinWork(work *[]infras.Work) sql.NullString {
	if work == nil || len(*work) == 0 {
		return sql.NullString{Valid: false}
	}

	var workStrings []string
	for _, w := range *work {
		if w.Employer != nil && w.Employer.Name != nil {
			workStr := *w.Employer.Name
			if w.Position != nil && w.Position.Name != nil {
				workStr += " - " + *w.Position.Name
			}
			workStrings = append(workStrings, workStr)
		}
	}

	if len(workStrings) > 0 {
		return sql.NullString{String: strings.Join(workStrings, "; "), Valid: true}
	}
	return sql.NullString{Valid: false}
}

func JoinEducation(education *[]infras.Education) sql.NullString {
	if education == nil || len(*education) == 0 {
		return sql.NullString{Valid: false}
	}

	var eduStrings []string
	for _, edu := range *education {
		if edu.School != nil && edu.School.Name != nil {
			eduStrings = append(eduStrings, *edu.School.Name)
		}
	}

	if len(eduStrings) > 0 {
		return sql.NullString{String: strings.Join(eduStrings, "; "), Valid: true}
	}
	return sql.NullString{Valid: false}
}

func GetStringOrDefault(ptr *string, defaultValue string) string {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}

func ConvertGroupRow(row db.GetGroupsByAccountIdRow) GroupWithCategories {
	result := GroupWithCategories{
		GetGroupsByAccountIdRow: row,
	}

	if len(row.Categories) == 0 {
		result.Categories = json.RawMessage([]byte("[]"))
	} else {
		var test any
		if err := json.Unmarshal(row.Categories, &test); err != nil {
			result.Categories = json.RawMessage([]byte("[]"))
		} else {
			result.Categories = json.RawMessage(row.Categories)
		}
	}

	return result
}