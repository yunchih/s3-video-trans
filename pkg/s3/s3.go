package s3

type Config struct {
    Endpoint string
    AccessID string
    AccessKey string
    UseSSL bool
}

type Object struct {
    Key string
    LastModifiedUnix int64
}

type Objects []Object

type Service interface {
    Init(cfg Config) error
    Upload(bucket string, obj string, filename string) error
    Download(bucket string, obj string, filename string) error
    List(bucket string) (Objects, error)
}

