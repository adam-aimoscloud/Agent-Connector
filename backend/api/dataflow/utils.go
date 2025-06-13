package dataflow

import (
	"errors"

	"github.com/gin-gonic/gin"
)

// GetAuthInfoFromContext gets authentication info from gin context
func GetAuthInfoFromContext(c *gin.Context) (*AuthInfo, error) {
	authInfoValue, exists := c.Get("authInfo")
	if !exists {
		return nil, errors.New("authentication info not found in context")
	}

	authInfo, ok := authInfoValue.(*AuthInfo)
	if !ok {
		return nil, errors.New("invalid authentication info type in context")
	}

	return authInfo, nil
}
