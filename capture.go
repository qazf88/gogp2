package gogp2

// #cgo linux pkg-config: libgphoto2
// #include <gphoto2/gphoto2.h>
// #include <string.h>
// #include <stdlib.h>
import "C"
import (
	"bytes"
	"fmt"
	"io"
	"unsafe"

	Log "github.com/qazf88/golog"
)

func (c *Camera) CapturePhoto(buffer *bytes.Buffer) error {

	type cameraFilePathInternal struct {
		Name   [128]uint8
		Folder [1024]uint8
	}

	photoPath := cameraFilePathInternal{}
	res := C.gp_camera_capture(c.Camera, 0, (*C.CameraFilePath)(unsafe.Pointer(&photoPath)), c.Context)
	if res != OK {
		if c.Camera != nil {
			c.FreeCamera()
		}
		err := fmt.Sprintf("cannot capture photo, error code: %d", res)
		Log.Error(err)
		return fmt.Errorf(err)
	}

	buff := io.Writer(buffer)
	filePath := &CameraFilePath{
		Name:     string(photoPath.Name[:bytes.IndexByte(photoPath.Name[:], 0)]),
		Folder:   string(photoPath.Folder[:bytes.IndexByte(photoPath.Folder[:], 0)]),
		Isdir:    false,
		Children: nil,
	}

	err := c.DownloadImage(buff, filePath, true)
	if err != nil {
		Log.Error(err.Error())
	}

	return nil
}

func (c *Camera) CapturePreview(buffer io.Writer) error {

	gpFile, err := newFile()
	if err != nil {
		return err
	}

	res := C.gp_camera_capture_preview(c.Camera, gpFile, c.Context)
	if res != OK {
		err := fmt.Sprintf("cannot capture preview, error code: %d", res)
		Log.Error(err)
		if gpFile != nil {
			C.gp_file_unref(gpFile)
		}
		return fmt.Errorf(err)
	}

	result := getFileBytes(gpFile, buffer)

	if gpFile != nil {
		C.gp_file_unref(gpFile)
	}

	return result
}
