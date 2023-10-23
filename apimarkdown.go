package gow

import (
	"alicode.mukj.cn/yjkj.ink/work/markdown"
	"alicode.mukj.cn/yjkj.ink/work/utils/showdoc"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strings"
)

func WriteToApiMarkDown(domain, method, methodName, router string, notes string, request reflect.Type, response reflect.Type) error {
	apimk := markdown.NewMarkDown(notes, "")
	parms(apimk, domain, method, methodName, router, notes, request, response)
	api2 := &showdoc.API{}
	api2.CatName = notes
	api2.PageContent = apimk.Content()
	api2.PageTitle = methodName
	err := showdoc.Instance().WriteToApiMarkDown(api2)
	if err != nil {
		//logrus.Warningln()
		//fmt.Println("添加接口文档失败，请检查showdoc服务是否开启，可使用docker start showdoc命令开启")
		return errors.New("添加接口文档失败，请检查showdoc服务是否开启，可使用docker start showdoc命令开启:" + err.Error())
	}
	return nil
}

func getTags(tag string) map[string]string {
	pairs := strings.Split(tag, " ")
	tags := make(map[string]string)
	for _, pair := range pairs {
		parts := strings.Split(pair, ":")
		if len(parts) == 2 {
			key := parts[0]
			value := strings.Trim(parts[1], `"`)
			tags[key] = value
		}
	}
	return tags
}
func parms(mk *markdown.MarkDown, domain, method, methodName, router string, notes string, request reflect.Type, response reflect.Type) {
	{
		mk.WriteTitle(3, method+"方式："+methodName+"  "+notes+"\r\n")
		mk.WriteTitle(4, "http 调用方法\r\n")
		uri, _ := url.Parse(domain)
		uri.Path = router
		mk.WriteCode("URL:  "+uri.String()+"\r\n", "go")
		//mk.WriteCode("URL:  "+domain+router+"\r\n", "go")
	}

	{
		mk.WriteContent("\r\n请求参数：\r\n")
		var paramsKeys = make([]string, 0)
		paramsKeysMap := make(map[string]interface{})
		paramsTags := parseParamTags(request, 0)

		for _, paramTags := range paramsTags {
			for key, _ := range paramTags.Tags {
				if paramsKeysMap[key] == nil {
					paramsKeys = append(paramsKeys, key)
					paramsKeysMap[key] = &key
				}
			}
		}
		var params [][]string
		{
			var param []string
			param = append(param, "参数名")
			for _, key := range paramsKeys {
				param = append(param, key)
			}
			params = append(params, param)
		}

		////fmt.Println("params_list:",params_list)
		{
			for _, paramTags := range paramsTags {
				params = append(params, paramTags.Get(paramsKeys))
			}
			mk.WriteForm(params)
		}
	}

	{
		mk.WriteContent("\r\n返回值：\r\n")
		var paramsKeys = make([]string, 0)
		paramsKeysMap := make(map[string]interface{})
		paramsTags := parseParamTags(response, 0)

		for _, paramTags := range paramsTags {
			for key, _ := range paramTags.Tags {
				if paramsKeysMap[key] == nil {
					paramsKeys = append(paramsKeys, key)
					paramsKeysMap[key] = &key
				}
			}
		}
		var params [][]string
		{
			var param []string
			param = append(param, "参数名")
			for _, key := range paramsKeys {
				param = append(param, key)
			}
			params = append(params, param)
		}

		////fmt.Println("params_list:",params_list)
		{
			for _, paramTags := range paramsTags {
				params = append(params, paramTags.Get(paramsKeys))
			}
			mk.WriteForm(params)
		}
	}
}

type ParamTags struct {
	Name   string
	Tags   map[string]string
	Indexs []int
	Index  int
}

