package helper

import (
	"github.com/go-faster/errors"
	"github.com/stretchr/testify/assert"

	"net/textproto"
	"testing"
	"time"
)

func Test_CanCancel(t *testing.T) {
	t.Run("negative case", func(t *testing.T) {
		t.Run("can nor cancel", func(t *testing.T) {
			//Arrange
			note := Notification{
				SendAt:    time.Time{},
				CreatedAt: time.Time{},
				SentAt:    time.Time{},
				Status:    StatusDead,
				Attempts:  6,
			}
			expRes := false

			//Act
			res := CanCancel(note)

			//Assert

			assert.Equal(t, expRes, res)
		})
	})

	t.Run("positive case", func(t *testing.T) {
		t.Run("ok", func(t *testing.T) {
			//Arrange
			note := Notification{
				SendAt:    time.Time{},
				CreatedAt: time.Time{},
				SentAt:    time.Time{},
				Status:    StatusPending,
				Attempts:  2,
			}
			expRes := true

			//Act
			res := CanCancel(note)

			//Assert
			assert.Equal(t, expRes, res)
		})
	})
}

func Test_NextAttempt(t *testing.T) {
	t.Run("negative case", func(t *testing.T) {
		t.Run("attempts are maxed", func(t *testing.T) {
			//Arrange
			note := Notification{
				SendAt:    time.Time{},
				CreatedAt: time.Time{},
				SentAt:    time.Time{},
				Status:    StatusSending,
				Attempts:  6,
			}
			status := note.Status

			//Act
			NextAttempt(&note)

			//Assert
			assert.NotEqual(t, status, note.Status)

		})
	})

	t.Run("positive case", func(t *testing.T) {
		t.Run("ok", func(t *testing.T) {
			//Arrange
			note := Notification{
				SendAt:    time.Date(2013, time.October, 13, 0, 0, 0, 0, time.UTC),
				CreatedAt: time.Time{},
				SentAt:    time.Time{},
				Status:    StatusPending,
				Attempts:  5,
			}
			sendAt := note.SendAt

			//Act
			NextAttempt(&note)

			//Assert
			assert.NotEqual(t, sendAt, note.SendAt)
		})
	})
}

func Test_IsErrorRetryable(t *testing.T) {
	t.Run("negative case", func(t *testing.T) {
		t.Run("unable to convert types", func(t *testing.T) {
			//Arrange
			errorka := errors.New("lol")
			expectedError := ErrPermanent

			//Act
			_, err := IsErrorRetryable(errorka)

			//Assert
			assert.ErrorIs(t, expectedError, err)
		})
	})

	t.Run("positive case", func(t *testing.T) {
		t.Run("ok", func(t *testing.T) {
			//Arrange
			errorka := &textproto.Error{
				Code: 404,
				Msg:  "not found",
			}
			expRes := false

			//Act
			res, err := IsErrorRetryable(errorka)

			//Assert
			assert.Nil(t, err)
			assert.Equal(t, expRes, res)
		})
	})
}
