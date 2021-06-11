//
// This file implements the image upload-related functions of the oVirt client.
//

package govirt

import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	ovirtsdk4 "github.com/ovirt/go-ovirt"
)

func (o *oVirtClient) UploadImage(
	ctx context.Context,
	alias string,
	storageDomainID string,
	sparse bool,
	size uint64,
	reader io.Reader,
) (Disk, error) {
	progress, err := o.StartImageUpload(ctx, alias, storageDomainID, sparse, size, reader)
	if err != nil {
		return nil, err
	}
	<-progress.Done()
	if err := progress.Err(); err != nil {
		return nil, err
	}
	return progress.Disk(), nil
}

func (o *oVirtClient) StartImageUpload(
	ctx context.Context,
	alias string,
	storageDomainID string,
	sparse bool,
	size uint64,
	reader io.Reader,
) (UploadImageProgress, error) {
	bufReader := bufio.NewReaderSize(reader, qcowHeaderSize)

	format := ImageFormatCow
	qcowSize := size
	header, err := bufReader.Peek(qcowHeaderSize)
	if err != nil {
		return nil, fmt.Errorf("failed to read QCOW header (%w)", err)
	}
	isQCOW := string(header[0:len(qcowMagicBytes)]) == qcowMagicBytes
	if !isQCOW {
		format = ImageFormatRaw
	} else {
		// See https://people.gnome.org/~markmc/qcow-image-format.html
		qcowSize = binary.BigEndian.Uint64(header[qcowSizeStartByte : qcowSizeStartByte+8])
	}

	newCtx, cancel := context.WithCancel(ctx) //nolint:govet

	storageDomain, err := ovirtsdk4.NewStorageDomainBuilder().Id(storageDomainID).Build()
	if err != nil {
		panic(fmt.Errorf("bug: failed to build storage domain object from storage domain ID: %s", storageDomainID))
	}
	diskBuilder := ovirtsdk4.NewDiskBuilder().
		Alias(alias).
		Format(ovirtsdk4.DiskFormat(format)).
		ProvisionedSize(int64(qcowSize)).
		InitialSize(int64(qcowSize)).
		StorageDomainsOfAny(storageDomain)
	diskBuilder.Sparse(sparse)
	disk, err := diskBuilder.Build()
	if err != nil {
		cancel()
		return nil, fmt.Errorf(
			//nolint:govet
			"failed to build disk with alias %s, format %s, provisioned and initial size %d (%w)",
			alias,
			format,
			qcowSize,
			err,
		)
	}

	progress := &uploadImageProgress{
		uploadedBytes:   0,
		cowSize:         qcowSize,
		size:            size,
		reader:          bufReader,
		storageDomainID: storageDomainID,
		sparse:          sparse,
		alias:           alias,
		ctx:             newCtx,
		done:            make(chan struct{}),
		lock:            &sync.Mutex{},
		cancel:          cancel,
		err:             nil,
		conn:            o.conn,
		httpClient:      o.httpClient,
		disk:            disk,
		client:          o,
	}
	go progress.upload()
	return progress, nil
}

type uploadImageProgress struct {
	uploadedBytes   uint64
	cowSize         uint64
	size            uint64
	reader          *bufio.Reader
	storageDomainID string
	sparse          bool
	alias           string

	// ctx is the context that indicates that the upload should terminate as soon as possible. The actual upload may run
	// longer in order to facilitate proper cleanup.
	ctx context.Context
	// done is the channel that is closed when the upload is completely done, either with an error, or successfully.
	done chan struct{}
	// lock is a lock that prevents race conditions during the upload process.
	lock *sync.Mutex
	// cancel is the cancel function for the context. Is is called to ensure that the context is properly canceled.
	cancel context.CancelFunc
	// err holds the error that happened during the upload. It can be queried using the Err() method.
	err error
	// conn is the underlying oVirt connection.
	conn *ovirtsdk4.Connection
	// httpClient is the raw HTTP client for connecting the oVirt Engine.
	httpClient http.Client
	// disk is the oVirt disk that will be provisioned during the upload.
	disk *ovirtsdk4.Disk
	// client is the Client instance that created this image upload.
	client *oVirtClient
}

