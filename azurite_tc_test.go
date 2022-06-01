package azurite_tc

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"strconv"
)

func TestAzuriteTC(t *testing.T) {
	t.Run("azurite container", func(t *testing.T) {
		azuriteTC := NewAzuriteTC("devstoreaccount1", "Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==")

		azuriteTC.RunAzuriteContainer()

		azuriteTC.CreateTable("points")
		azuriteTC.UpdateTableValue("points", "user", "score", "120")

		value, err := azuriteTC.GetTableValue("points", "user", "score")
		updatedValue, err := strconv.ParseInt(value, 10, 32)

		assert.NoError(t, err)
		assert.Equal(t, int(updatedValue), 120)

		azuriteTC.UpdateTableValue("points", "user", "score", strconv.Itoa(int(updatedValue + 80)))
		value, err = azuriteTC.GetTableValue("points", "user", "score")

		assert.NoError(t, err)
		assert.Equal(t, value, "200")

		azuriteTC.RemoveAzuriteContainer()
	})
}
