package db

import (
	"database/sql"
	"testing"

	"github.com/qxbao/asfpc/infras"
)

func strPtr(s string) *string {
	return &s
}

func GetIntOrDefault(ptr *int32, defaultValue int32) int32 {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}

func TestToNullString(t *testing.T) {
	tests := []struct {
		name string
		input *string
		expected sql.NullString
	}{
		{"nil input", nil, sql.NullString{Valid: false}},
		{"empty string", strPtr(""), sql.NullString{String: "", Valid: true}},
		{"non-empty string", strPtr("test"), sql.NullString{String: "test", Valid: true}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToNullString(tt.input)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestExtractEntityName(t *testing.T) {
	tests := []struct {
		name string
		input *infras.EntityNameID
		expected sql.NullString
	}{
		{"nil input", nil, sql.NullString{Valid: false}},
		{"nil name", &infras.EntityNameID{Id: strPtr("1"), Name: nil}, sql.NullString{Valid: false}},
		{"valid name", &infras.EntityNameID{Id: strPtr("1"), Name: strPtr("Entity")}, sql.NullString{String: "Entity", Valid: true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractEntityName(tt.input)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestJoinWork(t *testing.T) {
	tests := []struct {
		name string
		input *[]infras.Work
		expected sql.NullString
	}{
		{"nil input", nil, sql.NullString{Valid: false}},
		{"empty slice", &[]infras.Work{}, sql.NullString{Valid: false}},
		{"single work entry", &[]infras.Work{
			{Employer: &infras.EntityNameID{Name: strPtr("Company A")}, Position: &infras.EntityNameID{Name: strPtr("Developer")}},
		}, sql.NullString{String: "Company A - Developer", Valid: true}},
		{"multiple work entries", &[]infras.Work{
			{Employer: &infras.EntityNameID{Name: strPtr("Company A")}, Position: &infras.EntityNameID{Name: strPtr("Developer")}},
			{Employer: &infras.EntityNameID{Name: strPtr("Company B")}, Position: &infras.EntityNameID{Name: strPtr("Manager")}},
		}, sql.NullString{String: "Company A - Developer; Company B - Manager", Valid: true}},
		{"work entry with nil position", &[]infras.Work{
			{Employer: &infras.EntityNameID{Name: strPtr("Company A")}, Position: nil},
		}, sql.NullString{String: "Company A", Valid: true}},
		{"work entry with nil employer", &[]infras.Work{
			{Employer: nil, Position: &infras.EntityNameID{Name: strPtr("Developer")}},
		}, sql.NullString{Valid: false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := JoinWork(tt.input)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestJoinEducation(t *testing.T) {
	tests := []struct {
		name string
		input *[]infras.Education
		expected sql.NullString
	}{
		{"nil input", nil, sql.NullString{Valid: false}},
		{"empty slice", &[]infras.Education{}, sql.NullString{Valid: false}},
		{"single education entry", &[]infras.Education{
			{School: &infras.EntityNameID{Name: strPtr("University A")}},
		}, sql.NullString{String: "University A", Valid: true}},
		{"multiple education entries", &[]infras.Education{
			{School: &infras.EntityNameID{Name: strPtr("University A")}},
			{School: &infras.EntityNameID{Name: strPtr("College B")}},
		}, sql.NullString{String: "University A; College B", Valid: true}},
		{"education entry with nil school", &[]infras.Education{
			{School: nil},
		}, sql.NullString{Valid: false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := JoinEducation(tt.input)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestGetStringOrDefault(t *testing.T) {
	tests := []struct {
		name string
		input *string
		defaultValue string
		expected string
	}{
		{"nil input", nil, "default", "default"},
		{"empty string", strPtr(""), "default", ""},
		{"non-empty string", strPtr("value"), "default", "value"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetStringOrDefault(tt.input, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
