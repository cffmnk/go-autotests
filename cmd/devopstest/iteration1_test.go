package main

// Basic imports
import (
	"context"
	"errors"
	"os"
	"syscall"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/cffmnk/go-autotests/internal/fork"
)

// Iteration1Suite is a suite of autotests
type Iteration1Suite struct {
	suite.Suite

	agentAddress string
	agentProcess *fork.BackgroundProcess
}

// SetupSuite bootstraps suite dependencies
func (suite *Iteration1Suite) SetupSuite() {
	// check required flags
	suite.Require().NotEmpty(flagAgentBinaryPath, "-agent-binary-path non-empty flag required")

	suite.agentAddress = "http://localhost:8080"

	envs := append(os.Environ(), []string{
		"RESTORE=false",
	}...)
	p := fork.NewBackgroundProcess(context.Background(), flagAgentBinaryPath,
		fork.WithEnv(envs...),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	err := p.Start(ctx)
	if err != nil {
		suite.T().Errorf("Невозможно запустить процесс командой %s: %s. Переменные окружения: %+v", p, err, envs)
		return
	}

	port := "8080"
	err = p.ListenPort(ctx, "tcp", port)
	if err != nil {
		suite.T().Errorf("Не удалось дождаться пока на порт %s начнут поступать данные: %s", port, err)
		if out := p.Stderr(ctx); len(out) > 0 {
			suite.T().Logf("Получен STDERR лог агента:\n\n%s\n\n", string(out))
		}
		if out := p.Stdout(ctx); len(out) > 0 {
			suite.T().Logf("Получен STDOUT лог агента:\n\n%s\n\n", string(out))
		}
		return
	}

	suite.agentProcess = p
}

// TearDownSuite teardowns suite dependencies
func (suite *Iteration1Suite) TearDownSuite() {
	if suite.agentProcess == nil {
		return
	}

	exitCode, err := suite.agentProcess.Stop(syscall.SIGINT, syscall.SIGKILL)
	if err != nil {
		if errors.Is(err, os.ErrProcessDone) {
			return
		}
		suite.T().Logf("Не удалось остановить процесс с помощью сигнала ОС: %s", err)
		return
	}

	if exitCode > 0 {
		suite.T().Logf("Процесс завершился с не нулевым статусом %d", exitCode)
	}

	// try to read stdout/stderr
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	if out := suite.agentProcess.Stderr(ctx); len(out) > 0 {
		suite.T().Logf("Получен STDERR лог агента:\n\n%s\n\n", string(out))
	}
	if out := suite.agentProcess.Stdout(ctx); len(out) > 0 {
		suite.T().Logf("Получен STDOUT лог агента:\n\n%s\n\n", string(out))
	}
}

// TestAgent проверяет
// агент успешно стартует и передает какие-то данные по tcp, на 127.0.0.1:8080
func (suite *Iteration1Suite) TestAgent() {
	suite.Run("receive data from agent", func() {
	})
}
