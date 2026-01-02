package ide

import (
	"github.com/pkg/errors"
	"golang.org/x/exp/slices"
)

type IDE string

const (
	PyCharm   IDE = "pycharm"
	IntelliJ  IDE = "intellij"
	GoLand    IDE = "goland"
	WebStorm  IDE = "webstorm"
	PhpStorm  IDE = "phpstorm"
	RubyMine  IDE = "rubymine"
	CLion     IDE = "clion"
	Rider     IDE = "rider"
	DataGrip  IDE = "datagrip"
	AppCode   IDE = "appcode"
	Windsurf  IDE = "windsurf"
	Cursor    IDE = "cursor"
	CursorCLI IDE = "cursor-cli"
	Claude    IDE = "claude"
)

type Assistant string

const (
	CursorAI   Assistant = "cursor"
	JunieAI    Assistant = "junie"
	WindsurfAI Assistant = "windsurf"
	ClaudeAI   Assistant = "claude"
)

var pathMap = map[Assistant]string{
	CursorAI:   ".cursor",
	JunieAI:    ".junie",
	WindsurfAI: ".windsurf",
	ClaudeAI:   ".",
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
	return append(GetJetbrainsIDEs(), CursorCLI, Cursor, Windsurf, Claude)
}

func GetAssistant(ide IDE) (Assistant, error) {
	if slices.Contains(GetJetbrainsIDEs(), ide) {
		return JunieAI, nil
	}
	switch ide {
	case Cursor:
		return CursorAI, nil
	case CursorCLI:
		return CursorAI, nil
	case Windsurf:
		return WindsurfAI, nil
	case Claude:
		return ClaudeAI, nil
	}
	return "", errors.Errorf("unexpected IDE: %s", ide)
}

func GetAssistants() []Assistant {
	return []Assistant{ClaudeAI, CursorAI, JunieAI, WindsurfAI}
}

func (i IDE) String() string {
	return i.DisplayName()
}

func (i IDE) DisplayName() string {
	switch i {
	case Claude:
		return "Claude Code"
	case Cursor:
		return "Cursor"
	case WebStorm:
		return "WebStorm"
	case PyCharm:
		return "PyCharm"
	case IntelliJ:
		return "IntelliJ IDEA"
	case GoLand:
		return "GoLand"
	case PhpStorm:
		return "PhpStorm"
	case RubyMine:
		return "RubyMine"
	case CLion:
		return "CLion"
	case Rider:
		return "Rider"
	case DataGrip:
		return "DataGrip"
	case AppCode:
		return "AppCode"
	case Windsurf:
		return "Windsurf"
	case CursorCLI:
		return "Cursor CLI"
	default:
		return string(i)
	}
}
