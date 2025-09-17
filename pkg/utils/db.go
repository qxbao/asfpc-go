package utils

import (
	"database/sql"
	"strings"
	"github.com/qxbao/asfpc/infras"
)

func ToNullString(ptr *string) sql.NullString {
	if ptr != nil {
		return sql.NullString{String: *ptr, Valid: true}
	}
	return sql.NullString{Valid: false}
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