// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

type Foo struct {
	ID   string  `json:"id"`
	Name *string `json:"name,omitempty"`
}

func (Foo) IsEntity() {}

type Query struct {
}
