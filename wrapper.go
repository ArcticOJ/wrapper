package wrapper

import (
	"github.com/ArcticOJ/blizzard/v0/logger"
	rice "github.com/GeertJohan/go.rice"
	"github.com/labstack/echo/v4"
	"io"
	"io/fs"
	"mime"
	"path"
	"regexp"
	"strconv"
	"strings"
)

type RouteMap = map[string]string

func Traverse(box *rice.Box) (files []string, err error) {
	if err := box.Walk(".", func(path string, inf fs.FileInfo, err error) error {
		if inf.IsDir() {
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

func ServeFile(box *rice.Box, c echo.Context, name string, mime string, is404 bool) error {
	f, e := box.Open(name)
	if e != nil {
		if !is404 {
			return ServeFile(box, c, "404.html", "text/html", true)
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

func Register(ec *echo.Echo, bundle *rice.Box) {
	if files, e := Traverse(bundle); e == nil {
		routes := BuildRoutes(files)
		for k, v := range routes {
			x := v
			t := "text/html"
			if strings.HasPrefix(k, "_next") {
				t = mime.TypeByExtension(path.Ext(x))
			}
			ec.GET(k, func(c echo.Context) error {
				return ServeFile(bundle, c, x, t, k == "404")
			})
		}
		ec.RouteNotFound("/*", func(c echo.Context) error {
			return ServeFile(bundle, c, "404.html", "text/html", true)
		})
	} else {
		logger.Panic(e, "failed to load embedded web bundle")
	}
}
