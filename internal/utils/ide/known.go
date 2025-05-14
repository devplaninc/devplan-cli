package ide

import (
	"github.com/pkg/errors"
	"golang.org/x/exp/slices"
)

type IDE string

const (
	PyCharm  IDE = "pycharm"
	IntelliJ IDE = "intellij"
	GoLand   IDE = "goland"
	WebStorm IDE = "webstorm"
	PhpStorm IDE = "phpstorm"
	RubyMine IDE = "rubymine"
	CLion    IDE = "clion"
	Rider    IDE = "rider"
	DataGrip IDE = "datagrip"
	AppCode  IDE = "appcode"
	Windsurf IDE = "windsurf"
	Cursor   IDE = "cursor"
)

type Assistant string

const (
	CursorAI   Assistant = "cursor"
	JunieAI    Assistant = "junie"
	WindsurfAI Assistant = "windsurf"
)

var pathMap = map[Assistant]string{
	CursorAI:   ".cursor",
	JunieAI:    ".junie",
	WindsurfAI: ".windsurf",
}

func GetJetbrainsIDEs() []IDE {
	return []IDE{
		PyCharm,
		IntelliJ,
		GoLand,
		WebStorm,
		PhpStorm,
		RubyMine,
		CLion,
		Rider,
		DataGrip,
		AppCode,
	}
}

func GetKnown() []IDE {
	return append(GetJetbrainsIDEs(), Cursor, Windsurf)
}

func GetAssistant(ide IDE) (Assistant, error) {
	if slices.Contains(GetJetbrainsIDEs(), ide) {
		return JunieAI, nil
	}
	switch ide {
	case Cursor:
		return CursorAI, nil
	case Windsurf:
		return WindsurfAI, nil
	}
	return "", errors.Errorf("unexpected IDE: %s", ide)
}

func GetAssistants() []Assistant {
	return []Assistant{CursorAI, JunieAI, WindsurfAI}
}
