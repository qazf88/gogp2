package gogp2

// #cgo linux pkg-config: libgphoto2
// #include <gphoto2/gphoto2.h>
// #include <string.h>
// #include <stdlib.h>
import "C"
import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"unsafe"

	Log "github.com/qazf88/golog"
)

// DownloadImage
func (c *Camera) DownloadImage(buffer io.Writer, file *CameraFilePath, leaveOnCamera bool) error {

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
		_err := fmt.Sprintf("cannot download photo by name %s, error code: %d", C.GoString(fileName), res)
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

// newFile
func newFile() (*C.CameraFile, error) {

	var file *C.CameraFile
	res := C.gp_file_new((**C.CameraFile)(unsafe.Pointer(&file)))
	if res != OK {
		err := fmt.Sprintf("error create file, error code: %d", res)
		Log.Error(err)
		return nil, fmt.Errorf(err)
	}

	if file == nil {
		err := fmt.Sprintf("error create file pointer, error code: %d", res)
		Log.Error(err)
		return nil, fmt.Errorf(err)
	}
	return file, nil
}

// getFileBytes
func getFileBytes(gpFileIn *C.CameraFile, bufferOut io.Writer) error {

	var fileData *C.char
	var fileLen C.ulong
	res := C.gp_file_get_data_and_size(gpFileIn, (**C.char)(unsafe.Pointer(&fileData)), &fileLen)
	if res != OK {
		err := fmt.Sprintf("error get data and size from camera file: error code: %d", res)
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
	return nil
}

// DeleteFile
func (c *Camera) DeleteFile(path *CameraFilePath) error {

	fileDir := C.CString(path.Folder)
	defer C.free(unsafe.Pointer(fileDir))

	fileName := C.CString(path.Name)
	defer C.free(unsafe.Pointer(fileName))

	res := C.gp_camera_file_delete(c.Camera, fileDir, fileName, c.Context)
	if res != OK {
		err := fmt.Sprintf("cannot delete fine on camera, error code :%d", res)
		Log.Error(err)
		return fmt.Errorf(err)
	}

	return nil
}

// ListFolders
func (c *Camera) ListFolders(folder string) ([]string, error) {
	if folder == "" {
		folder = "/"
	}

	var cameraList *C.CameraList
	C.gp_list_new(&cameraList)
	defer C.free(unsafe.Pointer(cameraList))

	C_folder := C.CString(folder)
	defer C.free(unsafe.Pointer(C_folder))

	C.gp_camera_folder_list_folders(c.Camera, C_folder, cameraList, c.Context)

	folderMap, _ := cameraListToMap(cameraList)

	names := make([]string, len(folderMap))
	i := 0
	for key, _ := range folderMap {
		names[i] = key
		i += 1
	}

	return names, nil
}

// RecursiveListFolders
func (c *Camera) RecursiveListFolders(folder string) []string {
	folders := make([]string, 0)
	path := folder
	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}
	subfolders, _ := c.ListFolders(path)
	for _, sub := range subfolders {
		subPath := path + sub
		folders = append(folders, subPath)
		folders = append(folders, c.RecursiveListFolders(subPath)...)
	}

	return folders
}

// ListFiles
func (c *Camera) ListFiles(folder string) ([]string, int) {
	if folder == "" {
		folder = "/"
	}

	if !strings.HasSuffix(folder, "/") {
		folder = folder + "/"
	}

	var cameraList *C.CameraList
	C.gp_list_new(&cameraList)
	defer C.free(unsafe.Pointer(cameraList))

	cFolder := C.CString(folder)
	defer C.free(unsafe.Pointer(cFolder))

	err := C.gp_camera_folder_list_files(c.Camera, cFolder, cameraList, c.Context)
	fileNameMap, _ := cameraListToMap(cameraList)

	names := make([]string, len(fileNameMap))
	i := 0
	for key, _ := range fileNameMap {
		names[i] = key
		i += 1
	}

	return names, int(err)
}

// cameraListToMap
func cameraListToMap(cameraList *C.CameraList) (map[string]string, int) {

	size := int(C.gp_list_count(cameraList))
	vals := make(map[string]string)

	if size < 0 {
		return vals, size
	}

	for i := 0; i < size; i++ {

		var C_key *C.char
		var C_val *C.char

		C.gp_list_get_name(cameraList, C.int(i), &C_key)
		C.gp_list_get_value(cameraList, C.int(i), &C_val)
		defer C.free(unsafe.Pointer(C_key))
		defer C.free(unsafe.Pointer(C_val))

		key := C.GoString(C_key)
		val := C.GoString(C_val)

		vals[key] = val
	}

	return vals, 0
}
