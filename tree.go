package gorouter

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/vardius/gorouter/v4/middleware"
	"github.com/vardius/gorouter/v4/mux"
	pathutils "github.com/vardius/gorouter/v4/path"
)

func addMiddleware(t mux.Tree, method, path string, mid middleware.Middleware) {
	fmt.Printf("tree: %v\n", t.PrettyPrint())
	type recFunc func(recFunc, mux.Node, middleware.Middleware)

	c := func(c recFunc, n mux.Node, m middleware.Middleware) {
		fmt.Printf("route handler: %v\n", n.Route())
		if n.Route() != nil {
			fmt.Println("append")
			n.Route().AppendMiddleware(m, path)
		}
		for _, child := range n.Tree() {
			fmt.Printf("child: %v\n", child)
			c(c, child, m)
		}
	}

	// routes tree roots should be http method nodes only
	if root := t.Find(method); root != nil {
		if path != "" {
			node := findNode(root, strings.Split(pathutils.TrimSlash(path), "/"))
			if node != nil {
				c(c, node, mid)
			}
		} else {
			c(c, root, mid)
		}
	}
}

func findNode(n mux.Node, parts []string) mux.Node {
	if len(parts) == 0 {
		return n
	}

	name, _ := pathutils.GetNameFromPart(parts[0])

	if node := n.Tree().Find(name); node != nil {
		return findNode(node, parts[1:])
	}

	return n
}

func allowed(t mux.Tree, method, path string) (allow string) {
	if path == "*" {
		// routes tree roots should be http method nodes only
		for _, root := range t {
			if root.Name() == http.MethodOptions {
				continue
			}
			if len(allow) == 0 {
				allow = root.Name()
			} else {
				allow += ", " + root.Name()
			}
		}
	} else {
		// routes tree roots should be http method nodes only
		for _, root := range t {
			if root.Name() == method || root.Name() == http.MethodOptions {
				continue
			}

			if n, _, _ := root.Tree().Match(path); n != nil && n.Route() != nil {
				if len(allow) == 0 {
					allow = root.Name()
				} else {
					allow += ", " + root.Name()
				}
			}
		}
	}
	if len(allow) > 0 {
		allow += ", " + http.MethodOptions
	}
	return allow
}
