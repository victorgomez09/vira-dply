package service

import (
	"context"
	"fmt"

	"github.com/mikrocloud/mikrocloud/internal/domain/users"
	"github.com/mikrocloud/mikrocloud/internal/domain/users/repository"
)

type UserService struct {
	userRepo repository.Repository
}

func NewUserService(userRepo repository.Repository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

type CreateUserCommand struct {
	Email        string
	PasswordHash string
	Username     string
}

type CreateOrganizationCommand struct {
	Name         string
	Slug         string
	Description  string
	OwnerID      string
	BillingEmail string
}

func (s *UserService) CreateUser(ctx context.Context, cmd CreateUserCommand) (*users.User, error) {
	email, err := users.NewEmail(cmd.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid email: %w", err)
	}

	exists, err := s.userRepo.Exists(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("user with email '%s' already exists", cmd.Email)
	}

	user := users.NewUser(email, cmd.PasswordHash)

	if cmd.Username != "" {
		username, err := users.NewUsername(cmd.Username)
		if err != nil {
			return nil, fmt.Errorf("invalid username: %w", err)
		}
		user.SetUsername(username)
	}

	if err := s.userRepo.Save(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	return user, nil
}

func (s *UserService) CreateOrganization(ctx context.Context, cmd CreateOrganizationCommand) (*users.Organization, error) {
	exists, err := s.userRepo.OrganizationExists(ctx, cmd.Slug)
	if err != nil {
		return nil, fmt.Errorf("failed to check organization existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("organization with slug '%s' already exists", cmd.Slug)
	}

	ownerID, err := users.UserIDFromString(cmd.OwnerID)
	if err != nil {
		return nil, fmt.Errorf("invalid owner ID: %w", err)
	}

	_, err = s.userRepo.FindByID(ctx, ownerID)
	if err != nil {
		return nil, fmt.Errorf("owner user not found: %w", err)
	}

	org := users.NewOrganization(cmd.Name, cmd.Slug, ownerID)

	if cmd.Description != "" {
		org.UpdateDescription(cmd.Description)
	}

	if cmd.BillingEmail != "" {
		org.SetBillingEmail(cmd.BillingEmail)
	}

	if err := s.userRepo.SaveOrganization(ctx, org); err != nil {
		return nil, fmt.Errorf("failed to save organization: %w", err)
	}

	return org, nil
}

func (s *UserService) GetUser(ctx context.Context, id string) (*users.User, error) {
	userID, err := users.UserIDFromString(id)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	return s.userRepo.FindByID(ctx, userID)
}

func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*users.User, error) {
	userEmail, err := users.NewEmail(email)
	if err != nil {
		return nil, fmt.Errorf("invalid email: %w", err)
	}

	return s.userRepo.FindByEmail(ctx, userEmail)
}

func (s *UserService) GetUserByUsername(ctx context.Context, username string) (*users.User, error) {
	return s.userRepo.FindByUsername(ctx, username)
}

func (s *UserService) GetOrganization(ctx context.Context, id string) (*users.Organization, error) {
	orgID, err := users.OrganizationIDFromString(id)
	if err != nil {
		return nil, fmt.Errorf("invalid organization ID: %w", err)
	}

	return s.userRepo.FindOrganizationByID(ctx, orgID)
}

func (s *UserService) GetOrganizationBySlug(ctx context.Context, slug string) (*users.Organization, error) {
	return s.userRepo.FindOrganizationBySlug(ctx, slug)
}

func (s *UserService) GetUserOrganizations(ctx context.Context, userID string) ([]*users.Organization, error) {
	id, err := users.UserIDFromString(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	return s.userRepo.FindOrganizationsByUser(ctx, id)
}

func (s *UserService) ListUsers(ctx context.Context) ([]*users.User, error) {
	return s.userRepo.FindAll(ctx)
}

func (s *UserService) UpdateUser(ctx context.Context, id string, username *string) (*users.User, error) {
	user, err := s.GetUser(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if username != nil {
		userUsername, err := users.NewUsername(*username)
		if err != nil {
			return nil, fmt.Errorf("invalid username: %w", err)
		}
		user.SetUsername(userUsername)
	}

	if err := s.userRepo.Save(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

func (s *UserService) UpdateUserPassword(ctx context.Context, id string, passwordHash string) (*users.User, error) {
	user, err := s.GetUser(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	user.UpdatePassword(passwordHash)

	if err := s.userRepo.Save(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user password: %w", err)
	}

	return user, nil
}

func (s *UserService) VerifyUserEmail(ctx context.Context, id string) (*users.User, error) {
	user, err := s.GetUser(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	user.VerifyEmail()

	if err := s.userRepo.Save(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to verify user email: %w", err)
	}

	return user, nil
}

func (s *UserService) UpdateUserLastLogin(ctx context.Context, id string) (*users.User, error) {
	user, err := s.GetUser(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	user.UpdateLastLogin()

	if err := s.userRepo.Save(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update last login: %w", err)
	}

	return user, nil
}

func (s *UserService) DeleteUser(ctx context.Context, id string) error {
	userID, err := users.UserIDFromString(id)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	_, err = s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	return s.userRepo.Delete(ctx, userID)
}

func (s *UserService) DeleteOrganization(ctx context.Context, id string) error {
	orgID, err := users.OrganizationIDFromString(id)
	if err != nil {
		return fmt.Errorf("invalid organization ID: %w", err)
	}

	_, err = s.userRepo.FindOrganizationByID(ctx, orgID)
	if err != nil {
		return fmt.Errorf("organization not found: %w", err)
	}

	return s.userRepo.DeleteOrganization(ctx, orgID)
}
