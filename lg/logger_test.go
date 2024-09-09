package lg

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewZapProdLogger(t *testing.T) {
	logger := NewZapProdStdoutLogger()
	if logger == nil {
		t.Fatal("Expected a zap.SugaredLogger instance, got nil")
	}
}

func TestNewZapDevLogger(t *testing.T) {
	logger := NewZapDevLogger()
	if logger == nil {
		t.Fatal("Expected a zap.SugaredLogger instance, got nil")
	}
}

func TestNewZapFileOnlyLogger(t *testing.T) {
	tempDir := os.TempDir()
	appName := "testApp"
	logger := NewZapFileOnlyLogger(appName, tempDir)
	if logger == nil {
		t.Fatal("Expected a zap.SugaredLogger instance, got nil")
	}

	// Check if log file is created
	runID := time.Now().Format("_2006-01-02-15-04-05")
	logLocation := filepath.Join(tempDir, appName+runID+".log")

	if _, err := os.Stat(logLocation); os.IsNotExist(err) {
		t.Fatalf("Expected log file to be created, but it does not exist: %s", logLocation)
	}

	// Clean up
	_ = os.Remove(logLocation)
}

func TestInitLOG(t *testing.T) {
	InitLOG(true)
	if LOG == nil {
		t.Fatal("Expected development logger, got nil")
	}
	LOG.Debug("This is a debug message")

	InitLOG(false)
	if LOG == nil {
		t.Fatal("Expected production logger, got nil")
	}
	LOG.Info("This is an info message")
}
