package jwt_test

import (
	"testing"

	"github.com/aoaYaoa/go-gin-starter/pkg/utils/jwt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestGenerateAccessTokenWithTenant_EmbedsTenantID(t *testing.T) {
	require.NoError(t, jwt.SetDefaultSecret("test-secret-that-is-long-enough-32chars"))

	userID := uuid.New()
	tenantID := uuid.New()

	token, err := jwt.GenerateAccessTokenWithTenant(userID, tenantID, "alice", "user")
	require.NoError(t, err)

	claims, err := jwt.ParseToken(token)
	require.NoError(t, err)
	require.Equal(t, userID, claims.UserID)
	require.Equal(t, tenantID, claims.TenantID)
}
