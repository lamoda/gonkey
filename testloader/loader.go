package testloader

import (
	"github.com/lamoda/gonkey/models"
)

type LoaderInterface interface {
	Load() ([]models.TestInterface, error)
}
