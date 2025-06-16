package main

import (
	"context"
	"path"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context
}

// startup is called when the app starts. The context is saved so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// shutdown is called when the app is closing. We can use this to save data
func (a *App) shutdown(ctx context.Context) {

}

// Get port number of the media server
func (a *App) GetMediaServerPort() int {
	return ServerPort
}

// Open a video
func (a *App) OpenVideo(videoPath string) (map[string]string, error) {
	var err error
	LastOpenedVideo, err = runtime.OpenFileDialog(MainApp.ctx, runtime.OpenDialogOptions{
		Title:                "Select a video file",
		CanCreateDirectories: false,
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Video Files",
				Pattern:     "*.mp4;*.avi;*.mkv;*.mov;*.flv;*.wmv;*.webm",
			},
		},
	})
	if err != nil {
		return nil, err
	}
	return map[string]string{
		"name":     path.Base(LastOpenedVideo),
		"location": path.Dir(LastOpenedVideo),
		"fullPath": LastOpenedVideo,
	}, nil
}

type Keyframe struct {
	Time float64 `json:"time"` // Time in seconds
	X    float64 `json:"x"`    // X coordinate
	Y    float64 `json:"y"`    // Y coordinate
	W    float64 `json:"w"`    // Width of the crop area
	H    float64 `json:"h"`    // Height of the crop area
}

// Save a video
func (a *App) SaveVideo(videoPath string, keyframes []Keyframe) error {
	err := ConvertVideoRatio(videoPath, keyframes)
	if err != nil {
		runtime.LogError(a.ctx, "Failed to convert video ratio: "+err.Error())
		a.Quit()
	}
	return err
}

// Check if ffmpeg and ffprobe are available on the system
func (a *App) HasFFmpeg() bool {
	return FFmpegExists()
}

// Quit the application
func (a *App) Quit() {
	runtime.Quit(a.ctx)
}
