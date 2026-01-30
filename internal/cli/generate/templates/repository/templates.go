package repository

// TemplateData contains all data needed for repository template rendering
type TemplateData struct {
	Name      string // Repository name (e.g., "User")
	NameLower string // Lowercase name (e.g., "user")
	Package   string // Package name
	Entity    string // Entity struct name
	Table     string // Database table name
	DB        string // Database type: postgres, mysql, sqlite
	IDType    string // ID type (e.g., "int64", "string", "uuid.UUID")
	Interface bool   // Generate interface
}

// InterfaceTemplate generates the repository interface
const InterfaceTemplate = `package {{.Package}}

import (
	"context"
)

// {{.Name}}Repository defines the interface for {{.NameLower}} data access
type {{.Name}}Repository interface {
	// Create creates a new {{.NameLower}}
	Create(ctx context.Context, entity *{{.Entity}}) error

	// GetByID retrieves a {{.NameLower}} by ID
	GetByID(ctx context.Context, id {{.IDType}}) (*{{.Entity}}, error)

	// Update updates an existing {{.NameLower}}
	Update(ctx context.Context, entity *{{.Entity}}) error

	// Delete removes a {{.NameLower}} by ID
	Delete(ctx context.Context, id {{.IDType}}) error

	// List retrieves {{.NameLower}}s with pagination
	List(ctx context.Context, limit, offset int) ([]*{{.Entity}}, error)

	// Count returns the total number of {{.NameLower}}s
	Count(ctx context.Context) (int64, error)
}
`

// PostgresRepositoryTemplate generates a PostgreSQL repository
const PostgresRepositoryTemplate = `package {{.Package}}

import (
	"context"
	"database/sql"
	"fmt"
)

// {{.Name}}RepositoryImpl implements {{.Name}}Repository for PostgreSQL
type {{.Name}}RepositoryImpl struct {
	db *sql.DB
}

// New{{.Name}}Repository creates a new {{.Name}}RepositoryImpl
func New{{.Name}}Repository(db *sql.DB) *{{.Name}}RepositoryImpl {
	return &{{.Name}}RepositoryImpl{db: db}
}

// Create creates a new {{.NameLower}}
func (r *{{.Name}}RepositoryImpl) Create(ctx context.Context, entity *{{.Entity}}) error {
	query := ` + "`" + `
		INSERT INTO {{.Table}} (/* columns */)
		VALUES (/* values */)
		RETURNING id
	` + "`" + `

	return r.db.QueryRowContext(ctx, query /* args */).Scan(&entity.ID)
}

// GetByID retrieves a {{.NameLower}} by ID
func (r *{{.Name}}RepositoryImpl) GetByID(ctx context.Context, id {{.IDType}}) (*{{.Entity}}, error) {
	query := ` + "`" + `
		SELECT /* columns */
		FROM {{.Table}}
		WHERE id = $1
	` + "`" + `

	entity := &{{.Entity}}{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(/* fields */)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("{{.NameLower}} not found: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get {{.NameLower}}: %w", err)
	}

	return entity, nil
}

// Update updates an existing {{.NameLower}}
func (r *{{.Name}}RepositoryImpl) Update(ctx context.Context, entity *{{.Entity}}) error {
	query := ` + "`" + `
		UPDATE {{.Table}}
		SET /* columns = values */
		WHERE id = $1
	` + "`" + `

	result, err := r.db.ExecContext(ctx, query, entity.ID /* other args */)
	if err != nil {
		return fmt.Errorf("failed to update {{.NameLower}}: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("{{.NameLower}} not found")
	}

	return nil
}

// Delete removes a {{.NameLower}} by ID
func (r *{{.Name}}RepositoryImpl) Delete(ctx context.Context, id {{.IDType}}) error {
	query := ` + "`" + `DELETE FROM {{.Table}} WHERE id = $1` + "`" + `

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete {{.NameLower}}: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("{{.NameLower}} not found")
	}

	return nil
}

// List retrieves {{.NameLower}}s with pagination
func (r *{{.Name}}RepositoryImpl) List(ctx context.Context, limit, offset int) ([]*{{.Entity}}, error) {
	query := ` + "`" + `
		SELECT /* columns */
		FROM {{.Table}}
		ORDER BY id
		LIMIT $1 OFFSET $2
	` + "`" + `

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list {{.NameLower}}s: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var entities []*{{.Entity}}
	for rows.Next() {
		entity := &{{.Entity}}{}
		if err := rows.Scan(/* fields */); err != nil {
			return nil, fmt.Errorf("failed to scan {{.NameLower}}: %w", err)
		}
		entities = append(entities, entity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating {{.NameLower}}s: %w", err)
	}

	return entities, nil
}

// Count returns the total number of {{.NameLower}}s
func (r *{{.Name}}RepositoryImpl) Count(ctx context.Context) (int64, error) {
	query := ` + "`" + `SELECT COUNT(*) FROM {{.Table}}` + "`" + `

	var count int64
	if err := r.db.QueryRowContext(ctx, query).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count {{.NameLower}}s: %w", err)
	}

	return count, nil
}
`

