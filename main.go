package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"slices"
	"strings"
)

var removeSoftwareList = []string{
	"apache",
	"aws-tools",
	"gfortran",
	"php",
	"postgresql",
	"pulumi",
	"bazel",
	"rust",
	"julia",
	"selenium",
	"vcpkg",
	"android-sdk",
	"leiningen",
	"kotlin",
	"sbt",
	"oc-cli",
	"aliyun-cli",
	"rlang",
	"heroku",
}

var addSoftwareList = []string{
	"trivy",
	"github-runner",
}

func checkDiff(filesPath string, localGitDir string, gitDir string) {
	os.Chdir(gitDir)
	filesByte, err := exec.Command("git", "ls-files", filesPath).CombinedOutput()
	if err != nil {
		fmt.Println(string(filesByte))
		log.Fatal(err)
	}

	files := string(filesByte)

	if files == "" {
		fmt.Println("  No files found.")
		return
	}

	for _, byteFile := range strings.Split(files, "\n") {
		file := string(byteFile)

		if file == "" {
			return
		}

		localGitDirFile := localGitDir + "/" + file
		if _, err := os.Stat(localGitDirFile); errors.Is(err, os.ErrNotExist) {
			// Special case for install scripts
			installScriptShouldBeRemoved := slices.ContainsFunc(removeSoftwareList, func(s string) bool {
				return strings.HasSuffix(file, "scripts/build/install-"+s+".sh")
			})
			if !installScriptShouldBeRemoved {
				fmt.Printf("File '%s' is not in the list of software to be removed; do you want to add it? [y/N]\n", file)
				var REPLY string
				fmt.Scanln(&REPLY)
				if REPLY == "Y" || REPLY == "y" || os.Getenv("ACCEPT_ALL") == "true" {
					cp, err := exec.Command("cp", gitDir+"/"+file, localGitDirFile).CombinedOutput()
					if err != nil {
						fmt.Println(string(cp))
						log.Fatal(err)
					}
				}
			} else {
				fmt.Printf("  File '%s' does not exist in local git directory.\n", file)
			}
			continue
		}

		gitDirFile := gitDir + "/" + file
		if _, err := os.Stat(gitDirFile); errors.Is(err, os.ErrNotExist) {
			fmt.Printf("  File '%s' does not exist in git temp directory.\n", file)
			continue
		}

		// Ignore error, diff will return non-zero exit code if files are different.
		fileDiffByte, _ := exec.Command("diff", "-u", "--color", localGitDirFile, gitDirFile).CombinedOutput()
		fileDiff := string(fileDiffByte)

		if fileDiff != "" {
			fmt.Printf("\nChanges in '%s':\n\n", file)
			fmt.Println(fileDiff)
			fmt.Printf("\n\n")
			fmt.Println("Do you want to apply these changes to " + file + "? [y/N]\n")
			var REPLY string
			fmt.Scanln(&REPLY)
			if REPLY == "Y" || REPLY == "y" || os.Getenv("ACCEPT_ALL") == "true" {
				cp, err := exec.Command("cp", gitDirFile, localGitDirFile).CombinedOutput()
				if err != nil {
					fmt.Println(string(cp))
					log.Fatal(err)
				}
			}
		} else {
			fmt.Printf("  No changes in '%s'.\n", file)
		}
	}
}

