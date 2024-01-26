package router

import "strings"

type Router struct {
	routes  map[string]interface{}
	reverse map[string]string
}

type Param struct {
	Key   string
	Value string
}

type Params []Param

func (p Params) ByName(name string) string {
	for _, param := range p {
		if param.Key == name {
			return param.Value
		}
	}
	return ""
}

// NewUrlManager creates a new UrlManager instance.
func New() *Router {
	return &Router{
		routes:  make(map[string]interface{}),
		reverse: make(map[string]string),
	}
}

// Add adds a URL pattern to the manager.
func (r *Router) Add(pattern, method string, call interface{}, name string) {
	method = strings.ToUpper(method)

	if !strings.HasSuffix(pattern, "/") {
		pattern += "/"
	}
	parts := strings.Split(pattern[1:], "/")
	node := r.routes
	for _, part := range parts {
		if _, ok := node[part]; !ok {
			node[part] = make(map[string]interface{})
		}
		node = node[part].(map[string]interface{})
	}

	if method == "" {
		node["GET"] = call
	} else {
		methods := strings.Split(method, ",")
		for _, m := range methods {
			node[strings.ToUpper(m)] = call
		}
	}

	if name != "" {
		r.reverse[name] = pattern
	}
}

// FindCall finds the callable for the specified URL path and HTTP method.
func (r *Router) FindCall(path, method string) (interface{}, Params) {
	method = strings.ToUpper(method)

	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	parts := strings.Split(path[1:], "/")
	return r.recursiveRouteMatch(r.routes, parts, method, nil)
}

// FindPattern finds the URL pattern for the specified name.
func (r *Router) FindPattern(name string) string {
	return r.reverse[name]
}

func (r *Router) recursiveRouteMatch(node map[string]interface{}, remaining []string, method string, params Params) (interface{}, Params) {
	if len(remaining) == 0 {
		if call, ok := node[method]; ok {
			return call, params
		}
		return nil, nil
	}

	var result interface{}
	for key, value := range node {
		if key == remaining[0] {
			result, params = r.recursiveRouteMatch(value.(map[string]interface{}), remaining[1:], method, params)
			if result != nil {
				return result, params
			}
		} else if len(key) > 0 && key[0] == '{' {
			continue
		}
	}

	for key, value := range node {
		if len(key) > 0 && key[0] == '{' {
			result, params = r.recursiveRouteMatch(
				value.(map[string]interface{}),
				remaining[1:],
				method,
				append(params, Param{Key: key[1 : len(key)-1], Value: remaining[0]}),
			)
			if result != nil {
				return result, params
			}
		} else if key == "*" {
			continue
		}
	}

	for key, value := range node {
		if key == "*" {
			result, params = r.recursiveRouteMatch(
				value.(map[string]interface{})[""].(map[string]interface{}), nil, method, params)
			if result != nil {
				return result, params
			}
		}
	}

	return nil, nil
}
