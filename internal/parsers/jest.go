package parsers

type JestParser struct {
	filePath string
}

func NewJestParser(filePath string) *JestParser {
	return &JestParser{filePath: filePath}
}

func (j *JestParser) Parse() ([]TestFailure, error) {
	return nil, nil
}

func (j *JestParser) RelevantFiles() []string {
	return []string{"*.js", "*.ts", "*.jsx", "*.tsx", "**/__tests__/*"}
}