func removeSoftware(templateFilRelative string, gitDir string) {
	templateFile := gitDir + "/" + templateFilRelative
	toolsetFile := gitDir + "/images/ubuntu/toolsets/toolset-2204.json"

	fmt.Println("Disabling software report generation...")

	softwareGenBlock := `  provisioner "shell" {
    environment_vars = ["IMAGE_VERSION=${var.image_version}", "INSTALLER_SCRIPT_FOLDER=${var.installer_script_folder}"]
    inline           = ["pwsh -File ${var.image_folder}/SoftwareReport/Generate-SoftwareReport.ps1 -OutputDirectory ${var.image_folder}", "pwsh -File ${var.image_folder}/tests/RunAll-Tests.ps1 -OutputDirectory ${var.image_folder}"]
  }

  provisioner "file" {
    destination = "${path.root}/../Ubuntu2204-Readme.md"
    direction   = "download"
    source      = "${var.image_folder}/software-report.md"
  }

  provisioner "file" {
    destination = "${path.root}/../software-report.json"
    direction   = "download"
    source      = "${var.image_folder}/software-report.json"
  }`

	templateFileContentsBytes, err := os.ReadFile(templateFile)
	if err != nil {
		log.Fatal(err)
	}
	templateFileContents := string(templateFileContentsBytes)

	templateFileNewContents := strings.Replace(templateFileContents, softwareGenBlock, "", 1)
	err_ := os.WriteFile(templateFile, []byte(templateFileNewContents), 0644)
	if err_ != nil {
		log.Fatal(err_)
	}
	fmt.Printf("Done.\n\n")

	fmt.Println("Removing software...")

	for _, software := range removeSoftwareList {
		fmt.Printf("    Removing install script for '%s'...\n", software)
		rmRfScript, err := exec.Command("rm", "-f", gitDir+"/images/ubuntu/scripts/build/install-"+software+".sh").CombinedOutput()
		if err != nil {
			fmt.Println(string(rmRfScript))
			log.Fatal(err)
		}

		fmt.Printf("    Removing line from Packer configuration for '%s'...\n", software)
		replaceScript, err := exec.Command("sed", "-i", "/install-"+software+".sh/d", templateFile).CombinedOutput()
		if err != nil {
			fmt.Println(string(replaceScript))
			log.Fatal(err)
		}

		// Special case for 'android-sdk' since toolset file uses 'android' instead
		if software == "android-sdk" {
			software = "android"
		}

		toolsetFileContentsBytes, err := os.ReadFile(toolsetFile)
		if err != nil {
			log.Fatal(err)
		}
		toolsetFileContents := string(toolsetFileContentsBytes)

		if strings.Contains(toolsetFileContents, software) {
			fmt.Printf("    Removing configuration from '%s' for '%s'...\n\n", toolsetFile, software)
			sed, err := exec.Command("sed", "-i", "/    \""+software+"\":/,/    },/d", toolsetFile).CombinedOutput()
			if err != nil {
				fmt.Println(string(sed))
				log.Fatal(err)
			}
		}
	}

	fmt.Printf("Done.\n\n")
}

func addSoftware(templateFileRelative string, localDir string, gitDir string) {
	templateFile := gitDir + "/" + templateFileRelative

	fmt.Println("Adding software...")

	for _, software := range addSoftwareList {
		err := copyScript(localDir, gitDir, "install-"+software+".sh")
		if err != nil {
			log.Fatal(err)
		}

		templateFileContentsBytes, err := os.ReadFile(templateFile)
		if err != nil {
			log.Fatal(err)
		}
		templateFileContents := string(templateFileContentsBytes)

		if strings.Contains(templateFileContents, "install-"+software+".sh") {
			fmt.Printf("    Line for '%s' already exists in Packer configuration.", software)
		} else {
			fmt.Printf("    Adding line to Packer configuration for '%s'...", software)

			templateFileLines := strings.Split(templateFileContents, "\n")
			foundLine := false
			for i, line := range templateFileLines {
				zstdLine := "      \"${path.root}/../scripts/build/install-zstd.sh\""
				newLine := "\n      \"${path.root}/../scripts/build/install-" + software + ".sh\","

				// Account for zstd line not having a comma at the end
				if line == zstdLine {
					templateFileLines[i] = line + "," + newLine
					foundLine = true
				} else if line == zstdLine+"," {
					templateFileLines[i] = line + newLine
					foundLine = true
				}
			}

			if !foundLine {
				log.Fatal("Could not find line to insert new software.")
			}

			templateFileOutput := strings.Join(templateFileLines, "\n")
			err_ := os.WriteFile(templateFile, []byte(templateFileOutput), 0644)
			if err_ != nil {
				log.Fatal(err_)
			}
		}
	}

	fmt.Printf("\nDone.\n\n")
}

