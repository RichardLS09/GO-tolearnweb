package web

import (
	"strconv"
	"strings"
	"time"
)

func RecoveryHandler(c *Context) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("[panic]", err)
			c.DieWithHttpStatus(500)
		}
	}()
	c.Next()
}

func str(v string) string {
	if v == "" {
		return "-"
	}
	return v
}

func LogHandler(c *Context) {
	start := time.Now()
	path := c.Request.URL.Path
	raw := c.Request.URL.RawQuery
	ips := c.Request.Header.Get("X-Forwarded-For")
	ip := ""
	if ips != "" {
		ip = strings.Split(ips, ",")[0]
	}
	if ip == "" {
		ip = c.Request.Header.Get("X-Real-Ip")
	}
	if ip == "" {
		ip = c.Request.RemoteAddr
	}

	traceId := c.Request.Header.Get("trace_id")
	uuid := c.Request.Header.Get("uuid")

	c.Next()

	end := time.Now()
	latency := end.Sub(start)
	method := c.Request.Method
	resp := string(c.Response)
	req := string(c.Body)
	if raw != "" {
		path = path + "?" + raw
	}

	log.Info(
		ip, // remote ip
		end.Format("2006/01/02 15:04:05"),
		latency.Nanoseconds(),
		str(method),
		str(path),
		str(traceId), // trace id
		str(uuid),    // uuid
		str(req),
		str(resp),
	)
}

func SessionWithCookieHandler(cookieSessionKey string, duration time.Duration) HandlerFunc {
	return func(c *Context) {
		token, err := c.GetCookie(cookieSessionKey)
		if err != nil || len(token) < 1 {
			token = RandMd5()
		}
		if err := c.StartSession(token); err != nil {
			c.Error(-1, err)
			return
		}
		c.SetCookie(cookieSessionKey, token, duration)
		c.Next()
	}
}

func SessionWithCookieDestoryHandler(cookieSessionKey string) HandlerFunc {
	return func(c *Context) {
		token, err := c.GetCookie(cookieSessionKey)
		if err == nil && len(token) > 0 {
			c.DestroySession()
			c.SetCookie(cookieSessionKey, token, -1*time.Second)
		}
		c.Next()
	}
}

func TokenAuthorization(token string) HandlerFunc {
	return func(c *Context) {
		auth := c.Request.Header.Get("Authorization")
		if auth != token {
			log.Error("[authorization failed]", auth)
			c.DieWithHttpStatus(401)
		} else {
			c.Next()
		}
	}
}

func SignCheck(sign string) HandlerFunc {
	return func(c *Context) {
		token := c.QueryDefault("token", "")
		timestamp, err := strconv.ParseInt(c.QueryDefault("timestamp", "0"), 10, 64)
		if err != nil {
			c.DieWithHttpStatus(401)
			return
		}
		if timestamp == 0 {
			if !Sha1Verify([]byte(sign), c.Body, []byte(token), 5) {
				c.DieWithHttpStatus(401)
				return
			}
		} else {
			if !Sha1VerifyTimestamp([]byte(sign), c.Body, []byte(token), 5, timestamp) {
				c.DieWithHttpStatus(401)
				return
			}
		}
		c.Next()
	}
}
