package db

import (
	"database/sql/driver"
	"reflect"

	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/pop/v6/columns"
	"github.com/pkg/errors"
)

type NullString string

func (s *NullString) Scan(value interface{}) error {
	if value == nil {
		*s = ""
		return nil
	}
	strVal, ok := value.(string)
	if !ok {
		return errors.New("column is not a string")
	}
	*s = NullString(strVal)
	return nil
}
func (s NullString) Value() (driver.Value, error) {
	if len(s) == 0 { // if nil or empty string
		return nil, nil
	}
	return string(s), nil
}

func (s NullString) String() string {
	return string(s)
}

func getExcludedColumns(model interface{}, includeColumns ...string) ([]string, error) {
	sm := &pop.Model{Value: model}
	st := reflect.TypeOf(model)
	if st.Kind() == reflect.Ptr {
		_ = st.Elem()
	}

	// get all columns and remove included to get excluded set
	cols := columns.ForStructWithAlias(model, sm.TableName(), sm.As, sm.IDField())
	for _, f := range includeColumns {
		if _, ok := cols.Cols[f]; !ok {
			return nil, errors.Errorf("Invalid column name %s", f)
		}
		cols.Remove(f)
	}

	xcols := make([]string, 0, len(cols.Cols))
	for n := range cols.Cols {
		// gobuffalo updates the updated_at column automatically
		if n == "updated_at" {
			continue
		}
		xcols = append(xcols, n)
	}

	return xcols, nil
}