func copyScript(localDir string, gitDir string, scriptName string) error {
	script := gitDir + "/images/ubuntu/scripts/build/" + scriptName
	localScript := localDir + "/scripts/" + scriptName

	if _, err := os.Stat(script); err == nil {
		fmt.Printf("    Install script '%s' already exists.\n", scriptName)
		if _, err := exec.Command("cmp", "-s", localScript, script).Output(); err == nil {
			fmt.Printf("    Install script '%s' is up-to-date.\n", scriptName)
		} else {
			fmt.Printf("    Install script '%s' is outdated, will update it.\n", scriptName)
			cpScript, err := exec.Command("cp", localScript, script).CombinedOutput()
			if err != nil {
				return fmt.Errorf(
					"Error copying install script '%s' to '%s': %s",
					localScript,
					script,
					cpScript,
				)
			}
		}
	} else {
		fmt.Printf("    Adding install script '%s'...\n", script)
		cpScript, err := exec.Command("cp", localScript, script).CombinedOutput()
		if err != nil {
			return fmt.Errorf(
				"Error copying install script '%s' to '%s': %s",
				localScript,
				script,
				cpScript,
			)
		}
	}

	return nil
}

func validatePacker(templateFile string) {
	fmt.Println("Validating Packer configuration for '" + templateFile + "'...")
	packerInit, err := exec.Command("packer", "init", templateFile).CombinedOutput()
	if err != nil {
		fmt.Println(string(packerInit))
		log.Fatal(err)
	}

	packerValidate, err := exec.Command("packer", "validate", "-var", "managed_image_resource_group_name=test", "-var", "location=westeurope", templateFile).CombinedOutput()
	if err != nil {
		fmt.Println(string(packerValidate))
		log.Fatal(err)
	}
	fmt.Printf("Done.\n\n")
}

func applyCustomizations(templateFile string, localDir string, gitDir string) {
	removeSoftware(templateFile, gitDir)
	addSoftware(templateFile, localDir, gitDir)
}

func update(templateFile string, localDir string) {
	tmpDir, err := os.MkdirTemp(os.TempDir(), "")
	if err != nil {
		log.Fatal(err)
	}

	gitDir := tmpDir + "/runner-images"
	defer os.RemoveAll(tmpDir)

	fmt.Println("Cloning runner-images repository...")
	clone, err := exec.Command("git", "clone", "https://github.com/actions/runner-images.git", "-q", gitDir).CombinedOutput()
	if err != nil {
		fmt.Println(string(clone))
		log.Fatal(err)
	}
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	err_ := os.Chdir(gitDir)
	if err_ != nil {
		log.Fatal(err_)
	}
	fmt.Printf("Done.\n\n")

	fmt.Println("Getting latest ubuntu22 release...")
	tagsList, err := exec.Command("git", "tag").CombinedOutput()
	if err != nil {
		fmt.Println(string(tagsList))
		log.Fatal(err)
	}
	tags := strings.Split(string(tagsList), "\n")
	latestTag := ""
	for _, tag := range tags {
		if strings.Contains(tag, "ubuntu22") {
			latestTag = tag
		}
	}
	if latestTag == "" {
		log.Fatal("No ubuntu22 tag found.")
	}
	fmt.Printf("Done.\n\n")

	fmt.Println("Checking out latest tag '" + latestTag + "'...")
	checkout, err := exec.Command("git", "checkout", latestTag, "-q").CombinedOutput()
	if err != nil {
		fmt.Println(string(checkout))
		log.Fatal(err)
	}
	err = os.Chdir(wd)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Done.\n\n")

	applyCustomizations(templateFile, localDir, gitDir)

	filesDirs := []string{
		"images/ubuntu",
		"helpers",
	}

	for _, fileDir := range filesDirs {
		localGitDir := localDir + "/runner-images"
		mkdir, err := exec.Command("mkdir", "-p", localGitDir+"/"+fileDir).CombinedOutput()
		if err != nil {
			fmt.Println(string(mkdir))
			log.Fatal(err)
		}
		fmt.Printf("Checking differences in '%s'...\n", fileDir)
		checkDiff(fileDir, localGitDir, gitDir)
		fmt.Printf("Done.\n\n")
	}

	validatePacker(localDir + "/runner-images/" + templateFile)
}

func main() {
	templateFile := "images/ubuntu/templates/ubuntu-22.04.pkr.hcl"
	localDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error: ", err)
	}

	if len(os.Args) > 1 && os.Args[1] == "--apply" {
		gitDir := localDir + "/runner-images"
		applyCustomizations(templateFile, localDir, gitDir)
		validatePacker(localDir + "/runner-images/" + templateFile)
	} else {
		update(templateFile, localDir)
	}
}
