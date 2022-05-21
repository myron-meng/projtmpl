package service

import (
	"projtmpl/internal/repository"
)

type Services struct {
	Repos *repository.Repositories
}

func New(repos *repository.Repositories) *Services {
	return &Services{Repos: repos}
}
