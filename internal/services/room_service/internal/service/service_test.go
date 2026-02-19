package service

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	paymentpb "we_ride/internal/services/payment_service/pb"
	roomrepo "we_ride/internal/services/room_service/internal/repository"
	roompb "we_ride/internal/services/room_service/pb"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type fakeRoomRepo struct {
	rooms   map[string]*roompb.Room
	members map[string][]string
	mu      sync.Mutex
}

func newFakeRoomRepo() *fakeRoomRepo {
	return &fakeRoomRepo{rooms: map[string]*roompb.Room{}, members: map[string][]string{}}
}

func (f *fakeRoomRepo) CreateRoom(_ context.Context, room *roompb.Room) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.rooms[room.RoomId] = room
	return nil
}
func (f *fakeRoomRepo) AddMember(_ context.Context, roomID, userID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.members[roomID] = append(f.members[roomID], userID)
	return nil
}
func (f *fakeRoomRepo) RemoveMember(_ context.Context, roomID, userID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	members := f.members[roomID]
	for i, m := range members {
		if m == userID {
			f.members[roomID] = append(members[:i], members[i+1:]...)
			break
		}
	}
	return nil
}
func (f *fakeRoomRepo) GetRoomByID(_ context.Context, roomID string) (*roompb.Room, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	room, ok := f.rooms[roomID]
	if !ok {
		return nil, errors.New("not found")
	}
	return room, nil
}
func (f *fakeRoomRepo) ListAvailableRooms(_ context.Context) ([]*roompb.Room, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []*roompb.Room
	for _, room := range f.rooms {
		if room.Status == roompb.RoomStatus_ROOM_STATUS_WAITING {
			out = append(out, room)
		}
	}
	return out, nil
}
func (f *fakeRoomRepo) UpdateRoomStatus(_ context.Context, roomID string, status roompb.RoomStatus) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.rooms[roomID].Status = status
	return nil
}
func (f *fakeRoomRepo) GetRoomMembers(_ context.Context, roomID string) ([]string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	members := make([]string, len(f.members[roomID]))
	copy(members, f.members[roomID])
	return members, nil
}
func (f *fakeRoomRepo) CompleteRoom(_ context.Context, roomID string, totalPrice, costPerMember float32) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.rooms[roomID].Status = roompb.RoomStatus_ROOM_STATUS_COMPLETED
	f.rooms[roomID].TotalPrice = totalPrice
	f.rooms[roomID].CostPerMember = costPerMember
	return nil
}

var _ roomrepo.Repository = (*fakeRoomRepo)(nil)

func TestCreateRoomAndJoinFlow(t *testing.T) {
	repo := newFakeRoomRepo()
	svc := New(repo, "", "")

	createResp, err := svc.CreateRoom(context.Background(), &roompb.CreateRoomRequest{
		CreatorId:     "driver-1",
		MaxMembers:    2,
		StartLocation: &roompb.Location{Address: "A"},
		EndLocation:   &roompb.Location{Address: "B"},
		ScheduledTime: timestamppb.New(time.Now()),
	})
	if err != nil {
		t.Fatalf("create room error: %v", err)
	}

	_, err = svc.JoinRoom(context.Background(), &roompb.JoinRoomRequest{RoomId: createResp.Room.RoomId, UserId: "passenger-1"})
	if err != nil {
		t.Fatalf("join room error: %v", err)
	}

	// third member should fail as room gets full (2 seats total)
	_, err = svc.JoinRoom(context.Background(), &roompb.JoinRoomRequest{RoomId: createResp.Room.RoomId, UserId: "passenger-2"})
	if err == nil {
		t.Fatal("expected room full error")
	}
}

func TestCompleteRideSuccess(t *testing.T) {
	repo := newFakeRoomRepo()
	room := &roompb.Room{
		RoomId:         "room-1",
		CreatorId:      "driver-1",
		AvailableSeats: 3,
		Status:         roompb.RoomStatus_ROOM_STATUS_WAITING,
		StartLocation:  &roompb.Location{Address: "Start"},
		EndLocation:    &roompb.Location{Address: "End"},
	}
	repo.rooms[room.RoomId] = room
	repo.members[room.RoomId] = []string{"driver-1", "u2", "u3"}

	svc := New(repo, "", "")
	var paymentReq *paymentpb.ProcessPaymentRequest
	var routeSaved bool
	svc.processPaymentFn = func(_ context.Context, req *paymentpb.ProcessPaymentRequest) (*paymentpb.ProcessPaymentResponse, error) {
		paymentReq = req
		return &paymentpb.ProcessPaymentResponse{Success: true, Payments: []*paymentpb.Payment{{PaymentId: "p1"}, {PaymentId: "p2"}, {PaymentId: "p3"}}}, nil
	}
	svc.saveRouteFn = func(_ *roompb.CompleteRideRequest, _ []string, _, _ string, _ float32) {
		routeSaved = true
	}

	resp, err := svc.CompleteRide(context.Background(), &roompb.CompleteRideRequest{RoomId: "room-1", DriverId: "driver-1", TotalPrice: 900, DistanceKm: 15})
	if err != nil {
		t.Fatalf("complete ride error: %v", err)
	}
	if !resp.Success || resp.PaymentsCount != 3 {
		t.Fatalf("unexpected complete ride response: %+v", resp)
	}
	if paymentReq == nil || paymentReq.AmountPerUser != 300 {
		t.Fatalf("unexpected payment request: %+v", paymentReq)
	}
	if !routeSaved {
		t.Fatal("expected route to be saved")
	}
}

func TestCompleteRideNoMembers(t *testing.T) {
	repo := newFakeRoomRepo()
	repo.rooms["room-1"] = &roompb.Room{RoomId: "room-1", Status: roompb.RoomStatus_ROOM_STATUS_WAITING}
	repo.members["room-1"] = []string{}

	svc := New(repo, "", "")
	_, err := svc.CompleteRide(context.Background(), &roompb.CompleteRideRequest{RoomId: "room-1", DriverId: "driver-1", TotalPrice: 100})
	if err == nil {
		t.Fatal("expected no members error")
	}
}
