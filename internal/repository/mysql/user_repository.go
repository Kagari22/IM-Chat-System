package mysql

import (
	"context"
	"database/sql"
	"errors"

	"IM_Chat_System/internal/model"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, username, passwordHash, nickname string) (model.User, error) {
	result, err := r.db.ExecContext(
		ctx,
		`INSERT INTO users (username, password_hash, nickname) VALUES (?, ?, ?)`,
		username,
		passwordHash,
		nickname,
	)
	if err != nil {
		return model.User{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return model.User{}, err
	}

	user, ok, err := r.GetByID(ctx, id)
	if err != nil {
		return model.User{}, err
	}
	if !ok {
		return model.User{}, errors.New("user inserted but not found")
	}
	return user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (model.User, bool, error) {
	var user model.User
	err := r.db.QueryRowContext(
		ctx,
		`SELECT id, username, password_hash, nickname, created_at FROM users WHERE id = ?`,
		id,
	).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Nickname, &user.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return model.User{}, false, nil
	}
	if err != nil {
		return model.User{}, false, err
	}
	return user, true, nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (model.User, bool, error) {
	var user model.User
	err := r.db.QueryRowContext(
		ctx,
		`SELECT id, username, password_hash, nickname, created_at FROM users WHERE username = ?`,
		username,
	).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Nickname, &user.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return model.User{}, false, nil
	}
	if err != nil {
		return model.User{}, false, err
	}
	return user, true, nil
}

func (r *UserRepository) List(ctx context.Context, excludeUserID int64) ([]model.User, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, username, password_hash, nickname, created_at FROM users WHERE id <> ? ORDER BY id ASC`,
		excludeUserID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var user model.User
		if err := rows.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Nickname, &user.CreatedAt); err != nil {
			return nil, err
		}
		user.PasswordHash = ""
		users = append(users, user)
	}
	return users, rows.Err()
}
