package web

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type Server interface {

}

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

func AttachFile(mux *http.ServeMux, path string, filename string) error {
	handler, err := StaticHandlerFromFile(filename)
	if err != nil {
		return err
	}
	mux.Handle(path, handler)
	return nil
}

func CreateMux(s Server) (*http.ServeMux, error) {
	mux := http.NewServeMux()
	err := AttachFile(mux, "/style.css", "resources/style.css")
	if err != nil {
		return nil, err
	}
	err = AttachFile(mux, "/client.js", "resources/client.js")
	if err != nil {
		return nil, err
	}
	err = AttachFile(mux, "/", "resources/client.html")
	if err != nil {
		return nil, err
	}
	return mux, nil
}

func LaunchHTTP(s Server) error {
	mux, err := CreateMux(s)
	if err != nil {
		return err
	}
	println("launching server...")
	return http.ListenAndServe(":8080", mux)
}
