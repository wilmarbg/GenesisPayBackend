package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	Repo			*UserRepository
	JWTSecret		string
	JWTExpirationHours	int
}

func NewAuthService(repo *UserRepository, secret string, hours int) *AuthService {
	return &AuthService{
		Repo:			repo,
		JWTSecret:		secret,
		JWTExpirationHours:	hours,
	}
}

func (s *AuthService) Register(req RegisterRequest) (*AuthResponse, error) {
	if req.Role == "ADMINISTRADOR" {
		return nil, errors.New("No es posible registrar este tipo de usuario")
	}
	if req.Role != "CLIENTE" && req.Role != "COMERCIO" {
		return nil, errors.New("Rol no permitido. Solo se permite CLIENTE o COMERCIO")
	}

	existingUser, _ := s.Repo.FindByEmail(req.Email)
	if existingUser != nil {
		return nil, errors.New("El correo ya está registrado")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
	if err != nil {
		return nil, errors.New("Error al procesar la contraseña")
	}

	user := &User{
		Email:		req.Email,
		PasswordHash:	string(hashedPassword),
		Role:		req.Role,
		IsActive:	true,
	}

	if err := s.Repo.Create(user); err != nil {
		return nil, errors.New("Error al crear el usuario")
	}

	token, err := s.generateToken(user)
	if err != nil {
		return nil, errors.New("Error al generar el token")
	}

	return &AuthResponse{
		User:	*user,
		Token:	token,
	}, nil
}

func (s *AuthService) Login(req LoginRequest) (*AuthResponse, error) {

	user, err := s.Repo.FindByEmail(req.Email)
	if err != nil {
		return nil, errors.New("Correo o contraseña incorrectos")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, errors.New("Correo o contraseña incorrectos")
	}

	if !user.IsActive {
		return nil, errors.New("El usuario está inactivo")
	}

	token, err := s.generateToken(user)
	if err != nil {
		return nil, errors.New("Error al generar el token")
	}

	return &AuthResponse{
		User:	*user,
		Token:	token,
	}, nil
}

func (s *AuthService) GetProfile(userID uuid.UUID) (*User, error) {
	user, err := s.Repo.FindByID(userID)
	if err != nil {
		return nil, errors.New("Usuario no encontrado")
	}
	return user, nil
}

func (s *AuthService) generateToken(user *User) (string, error) {

	claims := jwt.MapClaims{
		"user_id":	user.ID.String(),
		"email":	user.Email,
		"role":		user.Role,
		"exp":		time.Now().Add(time.Hour * time.Duration(s.JWTExpirationHours)).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.JWTSecret))
}
