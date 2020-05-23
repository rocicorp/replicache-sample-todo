package db

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBasics(t *testing.T) {
	assert := assert.New(t)
	db := New()
	out, err := db.Exec("SELECT Now()", nil)
	assert.NoError(err)
	assert.Equal(1, len(out.Records))
	assert.Regexp(`\d{4}\-\d{2}\-\d{2} \d{2}:\d{2}:\d{2}`, *out.Records[0][0].StringValue)
}

func TestTransact(t *testing.T) {
	assert := assert.New(t)

	db := New()
	var err error
	_, err = db.Exec("DROP DATABASE IF EXISTS test", nil)
	assert.NoError(err)
	_, err = db.Exec("CREATE DATABASE test", nil)
	assert.NoError(err)
	db.Use("test")
	_, err = db.Exec("CREATE TABLE Foo (Id INT NOT NULL PRIMARY KEY, Count INT NOT NULL)", nil)
	assert.NoError(err)
	_, err = db.Exec("INSERT INTO Foo (Id, Count) VALUE (1, 0)", nil)
	assert.NoError(err)

	tc := []struct {
		ret           bool   // return value of user function
		panic         bool   // whether user function panics
		expectedRet   bool   // expected return from Transact()
		expectedError string // expected error from Transact()
		expectedPanic bool   // whether Transact() expected to panic
		expectedVal   int    // expected count in DB after Transact() returns
		// TODO: test BEGIN and COMMIT and ROLLBACK failing
		// Probably need to mock out RDS to do this well
	}{
		{false, false, false, "", false, 0},
		{true, false, true, "", false, 1},
		{false, true, false, "", true, 1},
	}

	for i, t := range tc {
		var ret bool
		var err error
		var recovered interface{}
		msg := fmt.Sprintf("test case %d", i)
		func() {
			defer func() {
				recovered = recover()
			}()
			out, err := db.Exec("SELECT Count FROM Foo WHERE Id = 1", nil)
			assert.NoError(err)
			count := *out.Records[0][0].LongValue
			ret, err = db.Transact(func(exec ExecFunc) (commit bool) {
				_, err := exec("UPDATE Foo SET Count = Count + 1 WHERE Id = 1", nil)
				assert.NoError(err, msg)
			
				// Here we check isolation. A query run outside of the transaction
				// should not see the above update. (We had a bug where transactions
				// were not actually creating transactions.)
				out, err := db.Exec("SELECT Count FROM Foo WHERE Id = 1", nil)
				assert.NoError(err, msg)
				assert.Equal(count, *out.Records[0][0].LongValue, msg)
			
				if t.panic {
					panic("bonk")
				}
				return t.ret
			})
			db.Use("test")
		}()

		assert.Equal(t.expectedRet, ret, msg)

		if t.expectedError != "" {
			assert.EqualError(err, t.expectedError, msg)
		} else {
			assert.NoError(err, msg)
		}

		if t.expectedPanic {
			assert.Equal("bonk", recovered, msg)
		} else {
			assert.Nil(recovered, msg)
		}

		out, err := db.Exec("SELECT Count FROM Foo WHERE Id = 1", nil)
		assert.NoError(err, msg)
		assert.Equal(1, len(out.Records), msg)
		assert.Equal(int64(t.expectedVal), *out.Records[0][0].LongValue, msg)
	}
}
