package postgres

import (
	"context"
	"fmt"

	"github.com/vladislav-koval/url-shortener/internal/shortener/auth/domain"
)

// UpsertUser создаёт пользователя или обновляет профильные поля у уже
// существующего (найден по google_sub — это внешний, стабильный идентификатор
// от Google, не email: email у аккаунта в принципе может смениться).
// ID в EXCLUDED-обновление не попадает — при конфликте вернётся уже
// существующий id, не переданный. ID генерируется на стороне сервиса
// (domain.NewUser), тем же паттерном, что и events.ClickEvent.ID.
func (r *Repository) UpsertUser(ctx context.Context, user domain.User) (domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		INSERT INTO urlshortener.users (id, google_sub, email, email_verified, name)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (google_sub) DO UPDATE SET
			email          = EXCLUDED.email,
			email_verified = EXCLUDED.email_verified,
			name           = EXCLUDED.name,
			updated_at     = now()
		RETURNING id, google_sub, email, email_verified, name, created_at, updated_at;
	`

	row := r.pool.QueryRow(
		ctx,
		query,
		user.ID,
		user.GoogleSub,
		user.Email,
		user.EmailVerified,
		user.Name,
	)

	var model userRow

	err := row.Scan(
		&model.ID,
		&model.GoogleSub,
		&model.Email,
		&model.EmailVerified,
		&model.Name,
		&model.CreatedAt,
		&model.UpdatedAt,
	)
	if err != nil {
		return domain.User{}, fmt.Errorf("upsert user: %w", err)
	}

	return userFromRow(model), nil
}
