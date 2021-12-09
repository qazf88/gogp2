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
	"strconv"
	"unsafe"

	Log "github.com/qazf88/golog"
)

//Init camera
func (c *Camera) Init() bool {
	if c.Context == nil {
		if c.NewContext() != nil {
			c.CameraStatus = false
			return false
		}
	}
	if c.Camera == nil {
		if c.NewCamera() != nil {
			c.CameraStatus = false
			return false
		}
	}
	if !c.CameraStatus {
		if c.InitCamera() != nil {
			return false
		}
	}

	c.CameraStatus = true
	Log.Info("New camera")
	return true
}

//Get new camera
func (c *Camera) NewCamera() error {
	Log.Trace("get new camera")
	if c.Context == nil {
		err := "could not get camera, context is empty"
		Log.Error(err)
		return fmt.Errorf(err)
	}
	if c.Camera != nil {
		err := "Camera is already initialized"
		Log.Error(err)
		return fmt.Errorf(err)
	}
	var Camera *C.Camera
	res := C.gp_camera_new((**C.Camera)(unsafe.Pointer(&Camera)))
	if res != OK {
		err := "Error get new camera: " + strconv.Itoa(int(res))
		Log.Error(err)
		return fmt.Errorf(err)
	}
	if Camera == nil {
		err := "could not initialize camera pointer"
		Log.Error(err)
		return fmt.Errorf(err)
	}
	c.Camera = Camera
	return nil
}

//init camera
func (c *Camera) InitCamera() error {
	Log.Trace("initializing camera")
	if c.Camera == nil {
		err := "could not initialize camera without pointer"
		Log.Error(err)
		return fmt.Errorf(err)
	}
	res := C.gp_camera_init(c.Camera, c.Context)
	if res != OK {
		err := "error camera initializing: " + strconv.Itoa(int(res))
		Log.Error(err)
		return fmt.Errorf(err)
	}
	c.CameraStatus = true
	return nil
}

//exit camera
func (c *Camera) ExitCamera() error {
	c.CameraStatus = false
	Log.Trace("exit camera")
	res := C.gp_camera_exit(c.Camera, c.Context)
	if res != OK {
		err := "error exit camera: " + strconv.Itoa(int(res))
		Log.Error(err)
		return fmt.Errorf(err)
	}
	// c.Camera = nil
	return nil
}

//unref camera
func (c *Camera) UnrefCamera() error {
	Log.Trace("unref camera")
	res := C.gp_camera_unref(c.Camera)
	if res != OK {
		err := "error unref camera: " + strconv.Itoa(int(res))
		Log.Error(err)
		return fmt.Errorf(err)
	}
	c.Camera = nil
	return nil
}

//CapturePreview  captures image preview and saves it in provided buffer
func (c *Camera) CapturePreview(buffer io.Writer) error {
	Log.Trace("capture preview")
	gpFile, err := newFile()
	if err != nil {
		Log.Error(err.Error())
		return err
	}

	if res := C.gp_camera_capture_preview(c.Camera, gpFile, c.Context); res != OK {
		var yy *C.char
		yy = C.gp_port_result_as_string(res)
		fmt.Println(C.GoString(yy))
		err := "cannot capture preview, error code: " + strconv.Itoa(int(res))
		Log.Error(err)
		if gpFile != nil {
			C.gp_file_unref(gpFile)
		}
		return fmt.Errorf(err)
	}

	res := getFileBytes(gpFile, buffer)

	if gpFile != nil {
		C.gp_file_unref(gpFile)
	}
	return res
}

//Capture photo
func (c *Camera) CapturePhoto(buffer *bytes.Buffer) error {
	Log.Trace("capture photo")
	type cameraFilePathInternal struct {
		Name   [128]uint8
		Folder [1024]uint8
	}
	photoPath := cameraFilePathInternal{}
	res := C.gp_camera_capture(c.Camera, 0, (*C.CameraFilePath)(unsafe.Pointer(&photoPath)), c.Context)
	if res != OK {
		err := "cannot capture photo, error code: " + strconv.Itoa(int(res))
		Log.Error(err)
		return fmt.Errorf(err)
	}
	buff := io.Writer(buffer)
	filePath := &CameraFilePath{
		Name:     string(photoPath.Name[:bytes.IndexByte(photoPath.Name[:], 0)]),
		Folder:   string(photoPath.Folder[:bytes.IndexByte(photoPath.Folder[:], 0)]),
		Isdir:    false,
		Children: nil,
		camera:   c,
	}
	err := filePath.DownloadImage(buff, true)
	if err != nil {
		Log.Error(err.Error())
	}
	return nil
}

// Download image from camera.
func (file *CameraFilePath) DownloadImage(buffer io.Writer, leaveOnCamera bool) error {
	Log.Trace("download image")
	_file, err := newFile()
	if err != nil {
		Log.Error(err.Error())
		return err
	}
	defer C.gp_file_free(_file)

	fileDir := C.CString(file.Folder)
	defer C.free(unsafe.Pointer(fileDir))

	fileName := C.CString(file.Name)
	defer C.free(unsafe.Pointer(fileName))

	res := C.gp_camera_file_get(file.camera.Camera, fileDir, fileName, FileTypeNormal, _file, file.camera.Context)
	if res != OK {
		_err := "cannot download photo file, error code: " + strconv.Itoa(int(res))
		Log.Error(_err)
		return fmt.Errorf(_err)
	}

	err = getFileBytes(_file, buffer)
	if err != nil && !leaveOnCamera {
		C.gp_camera_file_delete(file.camera.Camera, fileDir, fileName, file.camera.Context)
		Log.Error(err.Error())
		return err
	}
	return nil
}
