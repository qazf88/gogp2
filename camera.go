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

// Model
func (c *Camera) Model() (string, error) {

	var abilities C.CameraAbilities
	res := C.gp_camera_get_abilities(c.Camera, &abilities)
	if res != OK {
		return "", fmt.Errorf("error get model :%v", res)
	}

	model := C.GoString((*C.char)(&abilities.model[0]))

	return model, nil
}

// Init
func (c *Camera) Init() error {

	if c.Context == nil {
		err := c.NewContext()
		if err != nil {
			Log.Error(err.Error())
			return err
		}
	}

	if c.Camera != nil {
		//C.gp_camera_exit(c.Camera, c.Context)
		//C.gp_camera_unref(c.Camera)

		var Camera *C.Camera
		res := C.gp_camera_new((**C.Camera)(unsafe.Pointer(&Camera)))
		if res != OK {
			err := "Error get new camera: " + strconv.Itoa(int(res))
			Log.Error(err)
			return fmt.Errorf(err)
		}

		// res = C.gp_camera_ref(Camera)
		// if res != OK {
		// 	err := "error unref camera: " + strconv.Itoa(int(res))
		// 	Log.Error(err)
		// 	return fmt.Errorf(err)
		// }

		c.Camera = Camera
	}

	return nil

	err := c.InitCamera()
	if err != nil {
		Log.Error(err.Error())
		return err
	}

	Log.Info("New camera")

	return nil
}

// NewCamera
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
		err := fmt.Sprintf("Error get new camera: %d", res)
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

// InitCamera
func (c *Camera) InitCamera() error {

	if c.Camera == nil {
		err := "could not initialize camera without pointer"
		Log.Error(err)
		return fmt.Errorf(err)
	}

	res := C.gp_camera_init(c.Camera, c.Context)
	if res != OK {
		err := fmt.Sprintf("error camera initializing: %d", res)
		Log.Error(err)
		return fmt.Errorf(err)
	}

	res = C.gp_camera_exit(c.Camera, c.Context)
	if res != OK {
		fmt.Println(res)
	}

	return nil
}

// AvalibleCamera
func (c *Camera) AvalibleCamera() bool {

	err := c.InitCamera()
	return err == nil
}

// FreeCamera
func (c *Camera) FreeCamera() error {

	res := C.gp_camera_exit(c.Camera, c.Context)
	if res != OK {
		err := "error exit camera: " + strconv.Itoa(int(res))
		Log.Error(err)
		return fmt.Errorf(err)
	}

	return nil
}

// UnrefCamera
func (c *Camera) UnrefCamera() error {

	res := C.gp_camera_unref(c.Camera)
	if res != OK {
		err := "error unref camera: " + strconv.Itoa(int(res))
		Log.Error(err)
		return fmt.Errorf(err)
	}

	return nil
}

// RefCamera
func (c *Camera) RefCamera() error {

	res := C.gp_camera_ref(c.Camera)
	if res != OK {
		err := fmt.Sprintf("error ref camera: %d", res)
		Log.Error(err)
		return fmt.Errorf(err)
	}

	return nil
}

// HardResetCameraConnection !!! Warning !!!
func (c *Camera) HardResetCameraConnection() {

	Log.Info("Hard reset camera")

	C.gp_camera_free(c.Camera)
	c.Camera = nil
	c.Context = nil

}
