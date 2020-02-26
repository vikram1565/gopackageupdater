/*
packageUpdator is used to update the external packages from given directory.
This file collect all external packages and update these external packages.
*/

package gopackageupdater

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"strings"
)

// Package - struct
type updatePKG struct {
	directoryPath string
}

// New - New
func New(dirPath string) *updatePKG {
	if dirPath == "" {
		log.Fatal("Directory path is empty.")
	}
	if _, err := os.Stat(dirPath); err == nil {
		return &updatePKG{
			directoryPath: dirPath,
		}
	}
	log.Fatal(dirPath + " is not a directory")
	return &updatePKG{}
}

// Start - Start
func (p *updatePKG) Start() {

	log.Println("Starting To Update Packages")
	// check given directory is empty or not

	stdPkgData, getError := getStandardPackages(p.directoryPath)
	if getError != nil {
		log.Fatal(getError)
	}
	pkgArray, getError := getExternalPackages(p)
	if getError != nil {
		log.Fatal(getError)
	}
	updateError := updatePackages(pkgArray, stdPkgData, p.directoryPath)
	if updateError != nil {
		log.Fatal(updateError)
	}
	generateError := generateExe(p.directoryPath)
	if generateError != nil {
		log.Fatal(generateError)
	}
	log.Println("Successfully Updated Packages")
}

// getStandardPackages - getStandardPackages
func getStandardPackages(dirPath string) (string, error) {
	// list all standard packages
	cmd := exec.Command("go", "list", "std")
	cmd.Dir = dirPath
	stdPkg := bytes.Buffer{}
	cmd.Stdout = &stdPkg
	cmdError := cmd.Run()
	if cmdError != nil {
		log.Fatal("Failed to get standard packages :", cmdError)
		return "", cmdError
	}

	stdPkgData := stdPkg.String()
	stdPkgData = strings.Replace(stdPkgData, "[", "", -1)
	stdPkgData = strings.Replace(stdPkgData, "]", "", -1)
	return stdPkgData, nil
}

// getExternalPackages - getExternalPackages
func getExternalPackages(p *updatePKG) ([]string, error) {
	// list all imported packages
	var pkgArray []string
	cmd := exec.Command("go", "list", "-f", "{{.Imports}}", "./...")
	cmd.Dir = p.directoryPath
	importPkgData := bytes.Buffer{}
	er := bytes.Buffer{}
	cmd.Stdout = &importPkgData
	cmd.Stderr = &er
	cmdError := cmd.Run()
	if cmdError != nil {
		log.Fatal("Failed to get import packages :", cmdError, er.String())
		return pkgArray, cmdError
	}
	// removed extra symbols from data
	pkgData := importPkgData.String()
	pkgData = strings.Replace(pkgData, "[", "", -1)
	pkgData = strings.Replace(pkgData, "]", "", -1)
	pkgData = strings.Replace(pkgData, " ", "\n", -1)
	pkgArray = strings.Split(pkgData, "\n")
	return pkgArray, nil
}

// updatePackages - updatePackages
func updatePackages(pkgArray []string, stdPkgData, dirPath string) error {
	// get user defined package name
	dir := strings.Split(dirPath, "/")
	dirStr := ""
	if len(dir) > 0 {
		dirStr = dir[len(dir)-1]
	}
	m := make(map[string]interface{})
	for _, d := range pkgArray {
		// check standard packages and local packages
		if d == "" || strings.HasPrefix(d, dirStr) || strings.Contains(stdPkgData, d) {
			continue
		}
		if _, ok := m[d]; !ok {
			// update the external packages
			m[d] = d
			cmd := exec.Command("go", "get", "-u", d)
			cmd.Dir = dirPath
			cmdError := cmd.Run()
			if cmdError != nil {
				log.Print("Fail to update "+d+" package : ", cmdError)
			} else {
				log.Println("Package update success: ", d)
			}
		}
	}
	return nil
}

// generateExe - generateExe
func generateExe(dirPath string) error {
	// generate the executable
	buf := bytes.Buffer{}
	cmd := exec.Command("go", "build")
	cmd.Stderr = &buf
	cmd.Dir = dirPath
	cmdError := cmd.Run()
	if cmdError != nil {
		log.Fatal("Failed to generate executable: ", cmdError, buf.String())
		return cmdError
	}
	log.Println("Executable is generated at : " + dirPath)
	return nil
}
