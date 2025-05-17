package service

import "github.com/Jonathan0823/auth-go/internal/dto"

func (s *service) GetUserByID(id int) (dto.User, error) {
	return s.repo.GetUserByID(id)
}
