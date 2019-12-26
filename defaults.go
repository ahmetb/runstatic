package main

var (
	defaultProjectFile = mustDotDir().File("project_id")
)

func readConfig(file dotFile) (string, bool, error) {
	ok, err := file.Exists()
	if err != nil {
		return "", false, err
	}
	if !ok {
		return "", false, nil
	}
	v, err := file.Read()
	return v, true, err
}

func writeConfig(file dotFile, value string) error { return file.Write(value) }

func defaultProject() (string, bool, error) { return readConfig(defaultProjectFile) }
func setDefaultProject(p string) error      { return writeConfig(defaultProjectFile, p) }
