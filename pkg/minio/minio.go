package minio
import (
    "context"
    "io"
    "os"
    "time"

    "github.com/yunchih/s3-video-trans/pkg/s3"
    "github.com/minio/minio-go"
)

type Minio struct {
    client *minio.Client
}

const (
    objectContentType = "application/octet-stream"
    downloadTimeout = 3 * time.Minute
)

func New (cfg s3.Config) (Minio, error) {
    client, err := minio.New(cfg.Endpoint, cfg.AccessID, cfg.AccessKey, cfg.UseSSL)
    return Minio{client}, err
}

func (m *Minio) Upload (bucket string, obj string, file *os.File) error {
    stat, err := file.Stat()
    if err != nil {
        return err
    }

    opt := minio.PutObjectOptions{ContentType: objectContentType}
    _, err = m.client.PutObject(bucket, obj, file, stat.Size(), opt)
    return err
}

func (m *Minio) Download (bucket string, obj string, file *os.File) error {
    ctx, cancel := context.WithTimeout(context.Background(), downloadTimeout)
    defer cancel()

    opt := minio.GetObjectOptions{}
    objData, err := m.client.GetObjectWithContext(ctx, bucket, obj, opt)
    if err != nil {
        return err
    }

     _, err = io.Copy(file, objData)
     return err
}

func (m *Minio) List(bucket string) (s3.Objects, error) {
    finish := make(chan struct{})
    defer close(finish)

    var result s3.Objects
    recursively := true
    objChan := m.client.ListObjectsV2(bucket, "", recursively, finish)
    for obj := range objChan {
        if obj.Err != nil {
            return result, obj.Err
        }

        result = append(result, s3.Object{obj.Key, obj.LastModified.Unix()})
    }

    return result, nil
}

func (m *Minio) Remove (bucket string, obj string) error {
    return m.client.RemoveObject(bucket, obj)
}
