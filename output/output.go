package output

import (
	"github.com/lamoda/gonkey/models"
)

type OutputInterface interface {
	Process(models.TestInterface, *models.Result) error
}
