
CMD=cmd
BASE=github.com/yunchih/s3-video-trans

REMOVER=minio-remover
UPLOADER=minio-uploader
TRANSCODER=minio-video-transcoder

.PHONY: clean docker

all: static $(REMOVER) $(UPLOADER) $(TRANSCODER)

static: static-$(REMOVER) static-$(UPLOADER) static-$(TRANSCODER)

%: $(CMD)/%
	go build $(BASE)/$<

static-%: $(CMD)/%
	CGO_ENABLED=0 GOOS=linux go build -o $@ -a -ldflags '-s -w -extldflags "-static"' $(BASE)/$<

docker: static
	docker build -t docker.io/ctld/minio-video-transcoder:latest .

clean:
	rm -f static-* $(REMOVER) $(UPLOADER) $(TRANSCODER)
