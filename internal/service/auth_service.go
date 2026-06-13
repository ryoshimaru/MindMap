package service

import (
	"github.com/ryoshimaru/MindMap/internal/domain"
	"github.com/ryoshimaru/MindMap/internal/store"
)

type AuthService struct {
	store store.Store
}

func NewAuthService(store store.Store) *AuthService {
	return &AuthService{store: store}
}

func (s *AuthService) Register(req domain.RegisterRequest) (domain.AuthResponse, error) {
	return s.store.RegisterUser(req)
}

func (s *AuthService) Login(req domain.LoginRequest) (domain.AuthResponse, error) {
	return s.store.Login(req)
}

func (s *AuthService) UserByToken(token string) (domain.User, error) {
	return s.store.UserByToken(token)
}

func (s *AuthService) CreateOAuthUserSession(email, name, avatarURL string, provider domain.AuthProvider) (string, error) {
	return s.store.CreateOAuthUserSession(email, name, avatarURL, provider)
}

func (s *AuthService) GetProfile(userID string) (domain.UserProfile, error) {
	return s.store.GetUserProfile(userID)
}

func (s *AuthService) SaveProfile(userID string, req domain.UpsertUserProfileRequest) (domain.UserProfile, error) {
	return s.store.SaveUserProfile(userID, req)
}
