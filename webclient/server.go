package webclient

import (
	"bytes"
	"encoding/json"
	"github.com/pkg/errors"
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

func AttachResources(mux *http.ServeMux, path string, resources []string) error {
	modTime := time.Now()
	resourcesJson, err := json.Marshal(resources)
	if err != nil {
		return err
	}
	resourcesJs := []byte("resources = " + string(resourcesJson) + ";")
	mux.HandleFunc(path, func(writer http.ResponseWriter, request *http.Request) {
		http.ServeContent(writer, request, path, modTime, bytes.NewReader(resourcesJs))
	})
	return nil
}

func CreateMux(api ServerAPI) (*http.ServeMux, error) {
	mux := http.NewServeMux()
	coreResources := api.CoreResourcePath()
	err := AttachFile(mux, "/style.css", path.Join(coreResources, "style.css"))
	if err != nil {
		return nil, err
	}
	err = AttachFile(mux, "/client.js", path.Join(coreResources, "client.js"))
	if err != nil {
		return nil, err
	}
	err = AttachFile(mux, "/", path.Join(coreResources, "client.html"))
	if err != nil {
		return nil, err
	}
	resources, download, err := api.ListResources()
	if err != nil {
		return nil, errors.Wrap(err, "collecting resources")
	}
	for name, resource := range resources {
		err = AttachFile(mux, "/resource/"+name, resource)
		if err != nil {
			return nil, errors.Wrap(err, "attaching resources")
		}
	}
	err = AttachResources(mux, "/resources.js", download)
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
		return errors.Wrap(err, "preparing server")
	}
	println("launching server...")
	return http.ListenAndServe(":8080", mux)
}
