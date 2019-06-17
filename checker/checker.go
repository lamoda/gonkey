package checker

import "github.com/lamoda/gonkey/models"

type CheckerInterface interface {
	Check(models.TestInterface, *models.Result) ([]error, error)
}
