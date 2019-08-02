package web

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"time"
)

func StaticHandlerFromFile(filename string) (http.Handler, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(filename)
	if err != nil {
		return nil, err
	}
	return StaticHandler(info.ModTime(), filename, content), nil
}

func StaticHandler(modTime time.Time, nameForType string, data []byte) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		http.ServeContent(writer, request, nameForType, modTime, bytes.NewReader(data))
	})
}

func ExactMatchChecker(path string, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == path {
			handler.ServeHTTP(writer, request)
		} else {
			http.Error(writer, "Not Found", 404)
		}
	})
}

func AttachFile(mux *http.ServeMux, path string, filename string) error {
	handler, err := StaticHandlerFromFile(filename)
	if err != nil {
		return err
	}
	if path == "/" {
		handler = ExactMatchChecker("/", handler)
	}
	mux.Handle(path, handler)
	return nil
}

func AttachFolder(mux *http.ServeMux, basepath string, dirname string) error {
	files, err := ioutil.ReadDir(dirname)
	if err != nil {
		return err
	}
	for _, file := range files {
		if !file.IsDir() {
			err := AttachFile(mux, path.Join(basepath, file.Name()), path.Join(dirname, file.Name()))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func CreateMux(api ServerAPI) (*http.ServeMux, error) {
	mux := http.NewServeMux()
	err := AttachFile(mux, "/style.css", "resources/style.css")
	if err != nil {
		return nil, err
	}
	err = AttachFile(mux, "/client.js", "resources/client.js")
	if err != nil {
		return nil, err
	}
	err = AttachFolder(mux, "/img", "resources/img")
	if err != nil {
		return nil, err
	}
	err = AttachFile(mux, "/", "resources/client.html")
	if err != nil {
		return nil, err
	}
	wss := NewWebSocketServer(api)
	mux.Handle("/websocket", wss)
	return mux, nil
}

func LaunchHTTP(api ServerAPI) error {
	mux, err := CreateMux(api)
	if err != nil {
		return err
	}
	println("launching server...")
	return http.ListenAndServe(":8080", mux)
}
