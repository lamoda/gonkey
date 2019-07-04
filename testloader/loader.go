package testloader

import (
	"github.com/keyclaim/gonkey/models"
)

type LoaderInterface interface {
	Load() (chan models.TestInterface, error)
}
