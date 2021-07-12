package ovirtclient

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/ioutil"
)

func (m *mockClient) StartImageUpload(
	_ context.Context,
	alias string,
	storageDomainID string,
	sparse bool,
	size uint64,
	reader io.Reader,
) (UploadImageProgress, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if alias == "" {
		return nil, fmt.Errorf("alias cannot be empty")
	}
	if _, ok := m.storageDomains[storageDomainID]; !ok {
		return nil, fmt.Errorf("storage domain with ID %s not found", storageDomainID)
	}

	bufReader := bufio.NewReaderSize(reader, qcowHeaderSize)

	format, _, err := extractQCOWParameters(size, bufReader)
	if err != nil {
		return nil, err
	}

	progress := &mockImageUploadProgress{
		err: nil,
		disk: disk{
			id:              "",
			alias:           alias,
			provisionedSize: size,
			format:          format,
			storageDomainID: storageDomainID,
		},
		correlationID: fmt.Sprintf("image_transfer_%s", alias),
		client:        m,
		reader:        bufReader,
		size:          size,
		done:          make(chan struct{}),
		sparse:        sparse,
	}

	go progress.do()

	return progress, nil
}

func (m *mockClient) UploadImage(
	ctx context.Context,
	alias string,
	storageDomainID string,
	sparse bool,
	size uint64,
	reader io.Reader,
) (UploadImageResult, error) {
	progress, err := m.StartImageUpload(ctx, alias, storageDomainID, sparse, size, reader)
	if err != nil {
		return nil, err
	}
	<-progress.Done()
	if err := progress.Err(); err != nil {
		return nil, err
	}
	return progress, nil
}

type mockImageUploadProgress struct {
	err           error
	disk          disk
	correlationID string
	client        *mockClient
	reader        *bufio.Reader
	size          uint64
	uploadedBytes uint64
	done          chan struct{}
	sparse        bool
}

func (m *mockImageUploadProgress) Disk() Disk {
	disk := m.disk
	if disk.id == "" {
		return nil
	}
	return disk
}

func (m *mockImageUploadProgress) CorrelationID() string {
	return m.correlationID
}

func (m *mockImageUploadProgress) UploadedBytes() uint64 {
	return m.uploadedBytes
}

func (m *mockImageUploadProgress) TotalBytes() uint64 {
	return m.size
}

func (m *mockImageUploadProgress) Err() error {
	return m.err
}

func (m *mockImageUploadProgress) Done() <-chan struct{} {
	return m.done
}

func (m *mockImageUploadProgress) do() {
	m.client.lock.Lock()
	d := m.disk
	d.id = m.client.GenerateUUID()
	m.client.disks[d.id] = d
	m.disk = d
	m.client.lock.Unlock()

	_, err := ioutil.ReadAll(m.reader)
	m.err = err
	if err != nil {
		m.uploadedBytes = m.size
	}
	close(m.done)
}
