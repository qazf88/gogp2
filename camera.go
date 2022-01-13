package gogp2

// #cgo linux pkg-config: libgphoto2
// #include <gphoto2/gphoto2.h>
// #include <string.h>
// #include <stdlib.h>
import "C"
import (
	"fmt"
	"strconv"
	"unsafe"

	Log "github.com/qazf88/golog"
)

func (c *Camera) Model() (string, error) {

	var abilities C.CameraAbilities
	res := C.gp_camera_get_abilities(c.Camera, &abilities)
	if res != OK {
		return "", fmt.Errorf("error get model :%v", res)
	}

	model := C.GoString((*C.char)(&abilities.model[0]))

	return model, nil
}

func (c *Camera) Init() bool {

	if c.Context == nil {
		if c.NewContext() != nil {
			return false
		}
	}

	if c.Camera == nil {
		if c.NewCamera() != nil {
			return false
		}
	}

	if c.InitCamera() != nil {
		return false
	}

	Log.Info("New camera")

	return true
}

func (c *Camera) NewCamera() error {

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

func (c *Camera) InitCamera() error {

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

	return nil
}

func (c *Camera) AvalibleCamera() bool {

	err := c.InitCamera()
	return err == nil

}

func (c *Camera) FreeCamera() error {

	res := C.gp_camera_exit(c.Camera, c.Context)
	if res != OK {
		err := "error exit camera: " + strconv.Itoa(int(res))
		Log.Error(err)
		return fmt.Errorf(err)
	}

	return nil
}

func (c *Camera) UnrefCamera() error {

	res := C.gp_camera_unref(c.Camera)
	if res != OK {
		err := "error unref camera: " + strconv.Itoa(int(res))
		Log.Error(err)
		return fmt.Errorf(err)
	}

	c.Camera = nil
	return nil
}

func (c *Camera) ReInitCamera() error {
	res := C.gp_camera_exit(c.Camera, c.Context)
	if res != OK {
		err := "error exit camera: " + strconv.Itoa(int(res))
		Log.Error(err)
		return fmt.Errorf(err)
	}

	res = C.gp_camera_unref(c.Camera)
	if res != OK {
		err := "error unref camera: " + strconv.Itoa(int(res))
		Log.Error(err)
		return fmt.Errorf(err)
	}

	c.Camera = nil
	return nil

}