// MySQLRepositoryTemplate generates a MySQL repository
const MySQLRepositoryTemplate = `package {{.Package}}

import (
	"context"
	"database/sql"
	"fmt"
)

// {{.Name}}RepositoryImpl implements {{.Name}}Repository for MySQL
type {{.Name}}RepositoryImpl struct {
	db *sql.DB
}

// New{{.Name}}Repository creates a new {{.Name}}RepositoryImpl
func New{{.Name}}Repository(db *sql.DB) *{{.Name}}RepositoryImpl {
	return &{{.Name}}RepositoryImpl{db: db}
}

// Create creates a new {{.NameLower}}
func (r *{{.Name}}RepositoryImpl) Create(ctx context.Context, entity *{{.Entity}}) error {
	query := ` + "`" + `
		INSERT INTO {{.Table}} (/* columns */)
		VALUES (/* values */)
	` + "`" + `

	result, err := r.db.ExecContext(ctx, query /* args */)
	if err != nil {
		return fmt.Errorf("failed to create {{.NameLower}}: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	entity.ID = id
	return nil
}

// GetByID retrieves a {{.NameLower}} by ID
func (r *{{.Name}}RepositoryImpl) GetByID(ctx context.Context, id {{.IDType}}) (*{{.Entity}}, error) {
	query := ` + "`" + `
		SELECT /* columns */
		FROM {{.Table}}
		WHERE id = ?
	` + "`" + `

	entity := &{{.Entity}}{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(/* fields */)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("{{.NameLower}} not found: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get {{.NameLower}}: %w", err)
	}

	return entity, nil
}

// Update updates an existing {{.NameLower}}
func (r *{{.Name}}RepositoryImpl) Update(ctx context.Context, entity *{{.Entity}}) error {
	query := ` + "`" + `
		UPDATE {{.Table}}
		SET /* columns = values */
		WHERE id = ?
	` + "`" + `

	result, err := r.db.ExecContext(ctx, query /* args */, entity.ID)
	if err != nil {
		return fmt.Errorf("failed to update {{.NameLower}}: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("{{.NameLower}} not found")
	}

	return nil
}

// Delete removes a {{.NameLower}} by ID
func (r *{{.Name}}RepositoryImpl) Delete(ctx context.Context, id {{.IDType}}) error {
	query := ` + "`" + `DELETE FROM {{.Table}} WHERE id = ?` + "`" + `

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete {{.NameLower}}: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("{{.NameLower}} not found")
	}

	return nil
}

// List retrieves {{.NameLower}}s with pagination
func (r *{{.Name}}RepositoryImpl) List(ctx context.Context, limit, offset int) ([]*{{.Entity}}, error) {
	query := ` + "`" + `
		SELECT /* columns */
		FROM {{.Table}}
		ORDER BY id
		LIMIT ? OFFSET ?
	` + "`" + `

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list {{.NameLower}}s: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var entities []*{{.Entity}}
	for rows.Next() {
		entity := &{{.Entity}}{}
		if err := rows.Scan(/* fields */); err != nil {
			return nil, fmt.Errorf("failed to scan {{.NameLower}}: %w", err)
		}
		entities = append(entities, entity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating {{.NameLower}}s: %w", err)
	}

	return entities, nil
}

// Count returns the total number of {{.NameLower}}s
func (r *{{.Name}}RepositoryImpl) Count(ctx context.Context) (int64, error) {
	query := ` + "`" + `SELECT COUNT(*) FROM {{.Table}}` + "`" + `

	var count int64
	if err := r.db.QueryRowContext(ctx, query).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count {{.NameLower}}s: %w", err)
	}

	return count, nil
}
`

