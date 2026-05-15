package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/salary-manager/backend/internal/model"
	"github.com/salary-manager/backend/internal/repository"
)

type countryRepo struct {
	db *sql.DB
}

func NewCountryRepository(db *sql.DB) repository.CountryRepository {
	return &countryRepo{db: db}
}

func (r *countryRepo) Create(ctx context.Context, c *model.Country) error {
	now := time.Now().UTC()
	c.CreatedAt = now
	c.UpdatedAt = now
	c.IsActive = true
	c.Code = strings.ToUpper(c.Code)
	c.Currency = strings.ToUpper(c.Currency)

	res, err := r.db.ExecContext(ctx, `
		INSERT INTO countries (name, code, currency, is_active, created_at, updated_at)
		VALUES (?, ?, ?, 1, ?, ?)`,
		c.Name, c.Code, c.Currency, c.CreatedAt, c.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert country: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("last insert id: %w", err)
	}
	c.ID = id
	return nil
}

func (r *countryRepo) List(ctx context.Context, req model.CountryListRequest) (*model.CountryListResult, error) {
	page := req.Pagination.Normalized()

	where := ""
	var args []interface{}
	if !req.IncludeInactive {
		where = " WHERE is_active = 1"
	}

	var total int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM countries`+where, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count countries: %w", err)
	}

	query := `SELECT id, name, code, currency, is_active, created_at, updated_at
	          FROM countries` + where + ` ORDER BY name`
	query, args = applyPagination(query, args, page)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list countries: %w", err)
	}
	defer rows.Close()

	out := []model.Country{}
	for rows.Next() {
		var c model.Country
		if err := rows.Scan(&c.ID, &c.Name, &c.Code, &c.Currency, &c.IsActive, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan country: %w", err)
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}

	return &model.CountryListResult{
		Countries: out,
		Total:     total,
		Limit:     page.Limit,
		Offset:    page.Offset,
	}, nil
}

func (r *countryRepo) GetByID(ctx context.Context, id int64) (*model.Country, error) {
	c := &model.Country{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, code, currency, is_active, created_at, updated_at
		FROM countries WHERE id = ?`, id,
	).Scan(&c.ID, &c.Name, &c.Code, &c.Currency, &c.IsActive, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get country: %w", err)
	}
	return c, nil
}

func (r *countryRepo) GetByCode(ctx context.Context, code string) (*model.Country, error) {
	c := &model.Country{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, code, currency, is_active, created_at, updated_at
		FROM countries WHERE code = ?`, strings.ToUpper(code),
	).Scan(&c.ID, &c.Name, &c.Code, &c.Currency, &c.IsActive, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get country by code: %w", err)
	}
	return c, nil
}
