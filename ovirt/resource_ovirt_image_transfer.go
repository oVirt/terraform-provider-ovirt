// Copyright (C) 2019 oVirt Maintainers
// All rights reserved.
//
// This software may be modified and distributed under the terms
// of the BSD-2 license.  See the LICENSE file for details.

package ovirt

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	govirt "github.com/ovirt/go-ovirt-client"
	ovirtsdk4 "github.com/ovirt/go-ovirt"
)

const BufferSize = 50 * 1048576 // 50MiB

func resourceOvirtImageTransfer() *schema.Resource {
	return &schema.Resource{
		Create: resourceOvirtImageTransferCreate,
		Read:   resourceOvirtImageTransferRead,
		Delete: resourceOvirtImageTransferDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			// the name of the uploaded image
			"alias": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"source_url": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"storage_domain_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"sparse": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"disk_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceOvirtImageTransferCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(govirt.ClientWithLegacySupport)
	conn := client.GetSDKClient()

	alias := d.Get("alias").(string)
	sourceUrl := d.Get("source_url").(string)
	domainId := d.Get("storage_domain_id").(string)
	sparse := d.Get("sparse").(bool)

	ctx, cancel := context.WithTimeout(context.Background(), 60 * time.Minute)
	defer cancel()

	reader, size, err := LoadSourceURL(sourceUrl)
	if err != nil {
		return err
	}
	defer func() {
		_ = reader.Close()
	}()

	uploadResult, err := client.UploadImage(
		ctx,
		alias,
		domainId,
		sparse,
		uint64(size),
		reader,
	)
	if err != nil {
		return err
	}

	jobFinishedConf := &resource.StateChangeConf{
		// An empty list indicates all jobs are completed
		Target:     []string{
			string(ovirtsdk4.JOBSTATUS_STARTED),
			string(ovirtsdk4.JOBSTATUS_FINISHED),
		},
		Refresh:    jobRefreshFunc(conn, uploadResult.CorrelationID()),
		Timeout:    d.Timeout(schema.TimeoutUpdate),
		Delay:      10 * time.Second,
		MinTimeout: 15 * time.Second,

	}
	if _, err = jobFinishedConf.WaitForState(); err != nil {
		return fmt.Errorf("failed to wait for finished state (%w)", err)
	}

	d.SetId(uploadResult.Disk().ID())
	if err := d.Set("disk_id", uploadResult.Disk().ID()); err != nil {
		return fmt.Errorf("failed to set disk_id on Terraform resource (%w)", err)
	}

	return resourceOvirtDiskRead(d, meta)
}

func resourceOvirtImageTransferRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(govirt.ClientWithLegacySupport).GetSDKClient()
	getDiskResp, err := conn.SystemService().DisksService().
		DiskService(d.Id()).Get().Send()
	if err != nil {
		return err
	}

	disk, ok := getDiskResp.Disk()
	if !ok {
		d.SetId("")
		return nil
	}
	d.Set("name", disk.MustAlias())
	d.Set("size", disk.MustProvisionedSize()/int64(math.Pow(2, 30)))
	d.Set("format", disk.MustFormat())
	d.Set("disk_id", disk.MustId())

	if sds, ok := disk.StorageDomains(); ok {
		if len(sds.Slice()) > 0 {
			d.Set("storage_domain_id", sds.Slice()[0].MustId())
		}
	}
	if alias, ok := disk.Alias(); ok {
		d.Set("alias", alias)
	}
	if shareable, ok := disk.Shareable(); ok {
		d.Set("shareable", shareable)
	}
	if sparse, ok := disk.Sparse(); ok {
		d.Set("sparse", sparse)
	}

	return nil
}

func resourceOvirtImageTransferDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(govirt.ClientWithLegacySupport)
	conn := client.GetSDKClient()
	diskService := conn.SystemService().
		DisksService().
		DiskService(d.Id())

	return resource.Retry(2*time.Minute, func() *resource.RetryError {
		log.Printf("[DEBUG] Now to remove Disk (%s)", d.Id())
		_, e := diskService.Remove().Send()
		if e != nil {
			if _, ok := e.(*ovirtsdk4.NotFoundError); ok {
				log.Printf("[DEBUG] Disk (%s) has been removed", d.Id())
				return nil
			}
			return resource.RetryableError(fmt.Errorf("Error removing Disk (%s): %s", d.Id(), e))
		}
		return resource.RetryableError(fmt.Errorf("Disk (%s) is still being removed", d.Id()))
	})
}

func LoadSourceURL(sourceURL string) (reader io.ReadCloser, size uint, err error) {
	if strings.HasPrefix(sourceURL, "file://") || strings.HasPrefix(sourceURL, "/") {
		sourceURL = strings.TrimPrefix(sourceURL, "file://")
		// skip url download, its a local file
		fh, err := os.Open(sourceURL)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to open file %s (%w)", sourceURL, err)
		}
		stat, err := fh.Stat()
		if err != nil {
			return nil, 0, fmt.Errorf("failed to stat %s (%w)", sourceURL, err)
		}
		size = uint(stat.Size())
		return fh, size, nil
	} else {
		resp, err := http.Get(sourceURL)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to open remoteURL %s (%w)", sourceURL, err)
		}

		// We are buffering the image locally because:
		// - The server may not send a content-length header.
		// - Downloading may be too slow for directly piping it to the oVirt image proxy.
		sFile, err := ioutil.TempFile(os.TempDir(), "*-ovirt-image.downloaded")
		if err != nil {
			return nil, 0, fmt.Errorf("failed to create temporary file for image download (%w)", err)
		}
		tempFileName := sFile.Name()
		if _, err := io.Copy(sFile, resp.Body); err != nil {
			_ = resp.Body.Close()
			_ = sFile.Close()
			_ = os.Remove(tempFileName)
			return nil, 0, fmt.Errorf(
				"failed to download image from %s to local temporary file (%w)",
				sourceURL,
				err,
			)
		}
		if err := resp.Body.Close(); err != nil {
			_ = sFile.Close()
			_ = os.Remove(tempFileName)
			return nil, 0, fmt.Errorf(
				"failed to close download file handle from %s (%w)",
				sourceURL,
				err,
			)
		}
		if err := sFile.Close(); err != nil {
			_ = os.Remove(tempFileName)
			return nil, 0, fmt.Errorf(
				"failed to close temporary image file %s (%w)",
				tempFileName,
				err,
			)
		}

		fh, err := os.Open(tempFileName)
		if err != nil {
			_ = os.Remove(sFile.Name())
			return nil, 0, fmt.Errorf("failed to open temporary image file after download (%w)", err)
		}

		return &deletingReader{
			tempFileName: tempFileName,
			fh: fh,
		}, size, nil
	}
}

type deletingReader struct {
	fh           *os.File
	tempFileName string
}

func (d *deletingReader) Read(p []byte) (n int, err error) {
	return d.fh.Read(p)
}

func (d *deletingReader) Close() error {
	if err := d.fh.Close(); err != nil {
		_ = os.Remove(d.tempFileName)
		return fmt.Errorf("failed to close temporary image file %s (%w)", d.tempFileName, err)
	}
	if err := os.Remove(d.tempFileName); err != nil {
		return fmt.Errorf("failed to remove temporary image file %s (%w)", d.tempFileName, err)
	}
	return nil
}

func jobRefreshFunc(conn *ovirtsdk4.Connection, correlationID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		jobResp, err := conn.SystemService().JobsService().List().Search(fmt.Sprintf("correlation_id=%s", correlationID)).Send()
		if err != nil {
			return nil, "", fmt.Errorf("failed to list jobs (%w)", err)
		}
		if jobSlice, ok := jobResp.Jobs(); ok && len(jobSlice.Slice()) > 0 {
			jobs := jobSlice.Slice()
			for _, job := range jobs {
				if status, ok := job.Status(); ok {
					return job, string(status), nil
				}
			}
		}

		return nil, string(ovirtsdk4.JOBSTATUS_UNKNOWN), fmt.Errorf("job status unknown")
	}
}
