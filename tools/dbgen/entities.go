package dbgen

import (
	"reflect"
	"strings"

	"github.com/libramusic/libracore/media"
)

// ColumnDef represents a database column definition.
type ColumnDef struct {
	Name          string // Go field name
	DBName        string // Database column name (from json tag)
	GoType        string
	NeedsEncoding bool
}

// EntityDef represents a database entity/table definition.
type EntityDef struct {
	EntityConfig

	Columns      []ColumnDef
	PKGoName     string // Primary key Go field name (e.g., "ID")
	PKColumn     string // Primary key DB column name
	UserIDColumn string // User ID column name (for user-scoped queries)
}

// EntityConfig provides custom metadata for an entity detected via reflection.
// It only needs to override defaults derived from reflection.
type EntityConfig struct {
	Name         string
	TableName    string
	StructType   string
	HasUserID    bool
	AllMethod    string
	GetMethod    string
	NewMethod    string
	UpdateMethod string
	DeleteMethod string
}

// needsEncoding returns true if the Go type is a slice or map.
func needsEncoding(t reflect.Type) bool {
	return t.Kind() == reflect.Slice || t.Kind() == reflect.Map
}

// goTypeString returns the Go type as a string.
func goTypeString(t reflect.Type) string {
	if t.Kind() == reflect.Slice {
		return "[]" + goTypeString(t.Elem())
	}
	return t.String()
}

// inferColumns uses reflection to extract columns from a struct type,
// including fields from embedded structs.
func inferColumns(t reflect.Type) []ColumnDef {
	var columns []ColumnDef
	for i := range t.NumField() {
		field := t.Field(i)

		// Handle embedded structs - flatten their fields.
		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			embeddedCols := inferColumns(field.Type)
			columns = append(columns, embeddedCols...)
			continue
		}

		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		dbName, _, _ := strings.Cut(jsonTag, ",")
		if dbName == "" {
			continue
		}

		fieldType := field.Type
		needsEncoding := needsEncoding(fieldType)

		column := ColumnDef{
			Name:          field.Name,
			DBName:        dbName,
			GoType:        goTypeString(fieldType),
			NeedsEncoding: needsEncoding,
		}

		columns = append(columns, column)
	}
	return columns
}

// InferFromType uses reflection to introspect a struct type and infer ColumnDefs.
func InferFromType(typ any, config EntityConfig) EntityDef {
	t := reflect.TypeOf(typ)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	if config.StructType == "" {
		config.StructType = goTypeString(t)
	}
	if config.Name == "" {
		config.Name = t.Name()
	}
	if config.TableName == "" {
		config.TableName = strings.ToLower(config.Name) + "s"
	}

	// Default method names.
	if config.AllMethod == "" {
		config.AllMethod = "All" + config.Name + "s"
	}
	if config.GetMethod == "" {
		config.GetMethod = config.Name
	}
	if config.NewMethod == "" {
		config.NewMethod = "Add" + config.Name
	}
	if config.UpdateMethod == "" {
		config.UpdateMethod = "Update" + config.Name
	}
	if config.DeleteMethod == "" {
		config.DeleteMethod = "Delete" + config.Name
	}

	columns := inferColumns(t)

	// Find primary key - typically the field named ID or with json tag "id"
	pkGoName := ""
	pkColumn := ""
	for _, c := range columns {
		if c.Name == "ID" || c.DBName == "id" {
			pkGoName = c.Name
			pkColumn = c.DBName
			break
		}
	}

	// Find user_id column
	userIDColumn := ""
	for _, c := range columns {
		if c.DBName == "user_id" {
			userIDColumn = c.DBName
			break
		}
	}
	config.HasUserID = userIDColumn != ""

	return EntityDef{
		EntityConfig: config,
		Columns:      columns,
		PKGoName:     pkGoName,
		PKColumn:     pkColumn,
		UserIDColumn: userIDColumn,
	}
}

// EntityRegistry holds all entity definitions with their reflection-derived metadata.
var EntityRegistry = []EntityDef{
	InferFromType(media.Track{}, EntityConfig{
		Name:       "Track",
		StructType: "media.Track",
		TableName:  "tracks",
	}),
	InferFromType(media.Album{}, EntityConfig{
		Name:       "Album",
		StructType: "media.Album",
		TableName:  "albums",
	}),
	InferFromType(media.Video{}, EntityConfig{
		Name:       "Video",
		StructType: "media.Video",
		TableName:  "videos",
	}),
	InferFromType(media.Artist{}, EntityConfig{
		Name:       "Artist",
		StructType: "media.Artist",
		TableName:  "artists",
	}),
	InferFromType(media.Playlist{}, EntityConfig{
		Name:       "Playlist",
		StructType: "media.Playlist",
		TableName:  "playlists",
	}),
	InferFromType(media.DatabaseUser{}, EntityConfig{
		Name:       "User",
		StructType: "media.DatabaseUser",
		TableName:  "users",
		AllMethod:  "Users",
	}),
}

// GetEntity returns an entity definition by name.
func GetEntity(name string) (EntityDef, bool) {
	for _, e := range EntityRegistry {
		if e.Name == name {
			return e, true
		}
	}
	return EntityDef{}, false
}

// AllEntities returns all entity definitions.
func AllEntities() []EntityDef {
	return EntityRegistry
}

// AllColumnNames returns all column database names.
func (e EntityDef) AllColumnNames() []string {
	names := make([]string, len(e.Columns))
	for i, c := range e.Columns {
		names[i] = c.DBName
	}
	return names
}

// ColumnNames returns column database names for given columns.
func ColumnNames(cols []ColumnDef) []string {
	names := make([]string, len(cols))
	for i, c := range cols {
		names[i] = c.DBName
	}
	return names
}

// InsertColumns returns columns for INSERT operations.
func (e EntityDef) InsertColumns() []ColumnDef {
	var cols []ColumnDef
	cols = append(cols, e.Columns...)
	return cols
}

// UpdateColumns returns columns for UPDATE operations (excludes primary key).
func (e EntityDef) UpdateColumns() []ColumnDef {
	var cols []ColumnDef
	for _, c := range e.Columns {
		if c.DBName != e.PKColumn {
			cols = append(cols, c)
		}
	}
	return cols
}

// FieldReferences returns Go field references for pgx Scan (e.g., &item.ID).
func (e EntityDef) FieldReferences() []string {
	refs := make([]string, len(e.Columns))
	for i, c := range e.Columns {
		refs[i] = "&item." + c.Name
	}
	return refs
}

// PrefixedColumnNames returns column names with a table prefix (e.g., u.id, u.username).
func (e EntityDef) PrefixedColumnNames(prefix string) []string {
	names := make([]string, len(e.Columns))
	for i, c := range e.Columns {
		names[i] = prefix + c.DBName
	}
	return names
}