// SQLiteRepositoryTemplate generates a SQLite repository
const SQLiteRepositoryTemplate = `package {{.Package}}

import (
	"context"
	"database/sql"
	"fmt"
)

// {{.Name}}RepositoryImpl implements {{.Name}}Repository for SQLite
type {{.Name}}RepositoryImpl struct {
	db *sql.DB
}

// New{{.Name}}Repository creates a new {{.Name}}RepositoryImpl
func New{{.Name}}Repository(db *sql.DB) *{{.Name}}RepositoryImpl {
	return &{{.Name}}RepositoryImpl{db: db}
}

// Create creates a new {{.NameLower}}
func (r *{{.Name}}RepositoryImpl) Create(ctx context.Context, entity *{{.Entity}}) error {
	query := ` + "`" + `
		INSERT INTO {{.Table}} (/* columns */)
		VALUES (/* values */)
	` + "`" + `

	result, err := r.db.ExecContext(ctx, query /* args */)
	if err != nil {
		return fmt.Errorf("failed to create {{.NameLower}}: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	entity.ID = id
	return nil
}

// GetByID retrieves a {{.NameLower}} by ID
func (r *{{.Name}}RepositoryImpl) GetByID(ctx context.Context, id {{.IDType}}) (*{{.Entity}}, error) {
	query := ` + "`" + `
		SELECT /* columns */
		FROM {{.Table}}
		WHERE id = ?
	` + "`" + `

	entity := &{{.Entity}}{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(/* fields */)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("{{.NameLower}} not found: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get {{.NameLower}}: %w", err)
	}

	return entity, nil
}

// Update updates an existing {{.NameLower}}
func (r *{{.Name}}RepositoryImpl) Update(ctx context.Context, entity *{{.Entity}}) error {
	query := ` + "`" + `
		UPDATE {{.Table}}
		SET /* columns = values */
		WHERE id = ?
	` + "`" + `

	result, err := r.db.ExecContext(ctx, query /* args */, entity.ID)
	if err != nil {
		return fmt.Errorf("failed to update {{.NameLower}}: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("{{.NameLower}} not found")
	}

	return nil
}

// Delete removes a {{.NameLower}} by ID
func (r *{{.Name}}RepositoryImpl) Delete(ctx context.Context, id {{.IDType}}) error {
	query := ` + "`" + `DELETE FROM {{.Table}} WHERE id = ?` + "`" + `

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete {{.NameLower}}: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("{{.NameLower}} not found")
	}

	return nil
}

// List retrieves {{.NameLower}}s with pagination
func (r *{{.Name}}RepositoryImpl) List(ctx context.Context, limit, offset int) ([]*{{.Entity}}, error) {
	query := ` + "`" + `
		SELECT /* columns */
		FROM {{.Table}}
		ORDER BY id
		LIMIT ? OFFSET ?
	` + "`" + `

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list {{.NameLower}}s: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var entities []*{{.Entity}}
	for rows.Next() {
		entity := &{{.Entity}}{}
		if err := rows.Scan(/* fields */); err != nil {
			return nil, fmt.Errorf("failed to scan {{.NameLower}}: %w", err)
		}
		entities = append(entities, entity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating {{.NameLower}}s: %w", err)
	}

	return entities, nil
}

// Count returns the total number of {{.NameLower}}s
func (r *{{.Name}}RepositoryImpl) Count(ctx context.Context) (int64, error) {
	query := ` + "`" + `SELECT COUNT(*) FROM {{.Table}}` + "`" + `

	var count int64
	if err := r.db.QueryRowContext(ctx, query).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count {{.NameLower}}s: %w", err)
	}

	return count, nil
}
`

// RepositoryTestTemplate generates repository tests
const RepositoryTestTemplate = `package {{.Package}}

import (
	"context"
	"testing"
)

func TestNew{{.Name}}Repository(t *testing.T) {
	// TODO: Set up test database
	// db, err := sql.Open("{{.DB}}", "test connection string")
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// defer db.Close()

	// repo := New{{.Name}}Repository(db)
	// if repo == nil {
	// 	t.Fatal("New{{.Name}}Repository() returned nil")
	// }
	t.Skip("TODO: Implement with test database")
}

func Test{{.Name}}Repository_Create(t *testing.T) {
	t.Skip("TODO: Implement with test database")
}

func Test{{.Name}}Repository_GetByID(t *testing.T) {
	t.Skip("TODO: Implement with test database")
}

func Test{{.Name}}Repository_Update(t *testing.T) {
	t.Skip("TODO: Implement with test database")
}

func Test{{.Name}}Repository_Delete(t *testing.T) {
	t.Skip("TODO: Implement with test database")
}

func Test{{.Name}}Repository_List(t *testing.T) {
	t.Skip("TODO: Implement with test database")
}

func Test{{.Name}}Repository_Count(t *testing.T) {
	ctx := context.Background()
	_ = ctx
	t.Skip("TODO: Implement with test database")
}
`
