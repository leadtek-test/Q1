package file

type Workspace interface {
	Save(userID uint, fileName string, data []byte) (string, error)
}
