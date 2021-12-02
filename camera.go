package gogp2

// #cgo linux pkg-config: libgphoto2
// #include <gphoto2/gphoto2.h>
// #include <string.h>
import "C"
import (
	"fmt"
	"strconv"
	"unsafe"

	Log "github.com/qazf88/golog"
)

//Init camera
func (c *Camera) InitCamera() bool {

	c.CameraStatus = false
	if c.NewContext() != nil {
		return false
	}
	if c.newCamera() != nil {
		return false
	}
	if c.initCamera() != nil {
		return false
	}
	c.CameraStatus = true
	Log.Info("New camera")
	return true
}

//Get new camera
func (c *Camera) newCamera() error {
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
func (c *Camera) initCamera() error {
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
		_err := c.FreeCamera()
		if _err != nil {
			Log.Error(_err.Error())
			return _err
		}
		return fmt.Errorf(err)
	}
	return nil
}

//exit camera
func (c *Camera) exitCamera() error {
	Log.Trace("exit camera")
	res := C.gp_camera_exit(c.Camera, c.Context)
	if res != OK {
		err := "error exit camera: " + strconv.Itoa(int(res))
		Log.Error(err)
		return fmt.Errorf(err)
	}
	c.Camera = nil
	return nil
}

//unref camera
func (c *Camera) unrefCamera() error {
	Log.Trace("unref camera")
	res := C.gp_camera_unref(c.Camera)
	if res != OK {
		err := "error unref camera: " + strconv.Itoa(int(res))
		Log.Error(err)
		return fmt.Errorf(err)
	}
	return nil
}

//Free camera
func (c *Camera) FreeCamera() error {
	Log.Trace("free camera")
	err := c.exitCamera()
	if err != nil {
		return err
	}
	err = c.unrefCamera()
	if err != nil {
		return err
	}
	c.CameraStatus = false
	return nil
}
