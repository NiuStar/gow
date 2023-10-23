package gow

import (
	"alicode.mukj.cn/yjkj.ink/work/apidoc"
	"alicode.mukj.cn/yjkj.ink/work/utils/showdoc"
	"fmt"
	"github.com/zituocn/gow"
	"net/url"
	"reflect"
	"runtime"
	"strings"
)

type HandlerFunc func(ctx *Context)
type Context struct {
	*gow.Context
}
type Engine struct {
	postmanDoc  *apidoc.PostmanDoc
	projectName string
	engine      *gow.Engine
	groups      []*RouterGroup
	handlers    []*HandlerFuncStruct
}

type HandlerFuncStruct struct {
	router   string
	method   string
	function HandlerFunc
	request  interface{}
	response interface{}
}

type RouterGroup struct {
	basePath    string
	routerGroup *gow.RouterGroup
	groups      []*RouterGroup
	handlers    []*HandlerFuncStruct
}

func Default(projectName string) *Engine {
	return &Engine{projectName: projectName, engine: gow.Default()}
}

func (engine *Engine) Handle(httpMethod, relativePath string, request, response interface{}, handlers ...HandlerFunc) gow.IRoutes {
	for _, handler := range handlers {
		engine.handlers = append(engine.handlers, &HandlerFuncStruct{method: httpMethod, function: handler, router: relativePath, request: request, response: response})
	}
	handlers2 := make([]gow.HandlerFunc, len(handlers))
	for i, handler := range handlers {
		handlers2[i] = func(ctx *gow.Context) {
			handler(&Context{ctx})
		}
	}
	return engine.engine.Handle(httpMethod, relativePath, handlers2...)
}
func (engine *Engine) Group(path string, request, response interface{}, handlers ...HandlerFunc) *RouterGroup {
	for _, handler := range handlers {
		engine.handlers = append(engine.handlers, &HandlerFuncStruct{function: handler, router: path, request: request, response: response})
	}
	handlers2 := make([]gow.HandlerFunc, len(handlers))
	for i, handler := range handlers {
		handlers2[i] = func(ctx *gow.Context) {
			handler(&Context{ctx})
		}
	}
	group := &RouterGroup{basePath: path, routerGroup: engine.engine.Group(path, handlers2...)}
	engine.groups = append(engine.groups, group)
	return group
}

func (group *RouterGroup) Group(path string, request, response interface{}, handlers ...HandlerFunc) *RouterGroup {
	for _, handler := range handlers {
		group.handlers = append(group.handlers, &HandlerFuncStruct{function: handler, router: path, request: request, response: response})
	}
	handlers2 := make([]gow.HandlerFunc, len(handlers))
	for i, handler := range handlers {
		handlers2[i] = func(ctx *gow.Context) {
			handler(&Context{ctx})
		}
	}
	group2 := &RouterGroup{basePath: path, routerGroup: group.routerGroup.Group(path, handlers2...)}
	group.groups = append(group.groups, group2)
	return group2
}

func (group *RouterGroup) Handle(httpMethod, relativePath string, request, response interface{}, handlers ...HandlerFunc) gow.IRoutes {
	for _, handler := range handlers {
		group.handlers = append(group.handlers, &HandlerFuncStruct{method: httpMethod, function: handler, router: relativePath, request: request, response: response})
	}
	handlers2 := make([]gow.HandlerFunc, len(handlers))
	for i, handler := range handlers {
		handlers2[i] = func(ctx *gow.Context) {
			handler(&Context{ctx})
		}
	}
	return group.routerGroup.Handle(httpMethod, relativePath, handlers2...)
}

func (group *RouterGroup) writeShowdoc(domain, prefix string) {

	if strings.HasSuffix(prefix, "/") {
		prefix = strings.TrimSuffix(prefix, "/")
	}
	prefix += group.basePath + "/"
	if strings.HasSuffix(prefix, "/") {
		prefix = strings.TrimSuffix(prefix, "/")
	}

	for _, handler := range group.handlers {
		name := runtime.FuncForPC(reflect.ValueOf(handler.function).Pointer()).Name()

		fmt.Println(name)
		WriteToApiMarkDown(domain, handler.method, name, strings.TrimSuffix(prefix+handler.router, "/"), strings.TrimSuffix(prefix, "/"), reflect.TypeOf(handler.request).Elem(), reflect.TypeOf(handler.response).Elem())
	}
	for _, group2 := range group.groups {
		group2.writeShowdoc(domain, prefix)
	}
}

func (engine *Engine) Run(userName, password string, args ...interface{}) error {
	uri := `https://showdoc.ai00.xyz/`
	err := showdoc.Instance().Login(userName, password, uri)
	if err != nil {
		fmt.Println("showdoc login err:", err)
	} else {
		showdoc.Instance().CreateApiKey(engine.projectName)

		for _, handler := range engine.handlers {
			WriteToApiMarkDown(uri, handler.method, reflect.TypeOf(handler.function).Name(), handler.router, "", reflect.TypeOf(handler.request).Elem(), reflect.TypeOf(handler.response).Elem())

		}
		for _, group := range engine.groups {
			group.writeShowdoc(uri, "")
		}
	}
	engine.WritePostman(uri)
	return engine.engine.Run(args...)
}
func (group *RouterGroup) WritePostman(postmanDoc *apidoc.PostmanDoc, domain, prefix string) {

	if strings.HasSuffix(prefix, "/") {
		prefix = strings.TrimSuffix(prefix, "/")
	}
	prefix += group.basePath + "/"
	if strings.HasSuffix(prefix, "/") {
		prefix = strings.TrimSuffix(prefix, "/")
	}

	for _, handler := range group.handlers {
		name := runtime.FuncForPC(reflect.ValueOf(handler.function).Pointer()).Name()
		method := "GET"
		if strings.Contains(handler.method, "POST") {
			method = "POST"
		}
		uri, _ := url.Parse(domain)
		uri.Path = strings.TrimSuffix(prefix+handler.router, "/")
		fmt.Println(name)
		//WriteToApiMarkDown(domain, handler.method, name, strings.TrimSuffix(prefix+handler.router, "/"), strings.TrimSuffix(prefix, "/"), reflect.TypeOf(handler.request).Elem(), reflect.TypeOf(handler.response).Elem())
		postmanDoc.WriteItem(name, strings.TrimSuffix(prefix, "/"), method, uri.String(), apidoc.Json, nil, handler.request)
	}
	for _, group2 := range group.groups {
		group2.WritePostman(postmanDoc, domain, prefix)
	}
}
func (engine *Engine) WritePostman(domain string) error {
	engine.postmanDoc = apidoc.NewPostmanDoc(engine.projectName)

	for _, handler := range engine.handlers {
		name := runtime.FuncForPC(reflect.ValueOf(handler.function).Pointer()).Name()

		method := "GET"
		if strings.Contains(handler.method, "POST") {
			method = "POST"
		}
		uri, _ := url.Parse(domain)
		uri.Path = handler.router
		engine.postmanDoc.WriteItem(name, "", method, uri.String(), apidoc.Json, nil, handler.request)

	}
	for _, group := range engine.groups {
		group.WritePostman(engine.postmanDoc, domain, "")
	}
	engine.postmanDoc.Save(engine.projectName + "_postman.json")
	return nil
}
