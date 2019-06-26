//go:generate go-bindata -nometadata -ignore=\.go$ -pkg resources -o resources.go ./...

package resources

import "fmt"

// Get gets the template with the specified name
func Get(name string) string {
	templateBytes, err := Asset(name)
	if err != nil {
		panic(fmt.Sprintf("Cannot load resource \"%s\": %v", name, err))
	}

	return string(templateBytes)
}

