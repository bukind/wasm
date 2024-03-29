// Package wasm is a WASM helper for Go.
package wasm

import (
	"errors"
	"fmt"
	"strings"
	"syscall/js"
)

// DocHolder is a main entry object.  Create it by calling GetDoc().
type DocHolder struct {
	doc js.Value
}

func GetDoc() (*DocHolder, error) {
	doc := js.Global().Get("document")
	if !doc.Truthy() {
		return nil, errors.New("cannot get document")
	}
	return &DocHolder{doc}, nil
}

func (g *DocHolder) GetElementByID(id string) (js.Value, error) {
	elt := g.doc.Call("getElementById", id)
	if !elt.Truthy() {
		return js.Undefined(), fmt.Errorf("cannot find elt with id %q", id)
	}
	return elt, nil
}

func (g *DocHolder) CreateElement(typ string) js.Value {
	return g.doc.Call("createElement", typ)
}

// EventListener is an event listener.
type EventListener struct {
	name string
	fn   js.Func
}

// NewEventListener creates a new event listener.
func NewEventListener(evt string, fn func(js.Value, js.Value) any) *EventListener {
	return &EventListener{
		name: evt,
		fn: js.FuncOf(func(this js.Value, args []js.Value) any {
			if !this.Truthy() {
				fmt.Printf("event %q this is not truthy\n", evt)
				return nil
			}
			if len(args) != 1 {
				fmt.Printf("event %q len(args)=%d\n", evt, len(args))
				return nil
			}
			if !args[0].Truthy() {
				fmt.Printf("event %q arg[0] is not truthy\n", evt)
				return nil
			}
			fmt.Printf("Event %q called on %s evt=%s target=%s\n", evt, Dbg(this), Dbg(args[0]), Dbg(args[0].Get("target")))
			return fn(this, args[0])
		}),
	}
}

// Add adds the event listener to a JS object.
func (e *EventListener) Add(elt js.Value) {
	fmt.Printf("Adding event listener %q to %s\n", e.name, Dbg(elt))
	elt.Call("addEventListener", e.name, e.fn)
}

// Remove removes the event listener from a JS object.
func (e *EventListener) Remove(elt js.Value) {
	fmt.Printf("Removing event listener %q from %s\n", e.name, Dbg(elt))
	elt.Call("removeEventListener", e.name, e.fn)
}

// ClassList represents CSS class list of an element.
type ClassList js.Value

// GetClassList returns a ClassList of the object.
func GetClassList(obj js.Value) ClassList {
	clist := obj.Get("classList")
	if clist.Type() == js.TypeUndefined {
		return ClassList(clist)
	}
	if len := clist.Get("length"); len.Type() == js.TypeNumber && len.Int() > 0 {
		return ClassList(clist)
	}
	return ClassList(js.Undefined())
}

// Contains returns true if a class list contains the specific class.
func (c ClassList) Contains(s string) bool {
	clist := js.Value(c)
	if clist.Type() == js.TypeUndefined {
		return false
	}
	return clist.Call("contains", s).Bool()
}

// Add adds the class to the class list.
func (c ClassList) Add(s string) {
	clist := js.Value(c)
	if clist.Type() == js.TypeUndefined {
		return
	}
	clist.Call("add", s)
}

// Remove removes the class from the class list.
func (c ClassList) Remove(s string) {
	clist := js.Value(c)
	if clist.Type() == js.TypeUndefined {
		return
	}
	clist.Call("remove", s)
}

// String returns the string representation of the class list.
func (c ClassList) String() string {
	clist := js.Value(c)
	if clist.Type() == js.TypeUndefined {
		return ""
	}
	return js.Value(c).Get("value").String()
}

// Dbg returns a human-readable representation of a js.Value, useful for debugging.
func Dbg(v js.Value) string {
	switch v.Type() {
	case js.TypeObject:
		sb := &strings.Builder{}
		sb.WriteString("<obj")
		if id := v.Get("id"); id.Type() != js.TypeUndefined && id.String() != "" {
			fmt.Fprintf(sb, " id=%s", id)
		}
		if typ := v.Get("type"); typ.Type() != js.TypeUndefined {
			fmt.Fprintf(sb, " type=%s", typ)
		}
		clist := GetClassList(v)
		if cls := clist.String(); cls != "" {
			fmt.Fprintf(sb, " cls=%s", cls)
		}
		sb.WriteString(">")
		return sb.String()
	default:
		return v.String()
	}
}