func (param *ParamTags) Get(keys []string) []string {
	params := make([]string, 0)
	prefix := ""
	for i := 0; i < len(param.Indexs); i++ {
		prefix += "--"
	}
	params = append(params, prefix+param.Name)

	for _, key := range keys {
		if param.Tags[key] != "" {
			params = append(params, param.Tags[key])
		} else {
			params = append(params, "")
		}
	}
	return params
}
func parseParamTags(request reflect.Type, index int, indexs ...int) (list []*ParamTags) {
	indexs2 := make([]int, 0)
	indexs2 = append(indexs2, indexs...)
	indexs2 = append(indexs2, index)
	for i := 0; i < request.NumField(); i++ {
		paramTags := &ParamTags{}
		list3 := make([]*ParamTags, 0)
		var paramsKeyMap = make(map[string]string)
		//判断成员变量是否是对外可见的
		if request.Field(i).Tag.Get("json") != "-" && request.Field(i).IsExported() {
			// 现在我们不知道tag的名称，所以我们遍历整个tag内容
			tags := getTags(string(request.Field(i).Tag))
			for key, value := range tags {
				paramsKeyMap[key] = value
			}
			fieldType := request.Field(i).Type
			if fieldType.Kind() == reflect.Pointer {
				fieldType = fieldType.Elem()
			}
			if fieldType.Kind() == reflect.Struct {
				list2 := parseParamTags(fieldType, i, indexs2...)
				list3 = append(list3, list2...)
			}
		}
		paramTags.Name = request.Field(i).Name
		paramTags.Tags = paramsKeyMap
		paramTags.Indexs = indexs2
		for _, index3 := range indexs {
			paramTags.Index *= 10
			paramTags.Index += index3
		}
		list = append(list, paramTags)
		list = append(list, list3...)
	}

	return
}

func POSTJson(url string, payload interface{}, headers map[string]string) (string, error) {
	client := &http.Client{}

	var payload_body []byte
	switch payload.(type) {
	case string:
		{

			payload_body = []byte(payload.(string))
		}
	default:
		payload_body, _ = json.Marshal(payload)
	}

	////fmt.Println("payload:",string(payload_body))
	req, err := http.NewRequest("POST", url, strings.NewReader(string(payload_body)))

	if err != nil {
		//fmt.Println(err)
		return "", err
	}
	req.Header.Add("Content-Type", "application/json")
	//req.Header.Add("User-Agent", "Xcode")
	//req.Header.Add("Accept", "text/x-xml-plist")

	for name, header := range headers {
		req.Header.Add(name, header)
	}

	res, err := client.Do(req)
	if err != nil {
		//fmt.Println("POST error",err)
		return "", err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		//fmt.Println("POST read error",err)
		return "", err
	}

	////fmt.Println("POSTJson return",string(body))
	return string(body), nil
}

type ScodiApi struct {
	CatName    string
	MethodName string
	Note       string
	Router     string
	IsGrpc     bool
	Export     bool
}

/*
func (manager *ServiceManager) MakeDoc(ctx *context.Context) (interface{}, error) {

		for _, method := range manager.HttpServiceManager.methods {
			//packageName := method.ServiceName[0]
			//ServiceName := method.ServiceName[1]

			if method.Service.IsValid() {
				if method.catName == "URL管理" {
					fmt.Println("ok")
				}
				if method.params != nil && len(method.params) > 0 {
					manager.addApiDoc2("", "", method.Pattern.String(), method.MethodName, method.noteName, method.catName, method.params,
						method.results, false, true, method.Service.Interface())
				} else {
					manager.addApiDoc("", "", method.Pattern.String(), method.MethodName, method.noteName, method.catName, method.requestValueType,
						method.responseValueType, false, true, method.Service.Interface())
				}

			} else {
				if method.params != nil && len(method.params) > 0 {
					manager.addApiDoc2("", "", method.Pattern.String(), method.MethodName, method.noteName, method.catName, method.params,
						method.results, false, true, nil)
				} else {
					manager.addApiDoc("", "", method.Pattern.String(), method.MethodName, method.noteName, method.catName, method.requestValueType,
						method.responseValueType, false, true, nil)
				}
			}

		}
		for _, method := range manager.HttpServiceManager.gmethods {
			//doc := getDoc(method.ServiceName)
			//fmt.Println("method.requestValueType",method.requestValueType)
			//fmt.Println("method.responseValueType",method.responseValueType)
			packageName := method.ServiceName[0]
			ServiceName := method.ServiceName[1]
			manager.addApiDoc(packageName, ServiceName, method.Pattern.String(), method.MethodName, method.noteName, method.catName,
				method.requestValueType, method.responseValueType, true, method.export, method.Service.Interface())
		}

		return "success", nil
	}

	func (manager *ServiceManager) MakePostmanDoc() (interface{}, error) {
		if config.ShareInstance().APIMarkdown.Domain == nil {
			config.ShareInstance().APIMarkdown.Domain = config.ShareInstance().APIMarkdown.ApiDomain
		}
		for _, method := range manager.HttpServiceManager.methods {
			router := strings.TrimPrefix(method.Pattern.String(), "/")
			rList := strings.Split(router, "/")

			servicename2 := ""
			if method.Service.IsValid() && !method.Service.IsZero() {
				servicename2 = getServiceName(method.Service.Interface(), true)
				rList[0] = servicename2
			}

			router2 := ""
			for _, r := range rList {
				router2 += "/" + r
			}
			apiPath := *config.ShareInstance().APIMarkdown.Domain + router2
			if method.requestValueType != nil {
				manager.postmanDoc.WriteItem(method.noteName, method.catName, method.Method, apiPath, apidoc.Json, nil, reflect.New(method.requestValueType).Elem().Interface())
			} else {
				manager.postmanDoc.WriteItem(method.noteName, method.catName, method.Method, apiPath, apidoc.Json, nil, nil)
			}

		}
		for _, method := range manager.HttpServiceManager.gmethods {
			router := strings.TrimPrefix(method.Pattern.String(), "/")
			rList := strings.Split(router, "/")
			servicename2 := getServiceName(method.Service.Interface(), true)
			rList[0] = servicename2

			router2 := ""
			for _, r := range rList {
				router2 += "/" + r
			}
			apiPath := *config.ShareInstance().APIMarkdown.Domain + router2
			manager.postmanDoc.WriteItem(method.noteName, method.catName, "POST", apiPath, apidoc.Json, nil, reflect.New(method.requestValueType).Elem().Interface())
		}

		return "success", nil
	}
*/

