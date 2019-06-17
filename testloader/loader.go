package testloader

import (
	"github.com/lamoda/gonkey/models"
)

type LoaderInterface interface {
	Load() (chan models.TestInterface, error)
}
