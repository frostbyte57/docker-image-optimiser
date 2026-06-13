package rewrite

import (
	"strconv"

	"github.com/yuxiangchang/docker-image-optimiser/internal/rules"
)

func summary(f rules.Finding) string {
	return "line " + strconv.Itoa(f.Line) + " [" + f.Rule + "] " + f.Message
}
