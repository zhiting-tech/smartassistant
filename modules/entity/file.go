package entity

type FileInfo struct {
	ID          int
	Name        string
	StorageType int
	Hash        string
	UserID      int
	FileType    int
	Extension   string
	FileAuth    int
}

func (fileInfo *FileInfo) TableName() string {
	return "file_info"
}

// GetFileInfo 通过Id查找文件
func GetFileInfo(fileInfoID int) (FileInfo, error) {
	var fileInfo FileInfo
	if err := GetDB().Where("id = ?", fileInfoID).First(&fileInfo).Error; err != nil {
		return fileInfo, err
	}
	return fileInfo, nil
}

// CreateFileInfo 增加表记录
func CreateFileInfo(fileInfo *FileInfo) error {
	return GetDB().Create(fileInfo).Error
}
