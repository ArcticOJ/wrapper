package avalanche

import "net/http"

type CrossOriginMiddleware struct {
	Handler http.Handler
}

func (mw CrossOriginMiddleware) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Add("Cross-Origin-Opener-Policy", "same-origin")
	writer.Header().Add("Cross-Origin-Embedder-Policy", "credentialless")
	mw.Handler.ServeHTTP(writer, request)
}
