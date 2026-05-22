package permission

import (
	"context"
	"errors"
	"fmt"
	"regexp"
)

type Action int

const (
	ActionAllow Action = iota
	ActionDeny
	ActionAsk
)

type Approver interface {
	RequestApproval(ctx context.Context, toolName string, args map[string]any) (bool, error)
	ConfirmExternal(ctx context.Context, toolName string, args map[string]any) (bool, error)
}

type Rule struct {
	Tool   string `yaml:"tool"`
	Match  string `yaml:"match"`
	Action Action `yaml:"action"`
	re     *regexp.Regexp
}

type Checker struct {
	global  map[string]Action
	patterns []Rule
}

var ErrDenied = errors.New("permission denied")

func New() *Checker {
	return &Checker{
		global:   make(map[string]Action),
		patterns: nil,
	}
}

func (c *Checker) SetDefault(tool string, action Action) {
	c.global[tool] = action
}

func (c *Checker) AddRule(rule Rule) {
	if rule.Match != "" {
		rule.re = regexp.MustCompile(rule.Match)
	}
	c.patterns = append(c.patterns, rule)
}

func (c *Checker) Evaluate(toolName string, args map[string]any) (Action, error) {
	for _, rule := range c.patterns {
		if rule.Tool != toolName {
			continue
		}
		if rule.re != nil {
			for _, v := range args {
				if s, ok := v.(string); ok {
					if rule.re.MatchString(s) {
						return rule.Action, nil
					}
				}
			}
		}
	}

	if action, ok := c.global[toolName]; ok {
		return action, nil
	}

	return ActionAllow, nil
}

func (a Action) String() string {
	switch a {
	case ActionAllow:
		return "allow"
	case ActionDeny:
		return "deny"
	case ActionAsk:
		return "ask"
	default:
		return fmt.Sprintf("unknown(%d)", a)
	}
}