func (u *uploadImageProgress) Disk() Disk {
	sdkDisk := u.disk
	id, ok := sdkDisk.Id()
	if !ok || id == "" {
		return nil
	}
	disk, err := convertSDKDisk(sdkDisk)
	if err != nil {
		panic(fmt.Errorf("bug: failed to convert disk (%w)", err))
	}
	return disk
}

func (u *uploadImageProgress) UploadedBytes() uint64 {
	return u.uploadedBytes
}

func (u *uploadImageProgress) TotalBytes() uint64 {
	return u.size
}

func (u *uploadImageProgress) Err() error {
	u.lock.Lock()
	defer u.lock.Unlock()
	if u.err != nil {
		return u.err
	}
	return nil
}

func (u *uploadImageProgress) Done() <-chan struct{} {
	return u.done
}

func (u *uploadImageProgress) Read(p []byte) (n int, err error) {
	select {
	case <-u.ctx.Done():
		return 0, fmt.Errorf("timeout while uploading image")
	default:
	}
	n, err = u.reader.Read(p)
	u.uploadedBytes += uint64(n)
	return
}

// upload uploads the image file in the background. It is intended to be called as a goroutine. The error status can
// be obtained from Err(), while the done status can be queried using Done().
func (u *uploadImageProgress) upload() {
	defer func() {
		// Cancel context to indicate done.
		u.lock.Lock()
		u.cancel()
		close(u.done)
		u.lock.Unlock()
	}()

	if err := u.processUpload(); err != nil {
		u.err = err
	}
}

func (u *uploadImageProgress) processUpload() error {
	correlationID := fmt.Sprintf("image_transfer_%s", u.alias)
	diskID, diskService, err := u.createDisk(correlationID)
	if err != nil {
		return err
	}

	if err := u.waitForDiskOk(diskService); err != nil {
		u.removeDisk()
		return err
	}

	transfer, transferService, err := u.setupImageTransfer(diskID, correlationID)
	if err != nil {
		u.removeDisk()
		return err
	}

	transferURL, err := u.findTransferURL(transfer)
	if err != nil {
		u.removeDisk()
		return err
	}

	if err := u.uploadImage(transferURL); err != nil {
		u.removeDisk()
		return err
	}

	if err := u.finalizeUpload(transferService, correlationID); err != nil {
		u.removeDisk()
		return err
	}

	if err := u.waitForDiskOk(diskService); err != nil {
		u.removeDisk()
		return err
	}

	return nil
}

func (u *uploadImageProgress) removeDisk() {
	disk := u.disk
	if disk != nil {
		if id, ok := u.disk.Id(); ok {
			_ = u.client.RemoveDisk(id)
		}
	}
}

func (u *uploadImageProgress) finalizeUpload(
	transferService *ovirtsdk4.ImageTransferService,
	correlationID string,
) error {
	finalizeRequest := transferService.Finalize()
	finalizeRequest.Query("correlation_id", correlationID)
	_, err := finalizeRequest.Send()
	if err != nil {
		return fmt.Errorf("failed to finalize image upload (%w)", err)
	}
	return nil
}

func (u *uploadImageProgress) uploadImage(transferURL *url.URL) error {
	putRequest, err := http.NewRequest(http.MethodPut, transferURL.String(), u)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request (%w)", err)
	}
	putRequest.Header.Add("content-type", "application/octet-stream")
	putRequest.ContentLength = int64(u.size)
	_, err = u.httpClient.Do(putRequest)
	if err != nil {
		return fmt.Errorf("failed to upload image (%w)", err)
	}
	return nil
}

