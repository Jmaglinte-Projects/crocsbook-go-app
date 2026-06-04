package authsvc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Jmaglinte-Projects/crocsbook-go-app/domain/user"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/usecase/usersvc"
	jwt "github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidIDToken      = errors.New("authsvc: invalid Google ID token")
	ErrInvalidCredentials  = errors.New("authsvc: invalid email or password")
	ErrEmailAlreadyExists  = errors.New("authsvc: email already exists")
	ErrInvalidRegisterForm = errors.New("authsvc: invalid register form")
)

type Service interface {
	// GoogleSignIn verifies the Google ID token (from frontend GSI), finds or creates the user, returns our JWT.
	GoogleSignIn(ctx context.Context, in *GoogleSignInIn) (*GoogleSignInOut, error)
	RegisterUser(ctx context.Context, in *RegisterUserIn) (*RegisterUserOut, error)
	SignIn(ctx context.Context, in *SignInIn) (*SignInOut, error)
}

type GoogleSignInIn struct {
	IDToken string
}

type GoogleSignInOut struct {
	JwtToken string
}

type RegisterUserIn struct {
	Email    string
	Gender   string
	Nickname string
	Username string
	Password string
}

type RegisterUserOut struct {
	JwtToken string
}

type SignInIn struct {
	Email    string
	Password string
}

type SignInOut struct {
	JwtToken string
}

type service struct {
	jwtSecret      string
	jwtExpiration  time.Duration
	googleClientID string
	userRepo       usersvc.UserRepository
	userSvc        usersvc.UserService
}

// NewService builds the auth service for backend-auth flow (frontend sends ID token, no redirect).
func NewService(
	jwtSecret string,
	jwtExpiration time.Duration,
	googleClientID string,
	userRepo usersvc.UserRepository,
	userSvc usersvc.UserService,
) Service {
	if jwtExpiration <= 0 {
		jwtExpiration = 24 * time.Hour
	}
	return &service{
		jwtSecret:      jwtSecret,
		jwtExpiration:  jwtExpiration,
		googleClientID: googleClientID,
		userRepo:       userRepo,
		userSvc:        userSvc,
	}
}

func (s *service) GoogleSignIn(ctx context.Context, in *GoogleSignInIn) (*GoogleSignInOut, error) {
	if in.IDToken == "" {
		return nil, ErrInvalidIDToken
	}

	info, err := s.verifyGoogleIDToken(ctx, in.IDToken)
	if err != nil {
		return nil, err
	}

	users, err := s.userSvc.List(ctx, user.ListCond{Email: &info.Email}, usersvc.ListOption{})
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}

	var userID user.UserID
	if len(users) == 0 {
		userID, err = s.createUserFromGoogle(ctx, info)
		if err != nil {
			return nil, fmt.Errorf("create user: %w", err)
		}
	} else {
		userID = users[0].UserID
	}

	token, err := s.signJWT(userID)
	if err != nil {
		return nil, fmt.Errorf("sign jwt: %w", err)
	}

	return &GoogleSignInOut{JwtToken: token}, nil
}

