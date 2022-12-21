package utils

import (
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Generate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		test := GenerateID()
		id := uuid.NewV4()

		wow, _ := uuid.FromString(test)
		log.Warn("test: ", wow)
		log.Warn("id: ", id)
		assert.NotEqual(t, test, "")
	})
}
