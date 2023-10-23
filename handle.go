package gow

import (
	"alicode.mukj.cn/yjkj.ink/work/utils/showdoc"
	"fmt"
	"github.com/zituocn/gow"
	"reflect"
	"runtime"
	"strings"
)

type Engine struct {
	projectName string
	engine      *gow.Engine
	groups      []*RouterGroup
	handlers    []*HandlerFunc
}

type HandlerFunc struct {
	router   string
	method   string
	function gow.HandlerFunc
	request  interface{}
	response interface{}
}

type RouterGroup struct {
	basePath    string
	routerGroup *gow.RouterGroup
	groups      []*RouterGroup
	handlers    []*HandlerFunc
}

func Default(projectName string) *Engine {
	return &Engine{projectName: projectName, engine: gow.Default()}
}

func (engine *Engine) Handle(httpMethod, relativePath string, request, response interface{}, handlers ...gow.HandlerFunc) gow.IRoutes {
	for _, handler := range handlers {
		engine.handlers = append(engine.handlers, &HandlerFunc{method: httpMethod, function: handler, router: relativePath, request: request, response: response})
	}
	return engine.engine.Handle(httpMethod, relativePath, handlers...)
}
func (engine *Engine) Group(path string, request, response interface{}, handlers ...gow.HandlerFunc) *RouterGroup {
	for _, handler := range handlers {
		engine.handlers = append(engine.handlers, &HandlerFunc{function: handler, router: path, request: request, response: response})
	}
	group := &RouterGroup{basePath: path, routerGroup: engine.engine.Group(path, handlers...)}
	engine.groups = append(engine.groups, group)
	return group
}

func (group *RouterGroup) Group(path string, request, response interface{}, handlers ...gow.HandlerFunc) *RouterGroup {
	for _, handler := range handlers {
		group.handlers = append(group.handlers, &HandlerFunc{function: handler, router: path, request: request, response: response})
	}
	group2 := &RouterGroup{basePath: path, routerGroup: group.routerGroup.Group(path, handlers...)}
	group.groups = append(group.groups, group2)
	return group2
}

func (group *RouterGroup) Handle(httpMethod, relativePath string, request, response interface{}, handlers ...gow.HandlerFunc) gow.IRoutes {
	for _, handler := range handlers {
		group.handlers = append(group.handlers, &HandlerFunc{method: httpMethod, function: handler, router: relativePath, request: request, response: response})
	}
	return group.routerGroup.Handle(httpMethod, relativePath, handlers...)
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

	return engine.engine.Run(args...)
}
