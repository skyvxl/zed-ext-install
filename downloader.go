package main

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"time"
)

const (
	maxRetries    = 5
	baseBackoff   = 2 * time.Second
	maxBackoffSec = 30
)

// DownloadFile downloads a URL to a local file with retry and progress output.
func DownloadFile(url, destPath string) error {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			delay := time.Duration(math.Min(
				float64(baseBackoff)*math.Pow(2, float64(attempt-1)),
				float64(maxBackoffSec)*float64(time.Second),
			))
			fmt.Printf("  retry %d/%d in %s...\n", attempt, maxRetries, delay)
			time.Sleep(delay)
		}

		lastErr = downloadOnce(url, destPath)
		if lastErr == nil {
			return nil
		}
		fmt.Printf("  attempt failed: %v\n", lastErr)
	}

	return fmt.Errorf("download failed after %d retries: %w", maxRetries, lastErr)
}

func downloadOnce(url, destPath string) error {
	resp, err := httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer out.Close()

	totalSize := resp.ContentLength
	var reader io.Reader = resp.Body

	if totalSize > 0 {
		reader = &progressReader{
			reader: resp.Body,
			total:  totalSize,
		}
	}

	written, err := io.Copy(out, reader)
	if err != nil {
		os.Remove(destPath)
		return fmt.Errorf("write error: %w", err)
	}

	if totalSize > 0 {
		fmt.Println() // newline after progress
	}

	if totalSize > 0 && written != totalSize {
		os.Remove(destPath)
		return fmt.Errorf("incomplete download: got %d of %d bytes", written, totalSize)
	}

	return nil
}

type progressReader struct {
	reader  io.Reader
	total   int64
	current int64
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	pr.current += int64(n)
	pct := float64(pr.current) / float64(pr.total) * 100
	fmt.Printf("\r  downloading... %.1f%% (%s / %s)",
		pct, formatBytes(pr.current), formatBytes(pr.total))
	return n, err
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
