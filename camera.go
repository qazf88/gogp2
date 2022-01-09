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
func (c *CameraStruct) Init() bool {
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
	if c.InitCamera() != nil {
		return false
	}

	c.CameraStatus = true
	Log.Info("New camera")
	return true
}

//Get new camera
func (c *CameraStruct) NewCamera() error {
	Log.Trace("get new camera")
	if c.Context == nil {
		err := "could not get camera, context is empty"
		Log.Error(err)
		return fmt.Errorf(err)
	}
	if c.Camera != nil {
		err := "camera is already initialized"
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
func (c *CameraStruct) InitCamera() error {
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
	res = C.gp_camera_exit(c.Camera, c.Context)
	if res != OK {
		fmt.Println(res)
	}

	c.CameraStatus = true
	return nil
}

//exit camera
func (c *CameraStruct) ExitCamera() error {
	c.CameraStatus = false
	Log.Trace("exit camera")
	res := C.gp_camera_exit(c.Camera, c.Context)
	if res != OK {
		err := "error exit camera: " + strconv.Itoa(int(res))
		Log.Error(err)
		return fmt.Errorf(err)
	}
	return nil
}

//unref camera
func (c *CameraStruct) UnrefCamera() error {
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
func (c *CameraStruct) CapturePreview(buffer io.Writer) error {
	Log.Trace("capture preview")
	gpFile, err := newFile()
	if err != nil {
		Log.Error(err.Error())
		return err
	}

	if res := C.gp_camera_capture_preview(c.Camera, gpFile, c.Context); res != OK {
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
func (c *CameraStruct) CapturePhoto(buffer *bytes.Buffer) error {
	Log.Trace("capture photo")
	type cameraFilePathInternal struct {
		Name   [128]uint8
		Folder [1024]uint8
	}

	photoPath := cameraFilePathInternal{}
	res := C.gp_camera_capture(c.Camera, 0, (*C.CameraFilePath)(unsafe.Pointer(&photoPath)), c.Context)
	defer c.ExitCamera()
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
	}
	err := c.DownloadImage(buff, filePath, true)
	if err != nil {
		Log.Error(err.Error())
	}
	return nil
}

// Download image from camera.
func (c *CameraStruct) DownloadImage(buffer io.Writer, file *CameraFilePath, leaveOnCamera bool) error {
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

	res := C.gp_camera_file_get(c.Camera, fileDir, fileName, FileTypeNormal, _file, c.Context)
	if res != OK {
		_err := "cannot download photo file, error code: " + strconv.Itoa(int(res))
		Log.Error(_err)
		return fmt.Errorf(_err)
	}

	err = getFileBytes(_file, buffer)
	if err != nil && !leaveOnCamera {
		C.gp_camera_file_delete(c.Camera, fileDir, fileName, c.Context)
		Log.Error(err.Error())
		return err
	}
	return nil
}
