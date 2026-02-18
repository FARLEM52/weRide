package repository

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
	roomservice "we_ride/internal/services/room_service/pb"
)

type Repository interface {
	CreateRoom(ctx context.Context, room *roomservice.Room) error
	AddMember(ctx context.Context, roomID, userID string) error
	RemoveMember(ctx context.Context, roomID, userID string) error
	GetRoomByID(ctx context.Context, roomID string) (*roomservice.Room, error)
	ListAvailableRooms(ctx context.Context) ([]*roomservice.Room, error)
	UpdateRoomStatus(ctx context.Context, roomID string, status roomservice.RoomStatus) error
	GetRoomMembers(ctx context.Context, roomID string) ([]string, error)
	CompleteRoom(ctx context.Context, roomID string, totalPrice, costPerMember float32) error
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

// CreateRoom вставляет новую запись в таблицу rooms
func (r *repository) CreateRoom(ctx context.Context, room *roomservice.Room) error {
	query := `
	INSERT INTO rooms (
		room_id, creator_id, start_latitude, start_longitude, end_latitude, end_longitude,
		available_seats, status, created_at, scheduled_time, total_price, cost_per_member
	)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12);
	`

	_, err := r.db.Exec(ctx, query,
		room.RoomId,
		room.CreatorId,
		room.StartLocation.Latitude,
		room.StartLocation.Longitude,
		room.EndLocation.Latitude,
		room.EndLocation.Longitude,
		room.AvailableSeats,
		room.Status,
		room.CreatedAt.AsTime(),
		room.ScheduledTime.AsTime(),
		room.TotalPrice,
		room.CostPerMember,
	)
	return err
}

func (r *repository) AddMember(ctx context.Context, roomID, userID string) error {
	query := `INSERT INTO room_members (room_id, user_id) VALUES ($1,$2)
			  ON CONFLICT DO NOTHING;`
	_, err := r.db.Exec(ctx, query, roomID, userID)
	return err
}

func (r *repository) RemoveMember(ctx context.Context, roomID, userID string) error {
	query := `DELETE FROM room_members WHERE room_id=$1 AND user_id=$2;`
	_, err := r.db.Exec(ctx, query, roomID, userID)
	return err
}

func (r *repository) GetRoomByID(ctx context.Context, roomID string) (*roomservice.Room, error) {
	query := `
	SELECT room_id, creator_id, available_seats, status, total_price, cost_per_member, created_at, scheduled_time
	FROM rooms WHERE room_id=$1;
	`

	row := r.db.QueryRow(ctx, query, roomID)

	room := &roomservice.Room{}
	var createdAt, scheduled time.Time
	err := row.Scan(&room.RoomId, &room.CreatorId, &room.AvailableSeats, &room.Status,
		&room.TotalPrice, &room.CostPerMember, &createdAt, &scheduled)
	if err != nil {
		return nil, fmt.Errorf("GetRoomByID: %w", err)
	}

	room.CreatedAt = timestamppb.New(createdAt)
	room.ScheduledTime = timestamppb.New(scheduled)
	return room, nil
}

func (r *repository) ListAvailableRooms(ctx context.Context) ([]*roomservice.Room, error) {
	query := `SELECT room_id, creator_id, available_seats, status FROM rooms WHERE status = 1;`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []*roomservice.Room
	for rows.Next() {
		room := &roomservice.Room{}
		err := rows.Scan(&room.RoomId, &room.CreatorId, &room.AvailableSeats, &room.Status)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}
	return rooms, nil
}

func (r *repository) UpdateRoomStatus(ctx context.Context, roomID string, status roomservice.RoomStatus) error {
	query := `UPDATE rooms SET status=$1 WHERE room_id=$2;`
	_, err := r.db.Exec(ctx, query, status, roomID)
	return err
}

// GetRoomMembers возвращает список участников комнаты
func (r *repository) GetRoomMembers(ctx context.Context, roomID string) ([]string, error) {
	query := `SELECT user_id FROM room_members WHERE room_id = $1`
	rows, err := r.db.Query(ctx, query, roomID)
	if err != nil {
		return nil, fmt.Errorf("GetRoomMembers: %w", err)
	}
	defer rows.Close()

	var members []string
	for rows.Next() {
		var uid string
		if err := rows.Scan(&uid); err != nil {
			return nil, err
		}
		members = append(members, uid)
	}
	return members, nil
}

// CompleteRoom переводит комнату в статус COMPLETED и обновляет цену
func (r *repository) CompleteRoom(ctx context.Context, roomID string, totalPrice, costPerMember float32) error {
	query := `
		UPDATE rooms
		SET status = $1, total_price = $2, cost_per_member = $3
		WHERE room_id = $4
	`
	_, err := r.db.Exec(ctx, query,
		roomservice.RoomStatus_ROOM_STATUS_COMPLETED,
		totalPrice, costPerMember, roomID,
	)
	return err
}
