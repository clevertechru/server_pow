package quotes

var storage *Storage

func Init(configPath string) error {
	var err error
	storage, err = NewStorage(configPath)
	return err
}

func GetRandomQuote() string {
	if storage == nil {
		return ""
	}
	return storage.GetRandomQuote()
}
