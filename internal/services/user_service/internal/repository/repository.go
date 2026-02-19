package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"we_ride/internal/services/user_service/internal/jwt"
	"we_ride/internal/services/user_service/internal/models"
	pb "we_ride/internal/services/user_service/protoc/gen/go"
)

type Repository struct {
	db       *pgxpool.Pool
	tokenTTL time.Duration
	Secret   string
}

func NewRepository(db *pgxpool.Pool, tokenTTL time.Duration, secret string) Repository {
	return Repository{db: db, tokenTTL: tokenTTL, Secret: secret}
}

func (r *Repository) SaveUser(ctx context.Context, email, password, firstName, lastName string, gender int64) (string, error) {
	id := uuid.New().String()
	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("error hashing password: %v", err)
	}
	query := `
		INSERT INTO public.users (user_id, email, password_hash, first_name, last_name, gender, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err = r.db.Exec(ctx, query, id, email, passHash, firstName, lastName, gender, time.Now())
	if err != nil {
		if IsUniqueViolation(err) {
			return "", errors.New("user with this email already exists")
		}
		return "", err
	}
	return id, nil
}

func (r *Repository) LoginUser(ctx context.Context, email, password string) (string, error) {
	query := `
		SELECT user_id, email, password_hash, first_name, last_name, created_at
		FROM public.users
		WHERE email = $1
	`
	row := r.db.QueryRow(ctx, query, email)
	var user models.User
	err := row.Scan(&user.UserID, &user.Email, &user.PassHash, &user.FirstName, &user.LastName, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", errors.New("user not found")
		}
		return "", fmt.Errorf("error querying user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		return "", fmt.Errorf("error checking password: %v", err)
	}

	token, err := jwt.NewToken(user, r.Secret, r.tokenTTL)
	if err != nil {
		return "", fmt.Errorf("error creating token: %v", err)
	}
	return token, nil
}

// GetUserRoutes возвращает историю поездок пользователя.
// Работает через таблицы public.routes и public.room_passengers.
func (r *Repository) GetUserRoutes(ctx context.Context, userID uuid.UUID) ([]*pb.Route, error) {
	query := `
		SELECT
			rt.route_id,
			rt.driver_id,
			rt.total_price::text,
			rt.start_point,
			rt.end_point,
			rt.distance::text
		FROM public.room_passengers rp
		JOIN public.routes rt ON rp.route_id = rt.route_id
		WHERE rp.user_id = $1
		ORDER BY rt.completed_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("GetUserRoutes query: %w", err)
	}
	defer rows.Close()

	var routes []*pb.Route
	for rows.Next() {
		route := &pb.Route{}
		if err := rows.Scan(
			&route.RouteId,
			&route.DriverId,
			&route.TotalPrice,
			&route.StartPoint,
			&route.EndPoint,
			&route.Distance,
		); err != nil {
			return nil, fmt.Errorf("GetUserRoutes scan: %w", err)
		}
		routes = append(routes, route)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetUserRoutes rows: %w", err)
	}
	return routes, nil
}

// SaveRoute сохраняет завершённую поездку и список пассажиров.
// Вызывается из room_service при переводе комнаты в COMPLETED.
func (r *Repository) SaveRoute(ctx context.Context, roomID, driverID, startPoint, endPoint string, distance, totalPrice float64, passengerIDs []string) (string, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	routeID := uuid.New().String()
	_, err = tx.Exec(ctx, `
		INSERT INTO public.routes (route_id, room_id, driver_id, start_point, end_point, distance, total_price)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (room_id) DO NOTHING
	`, routeID, roomID, driverID, startPoint, endPoint, distance, totalPrice)
	if err != nil {
		return "", fmt.Errorf("insert route: %w", err)
	}

	for _, pid := range passengerIDs {
		_, err = tx.Exec(ctx, `
			INSERT INTO public.room_passengers (route_id, user_id)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, routeID, pid)
		if err != nil {
			return "", fmt.Errorf("insert passenger %s: %w", pid, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("commit tx: %w", err)
	}
	return routeID, nil
}

func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
