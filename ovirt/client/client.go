package client

import (
	"context"
	"io"
)

// OvirtClient is a simplified client for the oVirt API.
type OvirtClient interface {
	// UploadImage uploads an image file into a disk. The actual upload takes place in the
	// background and can be tracked using the returned UploadImageProgress object.
	//
	// Parameters are as follows:
	//
	// - ctx: this context can be used to abort the upload if it takes too long.
	// - alias: this is the name used for the uploaded image.
	// - storageDomainID: this is the UUID of the storage domain that the image should be uploaded to.
	// - sparse: use sparse provisioning
	// - size: this is the file size of the image. This must match the bytes read.
	// - reader: this is the source of the image data.
	//
	// You can wait for the upload to complete using the Done() method:
	//
	//     progress, err := cli.UploadImage(...)
	//     if err != nil {
	//         //...
	//     }
	//     <-progress.Done()
	//
	// After the upload is complete you can check the Err() method if it completed successfully:
	//
	//     if err := progress.Err(); err != nil {
	//         //...
	//     }
	//
	UploadImage(
		ctx context.Context,
		alias string,
		storageDomainID string,
		sparse bool,
		size uint64,
		reader io.Reader,
	) (UploadImageProgress, error)
}

// UploadImageProgress is a tracker for the upload progress happening in the background.
type UploadImageProgress interface {
	// UploadedBytes returns the number of bytes already uploaded.
	UploadedBytes() uint64
	// TotalBytes returns the total number of bytes to be uploaded.
	TotalBytes() uint64
	// Err returns the error of the upload once the upload is complete or errored.
	Err() error
	// Done returns a channel that will be closed when the upload is complete.
	Done() <-chan struct{}
}

// ImageFormat is a constant for representing the format that images can be in.
type ImageFormat string

// ImageFormatCow
const ImageFormatCow ImageFormat = "cow"
const ImageFormatRaw ImageFormat = "raw"
