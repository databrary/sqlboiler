package bdb

import "fmt"

// Table metadata from the database schema.
type Table struct {
	Name string
	// For dbs with real schemas, like Postgres.
	// Example value: "schema_name"."table_name"
	SchemaName string
	Columns    []Column
	HasCustom  bool
	IsView     bool
	PKey       *PrimaryKey
	FKeys      []ForeignKey

	IsJoinTable bool

	ToOneRelationships  []ToOneRelationship
	ToManyRelationships []ToManyRelationship
}

// GetTable by name. Panics if not found (for use in templates mostly).
func GetTable(tables []Table, name string) (tbl Table) {
	for _, t := range tables {
		if t.Name == name {
			return t
		}
	}

	panic(fmt.Sprintf("could not find table name: %s", name))
}

// GetColumn by name. Panics if not found (for use in templates mostly).
func (t Table) GetColumn(name string) (col Column) {
	for _, c := range t.Columns {
		if c.Name == name {
			return c
		}
	}

	panic(fmt.Sprintf("could not find column name: %s", name))
}

// CanLastInsertID checks the following:
// 1. Is there only one primary key?
// 2. Does the primary key column have a default value?
// 3. Is the primary key column type one of uintX/intX?
// If the above is all true, this table can use LastInsertId
func (t Table) CanLastInsertID() bool {
	if t.PKey == nil || len(t.PKey.Columns) != 1 {
		return false
	}

	col := t.GetColumn(t.PKey.Columns[0])
	if len(col.Default) == 0 {
		return false
	}

	switch col.Type {
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
	default:
		return false
	}

	return true
}

func (t Table) GetCustomColumns() []Column {
	ret := []Column{}
	if t.HasCustom {
		for _, v := range t.Columns {
			if v.IsCustom {
				ret = append(ret, v)
			}
		}
	}
	return ret
}

func (t Table) HasPrimaryKey() bool {
	return t.PKey != nil
}
