package output

import (
	"github.com/keyclaim/gonkey/models"
)

type OutputInterface interface {
	Process(models.TestInterface, *models.Result) error
}
