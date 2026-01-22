package api

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"http-win-notice/frp"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func GetFrpStatusApi(c *gin.Context) {
	status := frp.GetStatus()
	c.JSON(http.StatusOK, status)
}

func GetFrpLogsApi(c *gin.Context) {
	lines := parseLines(c.DefaultQuery("lines", "200"))
	status := frp.GetStatus()
	logs, err := tailFile(status.LogPath, lines)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"logs":     []string{},
			"message":  err.Error(),
			"log_path": status.LogPath,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"logs":     logs,
		"log_path": status.LogPath,
	})
}

func parseLines(raw string) int {
	lines, err := strconv.Atoi(raw)
	if err != nil || lines <= 0 {
		return 200
	}
	if lines > 1000 {
		return 1000
	}
	return lines
}

func tailFile(logPath string, maxLines int) ([]string, error) {
	file, err := os.Open(logPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	const maxBytes = int64(512 * 1024)
	start := int64(0)
	if info.Size() > maxBytes {
		start = info.Size() - maxBytes
	}
	if start > 0 {
		_, err = file.Seek(start, io.SeekStart)
		if err != nil {
			return nil, err
		}
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	if start > 0 {
		if idx := bytes.IndexByte(data, '\n'); idx != -1 {
			data = data[idx+1:]
		}
	}

	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	lines := strings.Split(text, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) > maxLines {
		lines = lines[len(lines)-maxLines:]
	}
	return lines, nil
}
