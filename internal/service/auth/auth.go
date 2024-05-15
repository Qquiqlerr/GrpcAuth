package auth

import (
	"GrpcAuth/internal/domain/models"
	"GrpcAuth/internal/lib/jwt"
	"GrpcAuth/internal/storage"
	"context"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"time"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAppNotFound        = errors.New("app not found")
)

type Auth struct {
	log         *slog.Logger
	usrSaver    UserSaver
	usrProvider UserProvider
	appProvider AppProvider
	tokenTTL    time.Duration
}

type UserSaver interface {
	SaveUser(ctx context.Context, email string, passHash []byte) (int64, error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (models.User, error)
	IsAdmin(ctx context.Context, uID int64) (bool, error)
}

type AppProvider interface {
	App(ctx context.Context, appID int64) (models.App, error)
}

func New(log *slog.Logger, appprovider AppProvider, userprovider UserProvider, usersaver UserSaver, tokenTTL time.Duration) *Auth {
	return &Auth{
		log:         log,
		usrSaver:    usersaver,
		usrProvider: userprovider,
		appProvider: appprovider,
		tokenTTL:    tokenTTL,
	}
}
func (a *Auth) Login(ctx context.Context, email string, password string, appID int64) (token string, err error) {
	const op = "auth.Login"
	log := a.log.With(
		slog.String("op", op))
	user, err := a.usrProvider.User(ctx, email)
	if err != nil {
		if errors.Is(storage.ErrUserNotFound, err) {
			log.Warn("user not found", slog.Any("err", err))
			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		} else if errors.Is(storage.ErrUserExists, err) {
			log.Warn("user already exists", slog.Any("err", err))
			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}
		log.Error("failed to get user", slog.Any("err", err))
		return "", fmt.Errorf("%s: %w", op, err)
	}
	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		log.Warn("invalid password", slog.Any("err", err))
		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}
	app, err := a.appProvider.App(ctx, appID)
	if err != nil {
		if errors.Is(storage.ErrAppNotFound, err) {
			log.Warn("app not found", slog.Any("err", err))
			return "", fmt.Errorf("%s: %w", op, ErrAppNotFound)
		}
		log.Error("failed to get app", slog.Any("err", err))
		return "", fmt.Errorf("%s: %w", op, err)
	}
	log.Info("user logged in successfully")
	token, err = jwt.NewToken(user, app, a.tokenTTL)
	if err != nil {
		log.Error("failed to create token", slog.Any("err", err))
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return token, nil
}
func (a *Auth) RegisterNewUser(ctx context.Context, email string, password string) (userID int64, err error) {
	const op = "auth.RegisterNewUser"
	log := a.log.With(
		slog.String("op", op))
	log.Info("register new user")
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate hash", err.Error())
		return 0, fmt.Errorf("%s : %w", op, err)
	}
	id, err := a.usrSaver.SaveUser(ctx, email, hash)
	if err != nil {
		log.Error("failed to save user", err.Error())
		return 0, fmt.Errorf("%s : %w", op, err)
	}
	log.Info("save user success")
	return id, nil
}

func (a *Auth) IsAdmin(ctx context.Context, userID int64) (isAdmin bool, err error) {
	const op = "auth.IsAdmin"
	log := a.log.With(
		slog.String("op", op),
		slog.Int64("userID", userID),
	)
	isAdmin, err = a.usrProvider.IsAdmin(ctx, userID)
	if err != nil {
		log.Error("failed to get user", slog.Any("err", err))
		return false, fmt.Errorf("%s : %w", op, err)
	}
	return isAdmin, nil
}
