package db

import (
    "database/sql/driver"
    "fmt"
    "strconv"
    "strings"
)

type Vector []float32

func (v *Vector) Scan(src any) error {
    switch s := src.(type) {
    case string:
        return v.parse(s)
    case []byte:
        return v.parse(string(s))
    default:
        return fmt.Errorf("unsupported type %T", src)
    }
}

func (v *Vector) parse(s string) error {
    s = strings.Trim(s, "[]")
    parts := strings.Split(s, ",")
    vec := make([]float32, len(parts))
    for i, p := range parts {
        f, err := strconv.ParseFloat(strings.TrimSpace(p), 32)
        if err != nil {
            return err
        }
        vec[i] = float32(f)
    }
    *v = vec
    return nil
}

func (v Vector) Value() (driver.Value, error) {
    parts := make([]string, len(v))
    for i, f := range v {
        parts[i] = fmt.Sprintf("%f", f)
    }
    return fmt.Sprintf("[%s]", strings.Join(parts, ",")), nil
}