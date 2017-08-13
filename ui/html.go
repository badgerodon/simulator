package main

import (
	"fmt"
	"strings"

	"github.com/gopherjs/gopherjs/js"
)

// An AttributeNode is an attribute on an element
type AttributeNode struct {
	name  string
	value interface{}
}

func (n AttributeNode) appendTo(parent ElementNode) {
	parent.object.Call("setAttribute", n.name, fmt.Sprint(n.value))
}

// A TextNode is a block of text
type TextNode struct {
	text string
}

func (n TextNode) appendTo(parent ElementNode) {
	parent.object.Call("appendChild", js.Global.Get("document").Call("createTextNode", n.text))
}

// An HTMLNode is a block of html
type HTMLNode struct {
	html string
}

func (n HTMLNode) appendTo(parent ElementNode) {
	parent.object.Call("insertAdjacentHTML", "beforeend", n.html)
}

// An ElementNode is an element
type ElementNode struct {
	object *js.Object
}

func (n ElementNode) appendTo(parent ElementNode) {
	parent.object.Call("appendChild", n.object)
}

// ReplaceWith replaces the element with the given element
func (n ElementNode) ReplaceWith(other ElementNode) {
	n.object.Get("parentNode").Call("replaceChild", other.object, n.object)
}

// A Node is a DOM node
type Node interface {
	appendTo(ElementNode)
}

// E creates an element
func E(name string, children ...Node) ElementNode {
	ids := strings.Split(name, "#")
	if len(ids) > 1 {
		children = append([]Node{
			A("id", strings.Join(ids[1:], " ")),
		}, children...)
		name = ids[0]
	}

	classes := strings.Split(name, ".")
	if len(classes) > 1 {
		children = append([]Node{
			A("class", strings.Join(classes[1:], " ")),
		}, children...)
		name = classes[0]
	}

	el := ElementNode{
		object: js.Global.Get("document").Call("createElement", name),
	}
	for _, child := range children {
		child.appendTo(el)
	}
	return el
}

// A creates an attribute
func A(name string, value interface{}) AttributeNode {
	return AttributeNode{name: name, value: value}
}

// T creates a text node
func T(text string) TextNode {
	return TextNode{text: text}
}

// H creates an html node
func H(html string) HTMLNode {
	return HTMLNode{html: html}
}

// F creates a document fragment
func F(children ...Node) ElementNode {
	el := ElementNode{
		object: js.Global.Get("document").Call("createDocumentFragment"),
	}
	for _, child := range children {
		child.appendTo(el)
	}
	return el
}

func GetElementByID(id string) (ElementNode, bool) {
	obj := js.Global.Get("document").Call("getElementById", id)
	if !obj.Bool() {
		return ElementNode{}, false
	}
	return ElementNode{
		object: obj,
	}, true
}