/*
func (manager *ServiceManager) addApiDoc(packageName, servicename, router, name, note, CatName string, request,
	response reflect.Type, grpc, export bool, srv interface{}) error {

	api := &API{}
	if config.ShareInstance().APIMarkdown.ApiName == nil {
		if config.ShareInstance().APIMarkdown.ApiKey != nil {
			api.ApiKey = *config.ShareInstance().APIMarkdown.ApiKey
		}
		if config.ShareInstance().APIMarkdown.ApiToken != nil {
			api.ApiToken = *config.ShareInstance().APIMarkdown.ApiToken
		}
	}

	api.CatName = CatName

	api.PageTitle = note
	if len(note) <= 0 {
		api.PageTitle = name
	}

	router = strings.TrimPrefix(router, "/")
	rList := strings.Split(router, "/")
	servicename2 := getServiceName(srv, true)
	rList[0] = servicename2

	router2 := ""
	for _, r := range rList {
		router2 += "/" + r
	}

	//logrus.Println("接口:  ", CatName, note, name)
	uri, _ := url.Parse(*config.ShareInstance().APIMarkdown.Domain)
	uri.Path = path.Join(uri.Path, router2)
	//logrus.Println("URL:  " + uri.String())

	entry := logrus.WithFields(logrus.Fields{
		"接口：":  fmt.Sprintf("%s %s %s", CatName, note, name),
		"URL：": uri.String(),
	})

	err := api.WriteToApiMarkDown(*config.ShareInstance().APIMarkdown.Domain, name, packageName, servicename, router2, note,
		request, response, grpc, export, srv)
	if err != nil {
		entry.Warningln(err.Error())
		return err
	}
	entry.Infoln()
	return nil
}

func (manager *ServiceManager) addApiDoc2(packageName, servicename, router, name, note, CatName string, request server.ParamsType,
	response []*server.Result, grpc, export bool, srv interface{}) error {

	api := &API{}
	if config.ShareInstance().APIMarkdown.ApiName == nil {
		if config.ShareInstance().APIMarkdown.ApiKey != nil {
			api.ApiKey = *config.ShareInstance().APIMarkdown.ApiKey
		}
		if config.ShareInstance().APIMarkdown.ApiToken != nil {
			api.ApiToken = *config.ShareInstance().APIMarkdown.ApiToken
		}
	}

	api.CatName = CatName

	api.PageTitle = note
	if len(note) <= 0 {
		api.PageTitle = name
	}

	router = strings.TrimPrefix(router, "/")
	rList := strings.Split(router, "/")
	servicename2 := getServiceName(srv, true)
	rList[0] = servicename2

	router2 := ""
	for _, r := range rList {
		router2 += "/" + r
	}
	logrus.Println("接口:  ", CatName, note, name)
	uri, _ := url.Parse(*config.ShareInstance().APIMarkdown.Domain)
	uri.Path = path.Join(uri.Path, router2)
	logrus.Println("URL:  " + uri.String())
	err := api.WriteToApiMarkDown2(*config.ShareInstance().APIMarkdown.Domain, name, packageName, servicename, router2, note,
		request, response, grpc, export, srv)
	if err != nil {
		logrus.Warningln(err.Error())
		return err
	}
	return nil
}
*/
