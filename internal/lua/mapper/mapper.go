package mapper

import (
	"errors"
	"fmt"
	"sync"

	lua "github.com/yuin/gopher-lua"
)

type LuaMapperCollection struct {
	mappers map[string]*LuaMapper

	mu sync.Mutex
}

func NewLuaMapperCollection() *LuaMapperCollection {
	return &LuaMapperCollection{
		mappers: make(map[string]*LuaMapper),
	}
}

func (c *LuaMapperCollection) Add(list []*LuaMapper) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var errs []error
	for _, m := range list {
		if _, ok := c.mappers[m.name]; ok {
			errs = append(errs, fmt.Errorf("mapper with name '%s' already exists", m.name))

			continue
		}

		c.mappers[m.name] = m
	}

	return errors.Join(errs...)
}

func (c *LuaMapperCollection) InitMappers(L *lua.LState) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, m := range c.mappers {
		err := m.Init(L)
		if err != nil {
			return fmt.Errorf("cannot init mapper '%s': %w", m.name, err)
		}
	}

	return nil
}

func (c *LuaMapperCollection) CloseMappers(L *lua.LState) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, m := range c.mappers {
		err := m.Close(L)
		if err != nil {
			return fmt.Errorf("cannot close mapper '%s': %w", m.name, err)
		}
	}

	return nil
}

type LuaMapper struct {
	name string
	fn   func(L *lua.LState) (*lua.LTable, error)
}

func NewLuaMapper(name string, fn func(L *lua.LState) (*lua.LTable, error)) *LuaMapper {
	return &LuaMapper{
		name: name,
		fn:   fn,
	}
}

func (m *LuaMapper) Init(L *lua.LState) error {
	tbl, err := m.fn(L)
	if err != nil {
		return fmt.Errorf("lua mapper '%s' cannot be applied: %w", m.name, err)
	}

	L.SetGlobal(m.name, tbl)

	return nil
}

func (m *LuaMapper) Close(L *lua.LState) error {
	L.SetGlobal(m.name, lua.LNil)

	return nil
}