func (s *service) RegisterUser(ctx context.Context, in *RegisterUserIn) (*RegisterUserOut, error) {
	email := strings.TrimSpace(in.Email)
	nickname := strings.TrimSpace(in.Nickname)
	username := strings.TrimSpace(in.Username)
	password := strings.TrimSpace(in.Password)

	if email == "" || nickname == "" || username == "" || password == "" {
		return nil, ErrInvalidRegisterForm
	}

	gender, err := parseGender(in.Gender)
	if err != nil {
		return nil, err
	}

	users, err := s.userSvc.List(ctx, user.ListCond{Email: &email}, usersvc.ListOption{})
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	if len(users) > 0 {
		return nil, ErrEmailAlreadyExists
	}

	userID, err := user.NewUserID()
	if err != nil {
		return nil, fmt.Errorf("new user id: %w", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	hashedPasswordValue := string(hashedPassword)

	entity := &user.User{
		UserID:      userID,
		Email:       email,
		Gender:      gender,
		Nickname:    &nickname,
		Username:    &username,
		Password:    &hashedPasswordValue,
		CreatedTime: time.Now(),
	}
	if err := s.userRepo.Store(ctx, entity); err != nil {
		return nil, fmt.Errorf("store user: %w", err)
	}

	token, err := s.signJWT(userID)
	if err != nil {
		return nil, fmt.Errorf("sign jwt: %w", err)
	}

	return &RegisterUserOut{JwtToken: token}, nil
}

func (s *service) SignIn(ctx context.Context, in *SignInIn) (*SignInOut, error) {
	email := strings.TrimSpace(in.Email)
	password := strings.TrimSpace(in.Password)
	if email == "" || password == "" {
		return nil, ErrInvalidCredentials
	}

	users, err := s.userSvc.List(ctx, user.ListCond{Email: &email}, usersvc.ListOption{})
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	if len(users) == 0 || users[0].Password == nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*users[0].Password), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := s.signJWT(users[0].UserID)
	if err != nil {
		return nil, fmt.Errorf("sign jwt: %w", err)
	}
	return &SignInOut{JwtToken: token}, nil
}

// verifyGoogleIDToken calls Google's tokeninfo endpoint and validates aud/iss/exp, returns claims or error.
func (s *service) verifyGoogleIDToken(ctx context.Context, idToken string) (*GoogleUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://oauth2.googleapis.com/tokeninfo?id_token="+url.QueryEscape(idToken), nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("tokeninfo request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ErrInvalidIDToken
	}

	var payload struct {
		Aud           string `json:"aud"`
		Iss           string `json:"iss"`
		Exp           string `json:"exp"`
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified string `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
		GivenName     string `json:"given_name"`
		FamilyName    string `json:"family_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, ErrInvalidIDToken
	}

	if payload.Aud != s.googleClientID {
		return nil, ErrInvalidIDToken
	}
	if payload.Iss != "accounts.google.com" && payload.Iss != "https://accounts.google.com" {
		return nil, ErrInvalidIDToken
	}
	if payload.Email == "" {
		return nil, ErrInvalidIDToken
	}

	// exp is already validated by tokeninfo (200 only if valid); optional double-check
	var exp int64
	if _, err := fmt.Sscanf(payload.Exp, "%d", &exp); err == nil && time.Now().Unix() > exp {
		return nil, ErrInvalidIDToken
	}

	return &GoogleUserInfo{
		ID:            payload.Sub,
		Email:         payload.Email,
		VerifiedEmail: payload.EmailVerified == "true",
		Name:          payload.Name,
		Picture:       payload.Picture,
		GivenName:     payload.GivenName,
	}, nil
}

func (s *service) createUserFromGoogle(ctx context.Context, info *GoogleUserInfo) (user.UserID, error) {
	id, err := user.NewUserID()
	if err != nil {
		return "", err
	}

	entity := &user.User{
		UserID:      id,
		Email:       info.Email,
		Gender:      user.Gender(""),
		ProfileURL:  stringPtr(info.Picture),
		Nickname:    stringPtr(info.Name),
		CreatedTime: time.Now(),
	}

	if err := s.userRepo.Store(ctx, entity); err != nil {
		return "", err
	}

	return id, nil
}

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func parseGender(value string) (user.Gender, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "male":
		return user.Gender_Male, nil
	case "female":
		return user.Gender_Female, nil
	default:
		return "", ErrInvalidRegisterForm
	}
}

type jwtClaims struct {
	jwt.RegisteredClaims
	UserID string `json:"user_id"`
}

func (s *service) signJWT(userID user.UserID) (string, error) {
	now := time.Now()
	claims := jwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.jwtExpiration)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		UserID: string(userID),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// ParseUserIDFromJWT verifies an HS256 JWT signed with the same secret as NewService and returns the user id from the `sub` claim.
func ParseUserIDFromJWT(secret, tokenString string) (user.UserID, error) {
	if secret == "" {
		return "", errors.New("authsvc: empty jwt secret")
	}
	tok, err := jwt.ParseWithClaims(tokenString, &jwtClaims{}, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return "", err
	}
	claims, ok := tok.Claims.(*jwtClaims)
	if !ok || !tok.Valid {
		return "", errors.New("invalid token")
	}
	if claims.UserID == "" {
		return "", errors.New("missing sub")
	}
	return user.UserID(claims.UserID), nil
}

// GoogleUserInfo from verified ID token (or userinfo).
type GoogleUserInfo struct {
	ID            string
	Email         string
	VerifiedEmail bool
	Name          string
	GivenName     string
	Picture       string
	Hd            string
}
