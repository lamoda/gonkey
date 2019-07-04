package checker

import "github.com/keyclaim/gonkey/models"

type CheckerInterface interface {
	Check(models.TestInterface, *models.Result) ([]error, error)
}
