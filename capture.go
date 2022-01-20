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
	"time"
	"unsafe"

	Log "github.com/qazf88/golog"
)

// CaptureExternalEvent
func (c *Camera) CaptureExternalEvent(timeout int, bufferOut io.Writer) error {

	file, err := newFile()
	if err != nil {
		Log.Error(err.Error())
		return err
	}

	var eventType C.CameraEventType
	var vp unsafe.Pointer

	defer C.free(vp)
	defer C.gp_file_free(file)

	timer1 := time.NewTimer(time.Duration(timeout) * time.Second)

	for {
		select {
		case <-timer1.C:
			return fmt.Errorf("timeout")
		default:
			res := C.gp_camera_wait_for_event(c.Camera, C.int(timeout+5), &eventType, &vp, c.Context)
			if res != OK {
				err := fmt.Sprintf("error wait for event, error code: %d", res)
				Log.Error(err)
				return fmt.Errorf(err)
			}

			if int(eventType) != EVENT_FILE_ADDED {
				continue
			}
			cameraFilePath := (*C.CameraFilePath)(vp)
			defer C.free(unsafe.Pointer(cameraFilePath))

			res = C.gp_camera_file_get(c.Camera, (*C.char)(&cameraFilePath.folder[0]), (*C.char)(&cameraFilePath.name[0]), FileTypeNormal, file, c.Context)
			if res != OK {
				err := fmt.Sprintf("error get file from camera, error code: %d", res)
				Log.Error(err)
				return fmt.Errorf(err)
			}

			err := getFileBytes(file, bufferOut)
			if err != nil {
				C.gp_camera_file_delete(c.Camera, (*C.char)(&cameraFilePath.folder[0]), (*C.char)(&cameraFilePath.name[0]), c.Context)
				Log.Error(err.Error())
				return err
			}

			return nil
		}
	}

}

// CapturePhoto
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

// CaptureCompletedEvent
func (c *Camera) CaptureCompletedEvent(bufferOut io.Writer) error {

	file, err := newFile()
	if err != nil {
		Log.Error(err.Error())
		return err
	}

	var eventType C.CameraEventType
	var vp unsafe.Pointer

	defer C.free(vp)
	defer C.gp_file_free(file)

	timer1 := time.NewTimer(time.Duration(2) * time.Second)

	for {
		select {
		case <-timer1.C:
			return fmt.Errorf("timeout")
		default:
			res := C.gp_camera_wait_for_event(c.Camera, C.int(5), &eventType, &vp, c.Context)
			if res != OK {
				err := fmt.Sprintf("error wait for event, error code: %d", res)
				Log.Error(err)
				return fmt.Errorf(err)
			}

			if int(eventType) != EVENT_FILE_ADDED {
				continue
			}

			cameraFilePath := (*C.CameraFilePath)(vp)
			defer C.free(unsafe.Pointer(cameraFilePath))

			res = C.gp_camera_file_get(c.Camera, (*C.char)(&cameraFilePath.folder[0]), (*C.char)(&cameraFilePath.name[0]), FileTypeNormal, file, c.Context)
			if res != OK {
				err := fmt.Sprintf("error get file from camera, error code: %d", res)
				Log.Error(err)
				return fmt.Errorf(err)
			}

			err := getFileBytes(file, bufferOut)
			if err != nil {
				C.gp_camera_file_delete(c.Camera, (*C.char)(&cameraFilePath.folder[0]), (*C.char)(&cameraFilePath.name[0]), c.Context)
				Log.Error(err.Error())
				return err
			}

			return nil
		}
	}

}

// ClearIramFile
func (c *Camera) ClearIramFile() {

	file, err := newFile()
	if err != nil {
		Log.Error(err.Error())
		return
	}

	var vp unsafe.Pointer
	var eventType C.CameraEventType

	defer C.free(vp)
	defer C.gp_file_free(file)

	timer1 := time.NewTimer(time.Duration(10) * time.Second)

loop:
	for {
		select {

		case <-timer1.C:
			break loop
		default:
			res := C.gp_camera_wait_for_event(c.Camera, C.int(6), &eventType, &vp, c.Context)
			if res != OK {
				break loop
			}

			if int(eventType) == EVENT_TIMEOUT {
				break loop
			}

			if int(eventType) != EVENT_FILE_ADDED {
				continue
			}

			cameraFilePath := (*C.CameraFilePath)(vp)
			defer C.free(unsafe.Pointer(cameraFilePath))
			res = C.gp_camera_file_delete(c.Camera, (*C.char)(&cameraFilePath.folder[0]), (*C.char)(&cameraFilePath.name[0]), c.Context)
			if res != OK {
				break loop
			}

		}
	}
}
