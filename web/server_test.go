package web

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

var server *Server

var route_cases = []string{
	"/",
	"/a",
	"/b/a",
	"/c/:a",
	"/d/",
}

func init() {
	go func() {
		server = New(":9999")
		server.Run()
	}()
	time.Sleep(1 * time.Second)
}

func TestServer(t *testing.T) {
	for _, path := range route_cases {
		server.GET(path, func(c *Context) {
			c.Success(c.Path())
		})
	}
	for _, path := range route_cases {
		realPath := strings.Replace(path, ":", "", -1)
		res, err := http.Get("http://127.0.0.1:9999" + realPath)
		if err != nil {
			t.Error(err)
		}
		resp, _ := ioutil.ReadAll(res.Body)
		if string(resp) != `{"data":"`+realPath+`","status":0}` {
			t.Error(string(resp))
		}
	}
}

var param_cases = []string{
	"/param/:a",
	"/param/:a/",
	"/param/:a/:b",
	"/param/:a/:b/a",
}

func TestParam(t *testing.T) {
	for _, path := range param_cases {
		server.GET(path, func(c *Context) {
			a, ok := c.Param("a")
			if !ok {
				a = "a"
			}
			b, ok := c.Param("b")
			if !ok {
				b = "b"
			}
			c.Success(map[string]string{
				"a": a,
				"b": b,
			})
		})
	}
	for _, path := range param_cases {
		realPath := strings.Replace(path, ":", "", -1)
		res, err := http.Get("http://127.0.0.1:9999" + realPath)
		if err != nil {
			t.Error(err)
		}
		resp, _ := ioutil.ReadAll(res.Body)
		if string(resp) != `{"data":{"a":"a","b":"b"},"status":0}` {
			t.Error(string(resp))
		}
	}
}

func TestQuery(t *testing.T) {
	server.GET("/query", func(c *Context) {
		q, _ := c.Query("q")
		_, ok := c.Query("nq")
		if ok {
			t.Error("nq should not found")
		}
		qd := c.QueryDefault("qd", "qd")

		c.Success(map[string]string{
			"q":  q,
			"qd": qd,
		})
	})
	res, err := http.Get("http://127.0.0.1:9999/query?q=q")
	if err != nil {
		t.Error(err)
	}
	resp, _ := ioutil.ReadAll(res.Body)
	if string(resp) != `{"data":{"q":"q","qd":"qd"},"status":0}` {
		t.Error(string(resp))
	}
}

func TestPostJson(t *testing.T) {
	server.POST("/post-json", func(c *Context) {
		a := struct {
			A string `json:"a"`
			B int    `json:"b"`
		}{}
		if err := c.Bind(&a); err != nil {
			t.Error("bind error", err)
			return
		}
		c.Success(a)
	})
	resp, err := PostJson("http://127.0.0.1:9999/post-json", []byte(`{"a":"aaa","b":1}`))
	if err != nil {
		t.Error("post error", err)
		return
	}
	if string(resp) != `{"data":{"a":"aaa","b":1},"status":0}` {
		t.Error(string(resp))
	}
}

func TestPostForm(t *testing.T) {
	server.POST("/post-form", func(c *Context) {
		a := c.Request.FormValue("a")
		c.Success(a)
	})
	resp, err := PostForm("http://127.0.0.1:9999/post-form", url.Values{"a": []string{"a"}})
	if err != nil {
		t.Error("post error", err)
		return
	}
	if string(resp) != `{"data":"a","status":0}` {
		t.Error(string(resp))
	}
}
