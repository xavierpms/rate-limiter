package database

import (
	"strconv"
	"strings"
)

type TokenLimit struct {
	Name  string `json:"name"`
	Limit int    `json:"limit"`
}

type TokenLimitList struct {
	List map[string]TokenLimit
}

func NewTokenLimitList(limitsParam string) (limitList TokenLimitList) {
	limitList.List = make(map[string]TokenLimit)
	if strings.TrimSpace(limitsParam) == "" {
		return limitList
	}
	arr := strings.Split(limitsParam, ",")
	for _, v := range arr {
		value := strings.TrimSpace(v)
		if value == "" {
			continue
		}

		limite, err := strconv.Atoi(value)
		if err != nil {
			continue
		}
		token := "Token" + value
		limitList.List[token] = TokenLimit{Name: token, Limit: limite}
	}
	return limitList
}

func (tll *TokenLimitList) GetLimit(token string) int {
	limite := tll.List[token].Limit
	return limite
}

func (tll *TokenLimitList) LimitFor(token string) int {
	return tll.GetLimit(token)
}
