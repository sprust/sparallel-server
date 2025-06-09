package handlers

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"sparallel_server/pkg/foundation/errs"
	"strings"
	"sync"
	"time"
)

type FileHandler struct {
	logFile            *os.File
	errorLogFile       *os.File
	logDirPath         string
	logKeepDays        int
	initFileMutex      sync.Mutex
	currentLogFileName string
}

func (h *FileHandler) Handle(ctx context.Context, r slog.Record) error {
	err := h.freshFileHandler()
	if err != nil {
		return err
	}

	msg := makeMessageByRecord(r) + "\n"

	// Записываем в основной лог
	_, err = h.logFile.WriteString(msg)
	if err != nil {
		return errs.Err(err)
	}

	// Если это ошибка, записываем в отдельный файл
	if r.Level >= slog.LevelError {
		err = h.writeError(msg)
		if err != nil {
			return errs.Err(err)
		}
	}

	return nil
}

func (h *FileHandler) writeError(msg string) error {
	errorFileName := h.makeErrorLogFileName(time.Now())
	filePath := filepath.Join(h.logDirPath, errorFileName)

	// Создаем или открываем файл для ошибок
	if h.errorLogFile == nil {
		var err error
		h.errorLogFile, err = os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			return errs.Err(err)
		}
	}

	_, err := h.errorLogFile.WriteString(msg)
	return errs.Err(err)
}

func (h *FileHandler) freshFileHandler() error {
	actualLogFileName := h.makeLogFileName(time.Now())

	if actualLogFileName == h.currentLogFileName {
		return nil
	}

	h.initFileMutex.Lock()
	defer h.initFileMutex.Unlock()

	if actualLogFileName == h.currentLogFileName {
		return nil
	}

	filePath := filepath.Join(h.logDirPath, actualLogFileName)
	dir := filepath.Dir(filePath)

	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return errs.Err(err)
	}

	if h.logFile != nil {
		_ = h.logFile.Close()
	}
	if h.errorLogFile != nil {
		_ = h.errorLogFile.Close()
		h.errorLogFile = nil
	}

	logFile, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		slog.Error("Failed to open log file: " + err.Error())
		return errs.Err(err)
	}

	h.logFile = logFile
	h.currentLogFileName = actualLogFileName

	return nil
}

func (h *FileHandler) Close() error {
	slog.Warn("Closing log files")

	h.initFileMutex.Lock()
	defer h.initFileMutex.Unlock()

	var err error
	if h.logFile != nil {
		err = h.logFile.Close()
	}
	if h.errorLogFile != nil {
		if errClose := h.errorLogFile.Close(); errClose != nil && err == nil {
			err = errClose
		}
	}

	return errs.Err(err)
}

func NewFileHandler(logDirPath string, logKeepDays int) (*FileHandler, error) {
	h := &FileHandler{
		logDirPath:  strings.Trim(logDirPath, "/"),
		logKeepDays: logKeepDays,
	}

	err := h.freshFileHandler()

	if err != nil {
		return nil, err
	}

	h.deleteOldFiles()

	return h, nil
}

func (h *FileHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func (h *FileHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *FileHandler) WithGroup(name string) slog.Handler {
	return h
}

func (h *FileHandler) deleteOldFiles() {
	go func() {
		for {
			var currentLogFilePaths []string

			filesInLogDir, err := os.ReadDir(h.logDirPath)

			if err != nil {
				slog.Error("Failed to read log directory: " + err.Error())
				time.Sleep(1 * time.Minute)
				return
			}

			for _, file := range filesInLogDir {
				if !file.IsDir() && (strings.HasSuffix(file.Name(), ".log") || strings.HasSuffix(file.Name(), "-ERROR.log")) {
					currentLogFilePaths = append(currentLogFilePaths, filepath.Join(h.logDirPath, file.Name()))
				}
			}

			keepDays := h.logKeepDays + 1

			var expectedFilePaths []string

			for i := 0; i < keepDays; i++ {
				date := time.Now().AddDate(0, 0, -i)
				expectedFilePaths = append(
					expectedFilePaths,
					filepath.Join(h.logDirPath, h.makeLogFileName(date)),
					filepath.Join(h.logDirPath, h.makeErrorLogFileName(date)),
				)
			}

			for _, currentLogFilePath := range currentLogFilePaths {
				if slices.Index(expectedFilePaths, currentLogFilePath) != -1 {
					continue
				}

				err = os.Remove(currentLogFilePath)

				if err != nil {
					slog.Error("Failed to remove old log file: " + currentLogFilePath + ": " + err.Error())
				} else {
					slog.Warn("Removed old log file: " + currentLogFilePath)
				}
			}

			time.Sleep(1 * time.Hour)
		}
	}()
}

func (h *FileHandler) makeLogFileName(time time.Time) string {
	return time.Format("2006-01-02") + ".log"
}

func (h *FileHandler) makeErrorLogFileName(time time.Time) string {
	return time.Format("2006-01-02") + "-ERROR.log"
}
