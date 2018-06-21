package web

const (
	HTTP_PROXY_METHOD = "PROXY"
)

func (this *Server) PROXY(path string, handlers ...HandlerFunc) *Router {
	return this.router.PROXY(path, handlers...)
}

func (this *Router) PROXY(path string, handlers ...HandlerFunc) *Router {
	return this.handle(HTTP_PROXY_METHOD, path, handlers...)
}
