package authorization

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestWithUserID_FromContext(t *testing.T) {
	t.Run("round trip", func(t *testing.T) {
		userID := uuid.New()
		ctx := WithUserID(context.Background(), userID)

		got := FromContext(ctx)

		if assert.NotNil(t, got) {
			assert.Equal(t, userID, *got)
		}
	})

	t.Run("missing value returns nil, not panic", func(t *testing.T) {
		got := FromContext(context.Background())

		assert.Nil(t, got)
	})
}
