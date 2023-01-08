package main

import "github.com/beranek1/goadsinterface"

type AdsSource struct {
	lib goadsinterface.AdsLibrary
}

func (s AdsSource) Get(key string) (any, error) {
	d, err := s.lib.GetSymbolValue(key)
	if err != nil {
		return nil, err
	}
	return d.Data, nil
}

func (s AdsSource) List() ([]string, error) {
	return s.lib.GetSymbolList()
}
