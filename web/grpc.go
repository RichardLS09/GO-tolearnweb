package web

import (
	"encoding/json"
	"golang.org/x/net/context"
	"google.golang.org/grpc/peer"
	"net"
	"net/http"
	"reflect"
	"strings"
)

// GRPC can route a http request to a grpcHandler
// and response as a http response
// The url path rule is:
//   YourGrpcHandle -> your_grpc_handle
// The param is YourRequestParam's json serialization
// The response is:
//     {
//       "status":0,
//       "data": <YourResponse's json serialization>
//     }
// The error response will return HTTP STATUS
//   400: invalid param
//   500: grpc handler returns error
func (this *Server) GRPC(path string, grpcHandler interface{}) *Router {
	return this.router.GRPC(path, grpcHandler)
}

func (this *Router) GRPC(path string, grpcHandler interface{}) *Router {
	handler := buildHandlerFromGrpcHandler(grpcHandler)
	return this.handle(http.MethodPost, path, handler)
}

func buildHandlerFromGrpcHandler(grpcHandler interface{}) HandlerFunc {
	handlerType := reflect.TypeOf(grpcHandler)
	handler := reflect.ValueOf(grpcHandler)
	paramType := handlerType.In(1)
	return func(c *Context) {
		param := reflect.New(paramType.Elem()).Interface()
		if err := json.Unmarshal(c.Body, &param); err != nil {
			log.Error("grpc param bind error", err, string(c.Body))
			c.DieWithHttpStatus(400)
			return
		}
		ctx := context.Background()
		// add peer
		ip := c.Request.RemoteAddr
		if index := strings.LastIndex(ip, ":"); index > 0 {
			ip = ip[:index]
		}
		if addr, err := net.ResolveIPAddr("", ip); err == nil {
			p := &peer.Peer{Addr: addr}
			ctx = peer.NewContext(ctx, p)
		}

		ret := handler.Call([]reflect.Value{
			reflect.ValueOf(ctx),
			reflect.ValueOf(param),
		})
		resp := ret[0].Interface()
		err := ret[1].Interface()
		if err != nil {
			log.Error("grpc deal error", err)
			c.DieWithHttpStatus(500)
			return
		}
		c.Success(resp)
	}
}
