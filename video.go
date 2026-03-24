// Video frame extraction helpers for the nsfw package.
// Requires ffmpeg and ffprobe to be installed and available in PATH.
package nsfwdetector

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/kinkist/nsfw-detector/logger"
)

const videoFrameCount = 5

// videoExts contains file extensions recognized as video.
var videoExts = map[string]bool{
	".mp4": true, ".webm": true, ".mov": true,
	".avi": true, ".mkv": true, ".ts": true,
	".flv": true, ".wmv": true, ".m4v": true,
}

// isVideoFile reports whether path has a recognized video extension.
func isVideoFile(path string) bool {
	return videoExts[strings.ToLower(filepath.Ext(path))]
}

// getVideoDuration returns the video duration in seconds via ffprobe. Returns 0 on failure.
func getVideoDuration(videoPath string) float64 {
	out, err := exec.Command("ffprobe",
		"-v", "quiet",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		videoPath,
	).Output()
	if err != nil {
		return 0
	}
	dur, _ := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	return dur
}

// extractVideoFrames extracts up to videoFrameCount frames from a video using ffmpeg.
// Returns the extracted frame paths and a cleanup function. Caller must call cleanup().
func extractVideoFrames(videoPath string) (frames []string, cleanup func(), err error) {
	if _, lookErr := exec.LookPath("ffmpeg"); lookErr != nil {
		return nil, nil, fmt.Errorf("ffmpeg not found (please install it): %w", lookErr)
	}

	tmpDir, err := os.MkdirTemp("", "nsfw_frames_*")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	cleanup = func() { os.RemoveAll(tmpDir) }

	dur := getVideoDuration(videoPath)
	n := videoFrameCount
	if dur > 0 && dur < float64(n) {
		n = max(1, int(dur))
	}
	baseDur := dur
	if baseDur <= 0 {
		baseDur = 1
	}
	fpsExpr := fmt.Sprintf("%.6f", float64(n)/baseDur)
	outPattern := filepath.Join(tmpDir, "frame%03d.jpg")

	logger.Debug("extracting video frames: %s (duration=%.1fs, n=%d, fps=%s)",
		filepath.Base(videoPath), dur, n, fpsExpr)

	out, runErr := exec.Command("ffmpeg",
		"-i", videoPath,
		"-vf", fmt.Sprintf("fps=%s", fpsExpr),
		"-frames:v", strconv.Itoa(n),
		"-y", outPattern,
	).CombinedOutput()
	if runErr != nil {
		cleanup()
		return nil, nil, fmt.Errorf("ffmpeg frame extraction failed: %w\n%s", runErr, string(out))
	}

	entries, readErr := os.ReadDir(tmpDir)
	if readErr != nil || len(entries) == 0 {
		cleanup()
		return nil, nil, fmt.Errorf("no frames extracted from %q", videoPath)
	}
	for _, e := range entries {
		if !e.IsDir() {
			frames = append(frames, filepath.Join(tmpDir, e.Name()))
		}
	}
	logger.Debug("frame extraction complete: %d frames", len(frames))
	return frames, cleanup, nil
}
