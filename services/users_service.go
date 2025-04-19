package services

import (
	"errors"

	d "github.com/gabrielksneiva/go-financial-transactions/domain"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo d.UserRepository
}

func NewUserService(repo d.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) CreateUser(name, email, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return errors.New("erro ao gerar hash da senha")
	}

	user := d.User{Name: name, Email: email, Password: string(hashedPassword)}
	err = s.repo.Create(user)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return errors.New("e-mail já cadastrado")
		}
		return err
	}

	return nil
}

func (s *UserService) Authenticate(email, password string) (*d.User, error) {
	user, err := s.repo.GetByEmail(email)
	if err != nil {
		return &d.User{}, errors.New("usuário não encontrado")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return &d.User{}, errors.New("senha inválida")
	}

	return user, nil
}

func (s *UserService) GetUserByID(id uint) (*d.User, error) {
	user, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetUserByEmail(email string) (*d.User, error) {
	user, err := s.repo.GetByEmail(email)
	if err != nil {
		return nil, err
	}

	return user, nil
}
