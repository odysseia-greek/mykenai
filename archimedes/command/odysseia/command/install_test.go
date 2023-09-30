package command

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInstallProfileCreation(t *testing.T) {
	first := "elastic"
	second := "solon"
	third := "homeros"
	fourth := "pheidias"
	fifth := "xerxes"

	t.Run("base", func(t *testing.T) {
		profile := "base"
		sut, err := readAppsYaml(profile)
		assert.Nil(t, err)
		assert.Contains(t, sut, first)
		assert.NotContains(t, sut, second)
		assert.NotContains(t, sut, third)
		assert.NotContains(t, sut, fourth)
		assert.NotContains(t, sut, fifth)
	})

	t.Run("full", func(t *testing.T) {
		profile := "full"
		sut, err := readAppsYaml(profile)
		assert.Nil(t, err)
		assert.Contains(t, sut, first)
		assert.Contains(t, sut, second)
		assert.Contains(t, sut, third)
		assert.Contains(t, sut, fourth)
		assert.NotContains(t, sut, fifth)
	})
}
