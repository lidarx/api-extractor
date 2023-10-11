package api_extractor

import (
	"github.com/robertkrimen/otto/ast"
	"regexp"
	"strings"
)

var funcNameReg = regexp.MustCompile(`(?i)POST|GET`)
var urlReg = regexp.MustCompile(`(?i)^/([\w-./?%&=#]*)+$`)
var wordReg = regexp.MustCompile(`(?i)\w+`)

func extractAPIFromCallExpression(exp *ast.CallExpression) (apis []string) {
	args := exp.ArgumentList
	if o, ok := exp.Callee.(*ast.DotExpression); ok && funcNameReg.MatchString(o.Identifier.Name) {
		for _, each := range args {
			// 类型断言取参数值
			if o, ok := each.(*ast.StringLiteral); ok {
				if urlReg.MatchString(o.Value) && wordReg.MatchString(o.Value) && strings.Contains(o.Value, "/") {
					apis = append(apis, o.Value)
				}
			}
		}
	}
	return
}

func extractAPIFromObjectLiteral(objlist *ast.ObjectLiteral) (apis []string) {
	for _, each := range objlist.Value {
		if each.Key == "component" || each.Key == "redirect" || each.Key == "to" {
			return
		}

		if o, ok := each.Value.(*ast.StringLiteral); ok && urlReg.MatchString(o.Value) && wordReg.MatchString(o.Value) && strings.Contains(o.Value, "/") {
			apis = append(apis, o.Value)
		}
	}
	return
}
