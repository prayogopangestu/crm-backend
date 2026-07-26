package memory

import (
	"context"
	"sync"

	"github.com/prayogopangestu/crm-system/backend/internal/domain/entities"
)

type UserRepository struct {
	mu    sync.RWMutex
	users map[string]*entities.User
}

func NewUserRepository() *UserRepository {
	return &UserRepository{users: make(map[string]*entities.User)}
}

func (r *UserRepository) Create(ctx context.Context, user *entities.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[user.ID] = user
	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*entities.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if u, ok := r.users[id]; ok {
		return u, nil
	}
	return nil, nil
}
