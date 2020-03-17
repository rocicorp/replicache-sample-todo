package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBasics(t *testing.T) {
	assert := assert.New(t)
	db := New()
	out, err := db.Exec("SELECT Now()")
	assert.NoError(err)
	assert.Equal(1, len(out.Records))
	assert.Regexp(`\d{4}\-\d{2}\-\d{2} \d{2}:\d{2}:\d{2}`, *out.Records[0][0].StringValue)
}
