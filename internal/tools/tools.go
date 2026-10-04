package tools

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

var ExecCommand = Exec

func Exec(ctx context.Context, command string, args []string) (string, error) {
	cmd := exec.CommandContext(ctx, command, args...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func CheckCommand(name string) error {
	_, err := exec.LookPath(name)
	if err != nil {
		return fmt.Errorf("команда %s не найдена: %w", name, err)
	}
	return nil
}

func ParseFloat(input string) float64 {
	buff := strings.ReplaceAll(input, ",", ".")
	output, err := strconv.ParseFloat(buff, 64)
	if err != nil {
		return 0.0
	}
	return output
}

func ParseUint64(str string) uint64 {
	val, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return 0
	}
	return val
}

func ParseUint32(str string) uint32 {
	val, err := strconv.ParseUint(str, 10, 32)
	if err != nil {
		return 0
	}
	if val >= uint64(^uint32(0)) {
		return 0
	}
	return uint32(val)
}
