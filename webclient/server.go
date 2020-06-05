package webclient

import (
	"bytes"
	"encoding/json"
	"github.com/celskeggs/mediator/resourcepack"
	"github.com/celskeggs/mediator/util"
	"github.com/pkg/errors"
	"net/http"
	"path"
	"time"
)

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

func AttachResource(mux *http.ServeMux, path string, resource resourcepack.Resource) error {
	handler := StaticHandler(resource.Modified, resource.Name, resource.Data)
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
	pack := api.ResourcePack()
	clientResource, err := pack.Resource("client.html")
	if err != nil {
		return nil, err
	}
	if err := AttachResource(mux, "/", clientResource); err != nil {
		return nil, err
	}
	for _, resource := range pack.Resources {
		util.FIXME("clean up this resource handling code")
		if resource.IsMap() {
			continue
		}
		basepath := "/resource/"
		if resource.IsWeb() {
			basepath = "/"
		}
		if err := AttachResource(mux, path.Join(basepath, resource.Name), resource); err != nil {
			return nil, err
		}
	}
	err = AttachResources(mux, "/resources.js", pack.Icons())
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
	println("launching server... visit http://localhost:8080/")
	return http.ListenAndServe(":8080", mux)
}
