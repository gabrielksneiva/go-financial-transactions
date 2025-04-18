package services

import (
	"errors"

	d "github.com/gabrielksneiva/go-financial-transactions/domain"
	"github.com/jackc/pgx/v5/pgconn"
)

type UserService struct {
	repo d.UserRepository
}

func NewUserService(repo d.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) CreateUser(name, email, password string) error {
	user := d.User{Name: name, Email: email, Password: password}
	err := s.repo.Create(user)
	if err != nil {
		// forma 2: com GORM v2 e tratamento de erro do driver
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return errors.New("e-mail já cadastrado")
		}

		return err
	}

	return nil
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
