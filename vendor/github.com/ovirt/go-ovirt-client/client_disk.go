package ovirtclient

import (
	"context"
	"fmt"
	"io"

	ovirtsdk4 "github.com/ovirt/go-ovirt"
)

type DiskClient interface {
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
	//     progress, err := cli.StartImageUpload(...)
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
	StartImageUpload(
		ctx context.Context,
		alias string,
		storageDomainID string,
		sparse bool,
		size uint64,
		reader io.Reader,
	) (UploadImageProgress, error)

	// UploadImage is identical to StartImageUpload, but waits until the upload is complete. It returns the disk ID
	// as a result, or the error if one happened.
	UploadImage(
		ctx context.Context,
		alias string,
		storageDomainID string,
		sparse bool,
		size uint64,
		reader io.Reader,
	) (UploadImageResult, error)

	// ListDisks lists all disks.
	ListDisks() ([]Disk, error)
	// GetDisk fetches a disk with a specific ID from the
	GetDisk(diskID string) (Disk, error)
	// RemoveDisk removes a disk with a specific ID.
	RemoveDisk(diskID string) error
}

type UploadImageResult interface {
	Disk() Disk
	CorrelationID() string
}

type Disk interface {
	ID() string
	Alias() string
	ProvisionedSize() uint
	Format() ImageFormat
	StorageDomainID() string
}

// UploadImageProgress is a tracker for the upload progress happening in the background.
type UploadImageProgress interface {
	// Disk returns the disk created as part of the upload process once the upload is complete. Before the upload
	// is complete it will return nil.
	Disk() Disk
	// CorrelationID returns the correlation ID for the upload.
	CorrelationID() string
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

const (
	ImageFormatCow ImageFormat = "cow"
	ImageFormatRaw ImageFormat = "raw"
)

func convertSDKDisk(sdkDisk *ovirtsdk4.Disk) (Disk, error) {
	id, ok := sdkDisk.Id()
	if !ok {
		return nil, fmt.Errorf("disk does not contain an ID")
	}
	var storageDomainID string
	if sdkStorageDomain, ok := sdkDisk.StorageDomain(); ok {
		storageDomainID, _ = sdkStorageDomain.Id()
	}
	if storageDomainID == "" {
		if sdkStorageDomains, ok := sdkDisk.StorageDomains(); ok {
			if len(sdkStorageDomains.Slice()) == 1 {
				storageDomainID, _ = sdkStorageDomains.Slice()[0].Id()
			}
		}
	}
	if storageDomainID == "" {
		return nil, fmt.Errorf("failed to find a valid storage domain ID for disk %s", id)
	}
	alias, ok := sdkDisk.Alias()
	if !ok {
		return nil, fmt.Errorf("disk %s does not contain an alias", id)
	}
	provisionedSize, ok := sdkDisk.ProvisionedSize()
	if !ok {
		return nil, fmt.Errorf("disk %s does not contain a provisioned size", id)
	}
	format, ok := sdkDisk.Format()
	if !ok {
		return nil, fmt.Errorf("disk %s has no format field", id)
	}
	return &disk{
		id:              id,
		alias:           alias,
		provisionedSize: uint(provisionedSize),
		format:          ImageFormat(format),
		storageDomainID: storageDomainID,
	}, nil
}

type disk struct {
	id              string
	alias           string
	provisionedSize uint
	format          ImageFormat
	storageDomainID string
}

func (d disk) ID() string {
	return d.id
}

func (d disk) Alias() string {
	return d.alias
}

func (d disk) ProvisionedSize() uint {
	return d.provisionedSize
}

func (d disk) Format() ImageFormat {
	return d.format
}

func (d disk) StorageDomainID() string {
	return d.storageDomainID
}
