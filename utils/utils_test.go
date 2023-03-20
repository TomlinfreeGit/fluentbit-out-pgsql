package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GetPostgresUrl_Ok(t *testing.T) {

	url := "postgres://user:sda%25&%5E%25@127.0.0.1:5432/agent?connect_timeout=30"
	res := GetPostgresUrl("user", "sda%&^%", "127.0.0.1", "5432", "agent")
	assert.Equal(t, url, res)
}
