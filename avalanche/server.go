package avalanche

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"log"
	"mime"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

type File struct {
	Name     string
	Mimetype string
}

type RouteMap = map[string]*File

func TransformRoute(files []string) (namedRoutes, normalRoutes, staticRoutes RouteMap) {
	namedRoutes = make(RouteMap)
	normalRoutes = make(RouteMap)
	staticRoutes = make(RouteMap)
	pattern := regexp.MustCompile(`\[(?P<id>(\w+))]$`)
	for _, file := range files {
		r := &File{
			Mimetype: "text/html",
			Name:     file,
		}
		route := strings.TrimPrefix(
			strings.TrimSuffix(
				strings.ReplaceAll(file, "/index.html", "/"),
				".html"),
			"bundle")
		if _, fname := path.Split(route); strings.HasPrefix(fname, "._") {
			continue
		} else {
			if strings.HasPrefix(route, "/_next") {
				ext := filepath.Ext(fname)
				r.Mimetype = mime.TypeByExtension(ext)
				staticRoutes[route] = r
				continue
			}
		}
		if sub := pattern.FindStringSubmatch(route); len(sub) > 0 {
			tPath := pattern.ReplaceAllLiteralString(route, fmt.Sprintf(":%s", sub[pattern.SubexpIndex("id")]))
			namedRoutes[tPath] = r
		} else {
			normalRoutes[route] = r
		}
	}
	return
}

func StartServer() {
	u, e := url.Parse("http://localhost:3000")
	if e != nil {
		panic(e)
	}
	var server http.Handler
	if os.Getenv("ENV") == "dev" {
		proxy := httputil.NewSingleHostReverseProxy(u)
		fmt.Println("Reversing proxy from dev server.")
		server = proxy
	} else if files, e := Traverse(Bundle); e == nil {
		router := httprouter.New()
		namedRoutes, normalRoutes, staticRoutes := TransformRoute(files)
		handler := RouteHandler{PageMap: normalRoutes, StaticMap: staticRoutes, FS: Bundle}
		for k, v := range namedRoutes {
			router.GET(k, func(writer http.ResponseWriter, request *http.Request, _ httprouter.Params) {
				handler.ServeFile(v, writer, request)
			})
		}
		for k, v := range normalRoutes {
			router.GET(k, func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
				handler.ServeFile(v, writer, request)
			})
		}
		router.NotFound = handler
		server = router
	} else {
		log.Fatalln("No suitable servers to start.")
	}
	addr := fmt.Sprintf(":%d", 8080)
	fmt.Println("Listening on", addr)
	panic(http.ListenAndServe(addr, &CrossOriginMiddleware{
		Handler: server,
	}))
}
