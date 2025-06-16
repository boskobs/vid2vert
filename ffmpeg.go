package main

import (
	"bufio"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"reliveman/helpers"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// buildCropExpr builds a nested FFmpeg expression for a given property ("X", "Y", "W", "H") from keyframes.
func buildCropExpr(keyframes []Keyframe, prop string) string {
	if len(keyframes) == 0 {
		return ""
	}
	var expr strings.Builder
	for i := range len(keyframes) - 1 {
		kf0 := keyframes[i]
		kf1 := keyframes[i+1]
		dt := kf1.Time - kf0.Time
		var v0, v1 float64
		switch prop {
		case "X":
			v0, v1 = kf0.X, kf1.X
		case "Y":
			v0, v1 = kf0.Y, kf1.Y
		case "W":
			v0, v1 = kf0.W, kf1.W
		case "H":
			v0, v1 = kf0.H, kf1.H
		default:
			return ""
		}
		dv := v1 - v0
		slope := dv / dt
		intercept := v0 - slope*kf0.Time
		fmt.Fprintf(&expr, "if(lte(t,%.3f),%.6f*t+%.6f,", kf1.Time, slope, intercept)
	}
	// Last value
	var last float64
	switch prop {
	case "X":
		last = keyframes[len(keyframes)-1].X
	case "Y":
		last = keyframes[len(keyframes)-1].Y
	case "W":
		last = keyframes[len(keyframes)-1].W
	case "H":
		last = keyframes[len(keyframes)-1].H
	}
	fmt.Fprintf(&expr, "%.6f", last)
	expr.WriteString(strings.Repeat(")", len(keyframes)-1))
	return expr.String()
}

func ConvertVideoRatio(videoPath string, keyframes []Keyframe) error {
	if len(keyframes) < 1 {
		return fmt.Errorf("no keyframes provided")
	}
	probeCmd := exec.Command("ffprobe", "-v", "error", "-select_streams", "v:0", "-show_entries", "stream=width,height", "-of", "csv=p=0", videoPath)
	probeOut, err := probeCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get video dimensions: %w", err)
	}
	var width, height float64
	_, err = fmt.Sscanf(strings.TrimSpace(string(probeOut)), "%f,%f", &width, &height)
	if err != nil {
		return fmt.Errorf("failed to parse video dimensions: %w", err)
	}
	for i := range keyframes {
		keyframes[i].X = keyframes[i].X * width / 100.0
		keyframes[i].Y = keyframes[i].Y * height / 100.0
		keyframes[i].W = keyframes[i].W * width / 100.0
		keyframes[i].H = keyframes[i].H * height / 100.0
	}
	keyframes[0].Time = 0

	cropW := buildCropExpr(keyframes, "W")
	cropH := buildCropExpr(keyframes, "H")
	cropX := buildCropExpr(keyframes, "X")
	cropY := buildCropExpr(keyframes, "Y")
	cropFilter := fmt.Sprintf("crop=%s:%s:%s:%s", cropW, cropH, cropX, cropY)

	// Escape commas for ffmpeg filter_complex usage
	cropFilterEscaped := strings.ReplaceAll(cropFilter, ",", "\\,")

	videoDir := filepath.Dir(videoPath)

	outputPath := filepath.Join(videoDir, "cropped_"+filepath.Base(videoPath))

	ffmpegCmd := helpers.Command(
		"ffmpeg",
		"-i", videoPath,
		"-vf", cropFilterEscaped,
		"-c:a", "copy", "-y",
		outputPath,
		"-progress", "pipe:1",
	)

	stdout, err := ffmpegCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get ffmpeg stdout: %w", err)
	}
	stderr, err := ffmpegCmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get ffmpeg stderr: %w", err)
	}

	if err := ffmpegCmd.Start(); err != nil {
		return fmt.Errorf("ffmpeg failed to start: %w", err)
	}

	// Regex to extract out_time or out_time_ms
	outTimeRe := regexp.MustCompile(`out_time(?:_ms)?=(.*)`)

	// Get video duration (in seconds)
	probeDurCmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", videoPath)
	durOut, err := probeDurCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get video duration: %w", err)
	}
	var duration float64
	_, err = fmt.Sscanf(strings.TrimSpace(string(durOut)), "%f", &duration)
	if err != nil {
		return fmt.Errorf("failed to parse video duration: %w", err)
	}

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			if outTimeRe.MatchString(line) {
				val := outTimeRe.FindStringSubmatch(line)[1]
				var seconds float64
				if strings.Contains(line, "out_time_ms") {
					fmt.Sscanf(val, "%f", &seconds)
					seconds = seconds / 1000000.0
				} else {
					// out_time is in HH:MM:SS.microseconds
					parts := strings.SplitN(val, ".", 2)
					timeParts := strings.Split(parts[0], ":")
					if len(timeParts) == 3 {
						hh := 0
						mm := 0
						ss := 0
						fmt.Sscanf(timeParts[0], "%d", &hh)
						fmt.Sscanf(timeParts[1], "%d", &mm)
						fmt.Sscanf(timeParts[2], "%d", &ss)
						seconds = float64(hh*3600 + mm*60 + ss)
						if len(parts) == 2 {
							frac := 0
							fmt.Sscanf(parts[1], "%d", &frac)
							seconds += float64(frac) / 1000000.0
						}
					}
				}
				percent := 0.0
				if duration > 0 {
					percent = (seconds / duration) * 100.0
				}
				if MainApp.ctx != nil {
					runtime.EventsEmit(MainApp.ctx, "app:progress", percent)
				}
				fmt.Printf("Progress: %.2f%%\n", percent)
			}
		}
	}()
	// Print errors from stderr
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()

	if err := ffmpegCmd.Wait(); err != nil {
		return fmt.Errorf("ffmpeg failed: %w", err)
	}

	return nil
}

// Check if ffmpeg and ffprobe are available on the system
func FFmpegExists() bool {
	_, err1 := exec.LookPath("ffmpeg")
	_, err2 := exec.LookPath("ffprobe")
	return err1 == nil || err2 == nil
}
