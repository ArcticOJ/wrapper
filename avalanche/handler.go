package avalanche

import (
	"io"
	"io/fs"
	"net/http"
	"strconv"
	"strings"
)

type RouteHandler struct {
	FS        fs.FS
	PageMap   RouteMap
	StaticMap RouteMap
}

func (handler RouteHandler) ServeFile(file *File, writer http.ResponseWriter, request *http.Request) {
	f, e := handler.FS.Open(file.Name)
	if e != nil {
		http.NotFound(writer, request)
		return
	}
	info, e := f.Stat()
	if e == nil {
		writer.Header().Add("Content-Length", strconv.FormatInt(info.Size(), 10))
	}
	buf, e := io.ReadAll(f)
	if e != nil {
		http.NotFound(writer, request)
		return
	}
	writer.Header().Add("Content-Type", file.Mimetype)
	writer.Write(buf)
}

func (handler RouteHandler) Handle404(writer http.ResponseWriter, request *http.Request) {
	if file, ok := handler.PageMap["bundle/404.html"]; ok {
		writer.WriteHeader(http.StatusNotFound)
		handler.ServeFile(file, writer, request)
		return
	}
	http.NotFound(writer, request)
}

func (handler RouteHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	path := request.URL.Path
	if len(path) > 1 && path[len(path)-1] == '/' {
		request.URL.Path = path[:len(path)-1]
		http.Redirect(writer, request, request.URL.String(), http.StatusPermanentRedirect)
		return
	}
	if strings.HasPrefix(path, "/_next/") {
		if file, ok := handler.StaticMap[path]; ok {
			handler.ServeFile(file, writer, request)
			return
		}
	}
	/*if file, ok := handler.PageMap[path]; ok {
		handler.ServeFile(file, writer, request)
		return
	}*/
	handler.Handle404(writer, request)
}
