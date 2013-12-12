package flog

import (
    "github.com/codegangsta/inject"
    //"../inject"
)

type blueprint struct {
    inject.Injector
    name string
    handlers []Handler
    routes []route
}

func NewBlueprint(name string) *blueprint {
    return &blueprint{inject.New(), name, []Handler{}, []route{} }
}
