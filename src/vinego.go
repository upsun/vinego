package vinego

import (
	"golang.org/x/tools/go/analysis"

	"github.com/golangci/plugin-module-register/register"
	"github.com/upsun/vinego/src/allfields"
	"github.com/upsun/vinego/src/capturederr"
	"github.com/upsun/vinego/src/explicitcast"
	"github.com/upsun/vinego/src/varinit"
)

type Settings struct {
	EnableVarinit      bool `json:"enable_varinit"`
	EnableExplicitcast bool `json:"enable_explicitcast"`
	EnableCapturedErr  bool `json:"enable_capturederr"`
}

type Vinego struct {
	settings Settings
}

func (f *Vinego) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	out := []*analysis.Analyzer{}
	out = append(out, allfields.New())
	if f.settings.EnableVarinit {
		out = append(out, varinit.New())
	}
	if f.settings.EnableExplicitcast {
		out = append(out, explicitcast.New())
	}
	if f.settings.EnableCapturedErr {
		out = append(out, capturederr.New())
	}
	return out, nil
}

func (f *Vinego) GetLoadMode() string {
	return register.LoadModeTypesInfo
}

func New(settings any) (register.LinterPlugin, error) {
	s, err := register.DecodeSettings[Settings](settings)
	if err != nil {
		return nil, err
	}
	return &Vinego{settings: s}, nil
}

func init() {
	register.Plugin("vinego", New)
}
