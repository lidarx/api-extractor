package api_extractor

import (
	"fmt"
	"github.com/robertkrimen/otto/ast"
	"github.com/robertkrimen/otto/parser"
	"github.com/thoas/go-funk"
	"sync"
)

func Walk(v ast.Visitor, n ast.Node) {
	if n == nil {
		return
	}
	if v = v.Enter(n); v == nil {
		return
	}

	defer v.Exit(n)

	switch n := n.(type) {
	case *ast.ArrayLiteral:
		if n != nil {
			for _, ex := range n.Value {
				Walk(v, ex)
			}
		}
	case *ast.AssignExpression:
		if n != nil {
			Walk(v, n.Left)
			Walk(v, n.Right)
		}
	case *ast.BadExpression:
	case *ast.BinaryExpression:
		if n != nil {
			Walk(v, n.Left)
			Walk(v, n.Right)
		}
	case *ast.BlockStatement:
		if n != nil {
			for _, s := range n.List {
				Walk(v, s)
			}
		}
	case *ast.BooleanLiteral:
	case *ast.BracketExpression:
		if n != nil {
			Walk(v, n.Left)
			Walk(v, n.Member)
		}
	case *ast.BranchStatement:
		if n != nil {
			Walk(v, n.Label)
		}
	case *ast.CallExpression:
		if n != nil {
			Walk(v, n.Callee)
			for _, a := range n.ArgumentList {
				Walk(v, a)
			}
		}
	case *ast.CaseStatement:
		if n != nil {
			Walk(v, n.Test)
			for _, c := range n.Consequent {
				Walk(v, c)
			}
		}
	case *ast.CatchStatement:
		if n != nil {
			Walk(v, n.Parameter)
			Walk(v, n.Body)
		}
	case *ast.ConditionalExpression:
		if n != nil {
			Walk(v, n.Test)
			Walk(v, n.Consequent)
			Walk(v, n.Alternate)
		}
	case *ast.DebuggerStatement:
	case *ast.DoWhileStatement:
		if n != nil {
			Walk(v, n.Test)
			Walk(v, n.Body)
		}
	case *ast.DotExpression:
		if n != nil {
			Walk(v, n.Left)
			Walk(v, n.Identifier)
		}
	case *ast.EmptyExpression:
	case *ast.EmptyStatement:
	case *ast.ExpressionStatement:
		if n != nil {
			Walk(v, n.Expression)
		}
	case *ast.ForInStatement:
		if n != nil {
			Walk(v, n.Into)
			Walk(v, n.Source)
			Walk(v, n.Body)
		}
	case *ast.ForStatement:
		if n != nil {
			Walk(v, n.Initializer)
			Walk(v, n.Update)
			Walk(v, n.Test)
			Walk(v, n.Body)
		}
	case *ast.FunctionLiteral:
		if n != nil {
			Walk(v, n.Name)
			for _, p := range n.ParameterList.List {
				Walk(v, p)
			}
			Walk(v, n.Body)
		}
	case *ast.FunctionStatement:
		if n != nil {
			Walk(v, n.Function)
		}
	case *ast.Identifier:
	case *ast.IfStatement:
		if n != nil {
			Walk(v, n.Test)
			Walk(v, n.Consequent)
			Walk(v, n.Alternate)
		}
	case *ast.LabelledStatement:
		if n != nil {
			Walk(v, n.Label)
			Walk(v, n.Statement)
		}
	case *ast.NewExpression:
		if n != nil {
			Walk(v, n.Callee)
			for _, a := range n.ArgumentList {
				Walk(v, a)
			}
		}
	case *ast.NullLiteral:
	case *ast.NumberLiteral:
	case *ast.ObjectLiteral:
		if n != nil {
			for _, p := range n.Value {
				Walk(v, p.Value)
			}
		}
	case *ast.Program:
		if n != nil {
			for _, b := range n.Body {
				Walk(v, b)
			}
		}
	case *ast.RegExpLiteral:
	case *ast.ReturnStatement:
		if n != nil {
			Walk(v, n.Argument)
		}
	case *ast.SequenceExpression:
		if n != nil {
			for _, e := range n.Sequence {
				Walk(v, e)
			}
		}
	case *ast.StringLiteral:
	case *ast.SwitchStatement:
		if n != nil {
			Walk(v, n.Discriminant)
			for _, c := range n.Body {
				Walk(v, c)
			}
		}
	case *ast.ThisExpression:
	case *ast.BadStatement:
	case *ast.ThrowStatement:
		if n != nil {
			Walk(v, n.Argument)
		}
	case *ast.TryStatement:
		if n != nil {
			Walk(v, n.Body)
			Walk(v, n.Catch)
			Walk(v, n.Finally)
		}
	case *ast.UnaryExpression:
		if n != nil {
			Walk(v, n.Operand)
		}
	case *ast.VariableExpression:
		if n != nil {
			Walk(v, n.Initializer)
		}
	case *ast.VariableStatement:
		if n != nil {
			for _, e := range n.List {
				Walk(v, e)
			}
		}
	case *ast.WhileStatement:
		if n != nil {
			Walk(v, n.Test)
			Walk(v, n.Body)
		}
	case *ast.WithStatement:
		if n != nil {
			Walk(v, n.Object)
			Walk(v, n.Body)
		}
	default:
		panic(fmt.Sprintf("Walk: unexpected node type %T", n))
	}
}

func NewExtractor() *Extractor {
	return &Extractor{seen: make(map[ast.Node]struct{})}
}

func (v *Extractor) Extract(js string) (apis []string, err error) {
	program, err := parser.ParseFile(nil, "", js, 0)
	if err != nil {
		return nil, err
	}
	Walk(v, program)
	return v.GetAPIs(), nil
}

// Extractor ast遍历结构体
type Extractor struct {
	stack []ast.Node
	//source    string
	duplicate int
	seen      map[ast.Node]struct{}
	apiSet    []string
	uniqueAPI []string
	l         sync.Mutex
}

func (v *Extractor) GetAPIs() []string {
	if len(v.uniqueAPI) == 0 {
		v.uniqueAPI = funk.UniqString(v.apiSet)
	}
	return v.uniqueAPI
}

// Enter 遍历时执行方法
func (v *Extractor) Enter(n ast.Node) ast.Visitor {
	v.push(n)
	if _, ok := v.seen[n]; ok {
		v.duplicate++
		return v
	}

	v.seen[n] = struct{}{}

	apis := []string{}
	switch node := n.(type) {
	// 函数调用解析
	case *ast.CallExpression:
		apis = extractAPIFromCallExpression(node)
	case *ast.ObjectLiteral:
		apis = extractAPIFromObjectLiteral(node)
	}
	if len(apis) > 0 {
		v.l.Lock()
		v.apiSet = append(v.apiSet, apis...)
		v.l.Unlock()
	}
	return v
}

func (v *Extractor) Exit(n ast.Node) {
	v.pop(n)
}

func (v *Extractor) push(n ast.Node) {
	v.stack = append(v.stack, n)
}

func (v *Extractor) pop(n ast.Node) {
	size := len(v.stack)
	if size <= 0 {
		panic("pop of empty stack")
	}

	toPop := v.stack[size-1]
	if toPop != n {
		panic("pop: nodes do not equal")
	}

	v.stack[size-1] = nil
	v.stack = v.stack[:size-1]
}
