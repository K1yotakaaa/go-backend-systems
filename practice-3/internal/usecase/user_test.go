package usecase

import (
	"errors"
	postgresUsers "practice-3/internal/repository/_postgres/users"
	"testing"

	"practice-3/pkg/modules"
)

type mockRepo struct {
	getByIDErr error
}

func (m *mockRepo) GetUsers(limit, offset int) ([]modules.User, error) {
	return []modules.User{{ID: 1, Name: "A", Email: "a@a.com", Age: 20}}, nil
}

func (m *mockRepo) GetUserByID(id int) (*modules.User, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	return &modules.User{ID: id, Name: "A", Email: "a@a.com", Age: 20}, nil
}

func (m *mockRepo) CreateUser(u modules.User) (int, error) { return 7, nil }
func (m *mockRepo) UpdateUser(id int, u modules.User) error { return nil }
func (m *mockRepo) DeleteUser(id int) (int64, error) { return 1, nil }

func TestGetUsers_OK(t *testing.T) {
	uc := NewUserUsecase(&mockRepo{})

	us, err := uc.GetUsers(10, 0)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(us) != 1 {
		t.Fatalf("expected 1 user, got %d", len(us))
	}
	if us[0].ID != 1 {
		t.Fatalf("expected id=1, got %d", us[0].ID)
	}
}

func TestGetUserByID_PropagatesError(t *testing.T) {
	wantErr := errors.New("db down")
	uc := NewUserUsecase(&mockRepo{getByIDErr: wantErr})

	_, err := uc.GetUserByID(123)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetUserByID_NotFound(t *testing.T) {
  uc := NewUserUsecase(&mockRepo{getByIDErr: postgresUsers.ErrUserNotFound})
	
  _, err := uc.GetUserByID(999)
  if err == nil {
    t.Fatal("expected error, got nil")
  }
  if !IsNotFound(err) {
    t.Fatalf("expected IsNotFound(err)=true, got false. err=%v", err)
  }
}

