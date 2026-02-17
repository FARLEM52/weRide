package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"time"

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
	return Repository{
		db:       db,
		tokenTTL: tokenTTL,
		Secret:   secret,
	}
}

func (r *Repository) SaveUser(
	ctx context.Context,
	email string,
	password string,
	firstName string,
	lastName string,
	gender int64) (string, error) {
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
	err := row.Scan(&user.UserID, &user.Email, &user.PassHash, &user.FirsName, &user.LastName, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", errors.New("user not found")
		}
		return "", fmt.Errorf("error saving user: %w", err)
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

func (r *Repository) GetUserRoutes(ctx context.Context, userID uuid.UUID) ([]pb.Route, error) {
	query := `
		SELECT r.route_id, r.driver_id, r.total_price, r.start_point, r.end_point, r.distance
		FROM room_passengers rp
		JOIN routes r ON rp.route_id = r.route_id
		WHERE rp.user_id = $1
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var routes []pb.Route
	for rows.Next() {
		var route pb.Route
		var totalPrice, distance float64
		err := rows.Scan(&route.RouteId, &route.DriverId, &totalPrice, &route.StartPoint, &route.EndPoint, &distance)
		if err != nil {
			return nil, err
		}
		route.TotalPrice = fmt.Sprintf("%.2f", totalPrice)
		route.Distance = fmt.Sprintf("%.2f", distance)
		routes = append(routes, route)
	}

	return routes, nil
}

func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
