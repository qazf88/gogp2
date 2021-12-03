package gogp2

// #cgo linux pkg-config: libgphoto2
// #include <gphoto2/gphoto2.h>
// #include <string.h>
import "C"
import (
	"fmt"
	"io"
	"reflect"
	"strconv"
	"unsafe"

	Log "github.com/qazf88/golog"
)

func newFile() (*C.CameraFile, error) {
	Log.Trace("create file pointer")
	var file *C.CameraFile
	res := C.gp_file_new((**C.CameraFile)(unsafe.Pointer(&file)))
	if res != OK {
		err := "error create file:" + strconv.Itoa(int(res))
		Log.Error(err)
		return nil, fmt.Errorf(err)
	}
	if file == nil {
		err := "error create file pointer"
		Log.Error(err)
		return nil, fmt.Errorf(err)
	}
	Log.Trace("create file pointer ok")
	return file, nil
}

//get file bytes
func getFileBytes(gpFileIn *C.CameraFile, bufferOut io.Writer) error {
	Log.Trace("get data and size from camera file")
	var fileData *C.char
	var fileLen C.ulong
	res := C.gp_file_get_data_and_size(gpFileIn, (**C.char)(unsafe.Pointer(&fileData)), &fileLen)
	if res != OK {
		err := "error get data and size from camera file:" + strconv.Itoa(int(res))
		Log.Error(err)
		return fmt.Errorf(err)
	}

	hdr := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(fileData)),
		Len:  int(fileLen),
		Cap:  int(fileLen),
	}
	goSlice := *(*[]byte)(unsafe.Pointer(&hdr))
	_, err := bufferOut.Write(goSlice)
	if err != nil {
		Log.Error(err.Error())
		return err
	}
	Log.Trace("get data and size from camera file ok")
	return nil
}
