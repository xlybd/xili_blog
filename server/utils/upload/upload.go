package upload

import (
	"mime/multipart"
	"server/global"
	"server/model/appTypes"
)

var WhiteImageList = map[string]struct{}{
	".jpg":  {},
	".png":  {},
	".jpeg": {},
	".ico":  {},
	".tiff": {},
	".gif":  {},
	".svg":  {},
	".webp": {},
}

type OSS interface {
	UploadImage(file *multipart.FileHeader) (string, string, error)
	DeleteImage(key string) error
}

func NewOss() OSS {
	switch global.Config.System.OssType {
	case "local":
		return &Local{}
	case "qiniu":
		return &Qiniu{}
	default:
		return &Local{}
	}
}

func NewOssWithStorage(storage appTypes.Storage) OSS {
	switch storage {
	case appTypes.Local:
		return &Local{}
	case appTypes.Qiniu:
		return &Qiniu{}
	default:
		return &Local{}
	}
}
