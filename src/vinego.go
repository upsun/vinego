package main

import (
	"os"

	"golang.org/x/tools/go/analysis"
	"gopkg.in/yaml.v2"

	"github.com/upsun/vinego/src/allfields"
	"github.com/upsun/vinego/src/capturederr"
	"github.com/upsun/vinego/src/explicitcast"
	"github.com/upsun/vinego/src/loopvariableref"
	"github.com/upsun/vinego/src/varinit"
)

type analyzerPlugin struct{}

func (*analyzerPlugin) GetAnalyzers() []*analysis.Analyzer {
	// check:allfields
	type Config struct {
		EnableVarinit         bool `yaml:"enable_varinit"`
		EnableExplicitcast    bool `yaml:"enable_explicitcast"`
		EnableLoopVariableRef bool `yaml:"enable_loopvariableref"`
		EnableCapturedErr     bool `yaml:"enable_capturederr"`
	}
	var config = Config{
		EnableVarinit:         false,
		EnableExplicitcast:    false,
		EnableLoopVariableRef: false,
		EnableCapturedErr:     false,
	}
	confBytes, err := os.ReadFile(".vinego.yaml")
	if err != nil {
		if os.IsNotExist(err) {
			goto PostConfig
		} else {
			panic("Error reading vinego linters config: " + err.Error())
		}
	}
	err = yaml.UnmarshalStrict(confBytes, &config)
	if err != nil {
		panic("Error parsing vinego linters config: " + err.Error())
	}
PostConfig:

	// Note via `lintersdb/manager.go`
	// all custom linters are hard-coded to have "LoadModeTypesInfo" which means we have
	// ast and type check info
	out := []*analysis.Analyzer{}
	out = append(out, allfields.New())
	if config.EnableVarinit {
		out = append(out, varinit.New())
	}
	if config.EnableExplicitcast {
		out = append(out, explicitcast.New())
	}
	if config.EnableLoopVariableRef {
		out = append(out, loopvariableref.New())
	}
	if config.EnableCapturedErr {
		out = append(out, capturederr.New())
	}
	return out
}

var AnalyzerPlugin analyzerPlugin = analyzerPlugin{}