func (u *uploadImageProgress) findTransferURL(transfer *ovirtsdk4.ImageTransfer) (*url.URL, error) {
	var tryURLs []string
	if transferURL, ok := transfer.TransferUrl(); ok && transferURL != "" {
		tryURLs = append(tryURLs, transferURL)
	}
	if proxyURL, ok := transfer.ProxyUrl(); ok && proxyURL != "" {
		tryURLs = append(tryURLs, proxyURL)
	}

	if len(tryURLs) == 0 {
		return nil, fmt.Errorf("neither a transfer URL nor a proxy URL was returned from the oVirt Engine")
	}

	var foundTransferURL *url.URL
	var lastError error
	for _, transferURL := range tryURLs {
		transferURL, err := url.Parse(transferURL)
		if err != nil {
			lastError = fmt.Errorf("failer to parse transfer URL %s (%w)", transferURL, err)
			continue
		}

		hostUrl, err := url.Parse(transfer.MustTransferUrl())
		if err == nil {
			optionsReq, err := http.NewRequest(http.MethodOptions, hostUrl.String(), strings.NewReader(""))
			if err != nil {
				lastError = err
				continue
			}
			res, err := u.httpClient.Do(optionsReq)
			if err == nil {
				if res.StatusCode == 200 {
					foundTransferURL = transferURL
					lastError = nil
					break
				} else {
					lastError = fmt.Errorf("non-200 status code returned from URL %s (%d)", hostUrl, res.StatusCode)
				}
			} else {
				lastError = err
			}
		} else {
			lastError = err
		}
	}
	if foundTransferURL == nil {
		return nil, fmt.Errorf("failed to find transfer URL (last error: %w)", lastError)
	}
	return foundTransferURL, nil
}

func (u *uploadImageProgress) createDisk(correlationID string) (string, *ovirtsdk4.DiskService, error) {
	addDiskRequest := u.conn.SystemService().DisksService().Add().Disk(u.disk)
	addDiskRequest.Query("correlation_id", correlationID)
	addResp, err := addDiskRequest.Send()
	if err != nil {
		diskAlias, _ := u.disk.Alias()
		return "", nil, fmt.Errorf("failed to create disk, alias: %s (%w)", diskAlias, err)
	}
	diskID := addResp.MustDisk().MustId()
	diskService := u.conn.SystemService().DisksService().DiskService(diskID)
	return diskID, diskService, nil
}

func (u *uploadImageProgress) setupImageTransfer(diskID string, correlationID string) (
	*ovirtsdk4.ImageTransfer,
	*ovirtsdk4.ImageTransferService,
	error,
) {
	imageTransfersService := u.conn.SystemService().ImageTransfersService()
	image := ovirtsdk4.NewImageBuilder().Id(diskID).MustBuild()
	transfer := ovirtsdk4.
		NewImageTransferBuilder().
		Image(image).
		MustBuild()
	transferReq := imageTransfersService.
		Add().
		ImageTransfer(transfer).
		Query("correlation_id", correlationID)
	transferRes, err := transferReq.Send()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start image transfer (%w)", err)
	}
	transfer = transferRes.MustImageTransfer()
	transferService := imageTransfersService.ImageTransferService(transfer.MustId())

	for {
		req, lastError := transferService.Get().Send()
		if lastError == nil {
			if req.MustImageTransfer().MustPhase() == ovirtsdk4.IMAGETRANSFERPHASE_TRANSFERRING {
				break
			} else {
				lastError = fmt.Errorf(
					"image transfer is in phase %s instead of transferring",
					req.MustImageTransfer().MustPhase(),
				)
			}
		}
		select {
		case <-time.After(time.Second * 5):
		case <-u.ctx.Done():
			return nil, nil, fmt.Errorf("timeout while waiting for image transfer (last error was: %w)", lastError)
		}
	}
	return transfer, transferService, nil
}

func (u *uploadImageProgress) waitForDiskOk(diskService *ovirtsdk4.DiskService) error {
	var lastError error
	for {
		req, err := diskService.Get().Send()
		if err == nil {
			disk, ok := req.Disk()
			if !ok {
				return fmt.Errorf("the disk was removed after upload, probably not supported")
			}
			if disk.MustStatus() == ovirtsdk4.DISKSTATUS_OK {
				return nil
			} else {
				lastError = fmt.Errorf("disk status is %s, not ok", disk.MustStatus())
			}
			u.disk = disk
		} else {
			lastError = err
		}
		select {
		case <-time.After(5 * time.Second):
		case <-u.ctx.Done():
			return fmt.Errorf("timeout while waiting for disk to be ok after upload (last error: %w)", lastError)
		}
	}
}
