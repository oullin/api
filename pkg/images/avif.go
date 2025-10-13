//go:build cgo

package images

/*
#cgo pkg-config: libheif
#include <stdlib.h>
#include <libheif/heif.h>
*/
import "C"

import (
	"errors"
	"fmt"
	stdimage "image"
	"image/color"
	"unsafe"
)

const avifSupported = true

type avifImage struct {
	img stdimage.Image
	cfg stdimage.Config
}

func decodeAVIF(data []byte) (stdimage.Image, error) {
	decoded, err := decodeAVIFInternal(data, false)
	if err != nil {
		return nil, err
	}

	return decoded.img, nil
}

func decodeAVIFConfig(data []byte) (stdimage.Config, error) {
	decoded, err := decodeAVIFInternal(data, true)
	if err != nil {
		return stdimage.Config{}, err
	}

	return decoded.cfg, nil
}

func decodeAVIFInternal(data []byte, configOnly bool) (avifImage, error) {
	if len(data) == 0 {
		return avifImage{}, errors.New("avif payload empty")
	}

	ctx := C.heif_context_alloc()
	if ctx == nil {
		return avifImage{}, errors.New("heif context alloc failed")
	}
	defer C.heif_context_free(ctx)

	ptr := C.CBytes(data)
	defer C.free(ptr)

	size := C.size_t(len(data))

	err := C.heif_context_read_from_memory(ctx, ptr, size, nil)
	if herr := translateHeifError("read avif payload", err); herr != nil {
		return avifImage{}, herr
	}

	var handle *C.struct_heif_image_handle
	err = C.heif_context_get_primary_image_handle(ctx, &handle)
	if herr := translateHeifError("get avif handle", err); herr != nil {
		return avifImage{}, herr
	}
	defer C.heif_image_handle_release(handle)

	width := int(C.heif_image_handle_get_width(handle))
	height := int(C.heif_image_handle_get_height(handle))
	if width <= 0 || height <= 0 {
		return avifImage{}, fmt.Errorf("invalid avif dimensions %dx%d", width, height)
	}

	cfg := stdimage.Config{ColorModel: color.NRGBAModel, Width: width, Height: height}

	if configOnly {
		return avifImage{cfg: cfg}, nil
	}

	var img *C.struct_heif_image
	err = C.heif_decode_image(handle, &img, C.heif_colorspace_RGB, C.heif_chroma_interleaved_RGBA, nil)
	if herr := translateHeifError("decode avif", err); herr != nil {
		return avifImage{}, herr
	}
	defer C.heif_image_release(img)

	var stride C.int
	plane := C.heif_image_get_plane_readonly(img, C.heif_channel_interleaved, &stride)
	if plane == nil {
		return avifImage{}, errors.New("avif plane unavailable")
	}

	rowStride := int(stride)
	if rowStride < width*4 {
		return avifImage{}, fmt.Errorf("avif stride smaller than width: %d < %d", rowStride, width*4)
	}

	total := rowStride * height
	if total <= 0 {
		return avifImage{}, fmt.Errorf("invalid avif buffer size: %d", total)
	}

	buf := C.GoBytes(unsafe.Pointer(plane), C.int(total))

	out := stdimage.NewNRGBA(stdimage.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		srcStart := y * rowStride
		srcEnd := srcStart + width*4
		dstStart := y * out.Stride
		dstEnd := dstStart + width*4
		copy(out.Pix[dstStart:dstEnd], buf[srcStart:srcEnd])
	}

	return avifImage{img: out, cfg: cfg}, nil
}

func translateHeifError(context string, err C.struct_heif_error) error {
	if err.code == C.heif_error_Ok {
		return nil
	}

	message := "heif"
	if err.message != nil {
		message = C.GoString(err.message)
	}

	return fmt.Errorf("%s: %s (code=%d subcode=%d)", context, message, int(err.code), int(err.subcode))
}
