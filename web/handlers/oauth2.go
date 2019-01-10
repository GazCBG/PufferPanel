package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/pufferpanel/apufferi/http"
	"github.com/pufferpanel/apufferi/logging"
	"github.com/pufferpanel/pufferpanel/services"
	webHttp "net/http"
	"strconv"
)

func OAuth2(scope string, requireServer bool) gin.HandlerFunc {
	return oauth2Handler(scope, requireServer, false)
}

func OAuth2WithLimit(scope string, requireServer bool) gin.HandlerFunc {
	return oauth2Handler(scope, requireServer, true)
}

func oauth2Handler (scope string, requireServer bool, permitWithLimit bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		os, err := services.GetOAuthService()
		if err != nil {
			http.Respond(c).Status(webHttp.StatusInternalServerError).Message("oauth2 service is not available").Fail().Send()
			logging.Errorf("oauth2 service is not available", err)
			c.Abort()
		}

		var serverId *uint
		var id uint

		i := c.Query("serverId")
		if i != "" {
			t, _ := strconv.Atoi(i)
			id = uint(t)
			serverId = &id
		}

		if requireServer && serverId == nil {
			http.Respond(c).Status(webHttp.StatusUnauthorized).Fail().Send()
			c.Abort()
			return
		}

		ti, err := os.ValidationBearerToken(c.Request)
		if err != nil {
			http.Respond(c).Status(webHttp.StatusUnauthorized).Fail().Send()
			c.Abort()
			return
		}

		ci, allowed, err := os.HasRights(ti.GetAccess(), serverId, scope)
		if err != nil {
			http.Respond(c).Status(webHttp.StatusInternalServerError).Message("error validating credentials").Fail().Send()
			logging.Errorf("error validating credentials", err)
			c.Abort()
			return
		}

		if !allowed && permitWithLimit {
			for _, v := range ci.ServerScopes {
				if v.Scope == scope {
					allowed = true
					break
				}
			}
		}

		if !allowed {
			if ci == nil {
				http.Respond(c).Status(webHttp.StatusUnauthorized).Fail().Send()
				c.Abort()
			} else {
				http.Respond(c).Status(webHttp.StatusForbidden).Fail().Send()
				c.Abort()
			}
		} else {
			c.Set("accessToken", ti.GetAccess())
			c.Next()
		}
	}
}