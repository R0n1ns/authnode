package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"

	"authmicro/internal/domain"
)

type RoleRepository struct {
	db *sqlx.DB
}

func NewRoleRepository(db *sqlx.DB) *RoleRepository {
	return &RoleRepository{
		db: db,
	}
}

// CreateRole creates a new role
func (r *RoleRepository) CreateRole(ctx context.Context, name string) (int64, error) {
	query := `
                INSERT INTO roles (name, created_at, updated_at)
                VALUES ($1, $2, $3)
                RETURNING id`

	var id int64
	now := time.Now().UTC()

	err := r.db.QueryRowContext(ctx, query, name, now, now).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// GetRoleByName retrieves a role by name
func (r *RoleRepository) GetRoleByName(ctx context.Context, name string) (domain.Role, error) {
	query := `
                SELECT id, name, created_at, updated_at
                FROM roles
                WHERE name = $1`

	var role domain.Role
	err := r.db.GetContext(ctx, &role, query, name)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Role{}, errors.New("role not found")
		}
		return domain.Role{}, err
	}

	return role, nil
}

// GetRoleByID retrieves a role by ID
func (r *RoleRepository) GetRoleByID(ctx context.Context, id int64) (domain.Role, error) {
	query := `
                SELECT id, name, created_at, updated_at
                FROM roles
                WHERE id = $1`

	var role domain.Role
	err := r.db.GetContext(ctx, &role, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Role{}, errors.New("role not found")
		}
		return domain.Role{}, err
	}

	return role, nil
}

// AssignRoleToUser assigns a role to a user
func (r *RoleRepository) AssignRoleToUser(ctx context.Context, userID, roleID int64) error {
	query := `
                INSERT INTO user_roles (user_id, role_id, created_at)
                VALUES ($1, $2, $3)`

	_, err := r.db.ExecContext(ctx, query, userID, roleID, time.Now().UTC())
	return err
}

// GetUserRoles retrieves all roles assigned to a user
func (r *RoleRepository) GetUserRoles(ctx context.Context, userID int64) ([]domain.Role, error) {
	query := `
                SELECT r.id, r.name, r.created_at, r.updated_at
                FROM roles r
                JOIN user_roles ur ON r.id = ur.role_id
                WHERE ur.user_id = $1`

	var roles []domain.Role
	err := r.db.SelectContext(ctx, &roles, query, userID)
	if err != nil {
		return nil, err
	}

	return roles, nil
}

// GetUserRoleNames retrieves all role names assigned to a user
func (r *RoleRepository) GetUserRoleNames(ctx context.Context, userID int64) ([]string, error) {
	query := `
                SELECT r.name
                FROM roles r
                JOIN user_roles ur ON r.id = ur.role_id
                WHERE ur.user_id = $1`

	var roleNames []string
	err := r.db.SelectContext(ctx, &roleNames, query, userID)
	if err != nil {
		return nil, err
	}

	return roleNames, nil
}

// RemoveRoleFromUser removes a role from a user
func (r *RoleRepository) RemoveRoleFromUser(ctx context.Context, userID, roleID int64) error {
	query := `
                DELETE FROM user_roles
                WHERE user_id = $1 AND role_id = $2`

	_, err := r.db.ExecContext(ctx, query, userID, roleID)
	return err
}

// HasRole checks if a user has a specific role
func (r *RoleRepository) HasRole(ctx context.Context, userID int64, roleName string) (bool, error) {
	query := `
                SELECT EXISTS (
                        SELECT 1
                        FROM user_roles ur
                        JOIN roles r ON ur.role_id = r.id
                        WHERE ur.user_id = $1 AND r.name = $2
                )`

	var hasRole bool
	err := r.db.GetContext(ctx, &hasRole, query, userID, roleName)
	if err != nil {
		return false, err
	}

	return hasRole, nil
}
