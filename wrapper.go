package wrapper

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"io"
	"io/fs"
	"mime"
	"net/url"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
)

type RouteMap = map[string]string

func Traverse(efs fs.FS) (files []string, err error) {
	if err := fs.WalkDir(efs, ".", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		files = append(files, path)
		return nil
	}); err != nil {
		return nil, err
	}
	return files, nil
}

func BuildRoutes(files []string) (routes RouteMap) {
	routes = make(RouteMap)
	pattern := regexp.MustCompile(`\[(?P<id>(\w+))]`)
	for _, file := range files {
		if strings.HasPrefix(file, "_next") {
			routes[file] = file
			continue
		}
		route := strings.TrimPrefix(
			strings.TrimSuffix(
				strings.ReplaceAll(file, "/index.html", "/"),
				".html"),
			"bundle")
		if _, fname := path.Split(route); strings.HasPrefix(fname, "._") {
			continue
		}
		if sub := pattern.FindStringSubmatch(route); len(sub) > 0 {
			param := sub[pattern.SubexpIndex("id")]
			tPath := pattern.ReplaceAllLiteralString(route, ":"+param)
			routes[tPath] = file
		} else {
			if route == "index" {
				route = "/"
			}
			routes[route] = file
		}
	}
	return
}

func ServeFile(_fs fs.FS, c echo.Context, name string, mime string, is404 bool) error {
	f, e := _fs.Open(name)
	if e != nil {
		if !is404 {
			return Handle404(_fs, c)
		}
		return e
	}
	defer f.Close()
	h := c.Response().Header()
	stat, _e := f.Stat()
	if _e == nil {
		h.Set("Content-Length", strconv.FormatInt(stat.Size(), 10))
		h.Set("Last-Modified", stat.ModTime().UTC().Format("Wed, 21 Oct 2015 07:28:00 GMT"))
	}
	h.Set("Content-Type", mime)
	buf, _ := io.ReadAll(f)
	_, err := c.Response().Write(buf)
	return err
}

func Handle404(_fs fs.FS, c echo.Context) error {
	return ServeFile(_fs, c, "404.html", "text/html", true)
}

func Register(ec *echo.Echo, Bundle fs.FS) {
	ec.Use(COEP())
	if os.Getenv("ENV") == "dev" {
		// my own ip, for local development only
		u, _ := url.Parse("http://localhost:3000")
		ec.Use(middleware.ProxyWithConfig(middleware.ProxyConfig{
			Skipper: func(c echo.Context) bool {
				return strings.HasPrefix(c.Request().URL.Path, "/api")
			},
			Balancer: middleware.NewRoundRobinBalancer([]*middleware.ProxyTarget{
				{
					URL: u,
				},
			}),
		}))
	} else if files, e := Traverse(Bundle); e == nil {
		routes := BuildRoutes(files)
		for k, v := range routes {
			x := v
			t := "text/html"
			if strings.HasPrefix(k, "_next") {
				t = mime.TypeByExtension(path.Ext(x))
			}
			ec.GET(k, func(c echo.Context) error {
				return ServeFile(Bundle, c, x, t, k == "404")
			})
		}
		ec.RouteNotFound("/*", func(c echo.Context) error {
			return ServeFile(Bundle, c, "404.html", "text/html", true)
		})
	}
}
