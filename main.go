package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

func check_diff(files_path string, local_git_dir string, git_dir string) {
	files, files_err := exec.Command("git", "ls-files", files_path).CombinedOutput()
	if files_err != nil {
		fmt.Println(string(files))
		log.Fatal(files_err)
	}

	for _, byte_file := range files {
		file := string(byte_file)
		file_diff_byte, file_diff_byte_err := exec.Command("diff", "-u", "--color", local_git_dir+"/"+file, git_dir+"/"+file).CombinedOutput()
		if file_diff_byte_err != nil {
			fmt.Println(string(file_diff_byte))
			log.Fatal(file_diff_byte_err)
		}
		file_diff := string(file_diff_byte)

		if file_diff != "" {
			fmt.Printf("\nChanges in '%s':\n\n", file)
			fmt.Println(file_diff)
			fmt.Printf("\n\n")
			fmt.Println("Do you want to apply these changes to " + file + "? [y/N]\n")
			var REPLY string
			fmt.Scanln(&REPLY)
			if REPLY == "Y" || REPLY == "y" || os.Getenv("ACCEPT_ALL") == "true" {
				cp, cp_err := exec.Command("cp", git_dir+"/"+file, local_git_dir+"/"+file).CombinedOutput()
				if cp_err != nil {
					fmt.Println(string(cp))
					log.Fatal(cp_err)
				}
			}
		} else {
			fmt.Printf("No changes in '%s'.\n", file)
		}
	}
}

func remove_software(template_file_rel string, git_dir string) {
	template_file := git_dir + "/" + template_file_rel
	toolset_file := git_dir + "/images/ubuntu/toolsets/toolset-2204.json"
	// Add software here to remove
	remove_software_list := []string{
		"apache",
		"aws-tools",
		"gfortran",
		"java-tools",
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

	fmt.Println("Disabling software report generation...")

	software_gen_block := `  provisioner "shell" {
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
	template_file_contents, template_file_contents_err := ioutil.ReadFile(template_file)
	if template_file_contents_err != nil {
		log.Fatal(template_file_contents_err)
	}
	strings.Replace(software_gen_block, string(template_file_contents), "", -1)
	fmt.Println("Done.\n")

	fmt.Println("Removing software...")

	for _, software := range remove_software_list {
		fmt.Printf("    Removing install script for '%s'...\n", software)
		rm_rf_script, rm_rf_script_err := exec.Command("rm", "-f", git_dir+"/images/ubuntu/scripts/build/install-"+software+".sh").CombinedOutput()
		if rm_rf_script_err != nil {
			fmt.Println(string(rm_rf_script))
			log.Fatal(rm_rf_script_err)
		}

		fmt.Printf("    Removing line from Packer configuration for '%s'...\n", software)
		replace_script, replace_script_err := exec.Command("sed", "-i", "/install-"+software+".sh/d", template_file).CombinedOutput()
		if replace_script_err != nil {
			fmt.Println(string(replace_script))
			log.Fatal(replace_script_err)
		}

		// Special case for 'android-sdk' since toolset file uses 'android' instead
		if software == "android-sdk" {
			software = "android"
		}

		if strings.Contains(toolset_file, software) {
			fmt.Printf("    Removing configuration from '%s' for '%s'...\n\n", toolset_file, software)
			sed, sed_err := exec.Command("sed", "-i", "/    \""+software+"\":/,/    },/d", toolset_file).CombinedOutput()
			if sed_err != nil {
				fmt.Println(string(sed))
				log.Fatal(sed_err)
			}
		}
	}

	fmt.Println("Done.\n")
	validate_packer(template_file)
}

func add_software(template_file_rel string, local_dir string, git_dir string) {
	template_file := git_dir + "/" + template_file_rel
	// Add software here to add
	add_software_list := []string{
		"trivy",
	}

	fmt.Println("Adding software...")

	for _, software := range add_software_list {
		install_script := git_dir + "/images/ubuntu/scripts/build/install-" + software + ".sh"
		if _, stat_script_err := os.Stat(install_script); stat_script_err == nil {
			fmt.Printf("    Install script for '%s' already exists.\n", software)
			if _, cmp_scripts_err := exec.Command("cmp", "-s", local_dir+"/scripts/install-"+software+".sh", install_script).Output(); cmp_scripts_err == nil {
				fmt.Printf("    Install script for '%s' is up-to-date.\n", software)
			} else {
				fmt.Printf("    Install script for '%s' is outdated, will update it.\n", software)
				cp_script, cp_script_err := exec.Command("cp", local_dir+"/scripts/install-"+software+".sh", install_script).CombinedOutput()
				if cp_script_err != nil {
					fmt.Println(string(cp_script))
					log.Fatal(cp_script_err)
				}
			}
		} else {
			fmt.Printf("    Adding install script for '%s'...\n", software)
			cp_script, cp_script_err := exec.Command("cp", local_dir+"/scripts/install-"+software+".sh", install_script).CombinedOutput()
			if cp_script_err != nil {
				fmt.Println(string(cp_script))
				log.Fatal(cp_script_err)
			}
		}

		if strings.Contains(template_file, "install-"+software+".sh") {
			fmt.Printf("    Line for '%s' already exists in Packer configuration.", software)
		} else {
			fmt.Printf("    Adding line to Packer configuration for '%s'...", software)
			template_file_contents, template_file_contents_err := ioutil.ReadFile(template_file)
			if template_file_contents_err != nil {
				log.Fatal(template_file_contents_err)
			}

			template_file_lines := strings.Split(string(template_file_contents), "\n")
			for i, line := range template_file_lines {
				zstd_line := "\"${path.root}/../scripts/build/install-zstd.sh\""
				new_line := "\n      \"${path.root}/../scripts/build/install-" + software + ".sh\","

				// Account for zstd line not having a comma at the end
				if line == zstd_line {
					template_file_lines[i] = line + "," + new_line
				} else if line == zstd_line+"," {
					template_file_lines[i] = line + new_line
				}
			}

			template_file_output := strings.Join(template_file_lines, "\n")
			write_err := ioutil.WriteFile(template_file, []byte(template_file_output), 0644)
			if write_err != nil {
				log.Fatal(write_err)
			}
		}
	}

	fmt.Println("\nDone.\n")

	validate_packer(template_file)
}

func validate_packer(template_file string) {
	fmt.Println("Validating Packer configuration for '" + template_file + "'...")
	packer, packer_err := exec.Command("packer", "validate", "-var", "managed_image_resource_group_name=test", "-var", "location=westeurope", template_file).CombinedOutput()
	if packer_err != nil {
		fmt.Println(string(packer))
		log.Fatal(packer_err)
	}
	fmt.Println("Done.\n")
}

func apply_customizations(template_file string, local_dir string, git_dir string) {
	remove_software(template_file, git_dir)
	add_software(template_file, local_dir, git_dir)
}

func update(template_file string, local_dir string) {
	tmp_dir, tmp_dir_err := ioutil.TempDir(os.TempDir(), "")
	if tmp_dir_err != nil {
		log.Fatal(tmp_dir_err)
	}
	git_dir := tmp_dir + "/runner-images"
	defer os.RemoveAll(tmp_dir)

	fmt.Println("Cloning runner-images repository...")
	clone, clone_err := exec.Command("git", "clone", "git@github.com:actions/runner-images.git", "-q", git_dir).CombinedOutput()
	if clone_err != nil {
		fmt.Println(string(clone))
		log.Fatal(clone_err)
	}
	wd, wd_err := os.Getwd()
	if wd_err != nil {
		log.Fatal(wd_err)
	}
	chdir_err := os.Chdir(git_dir)
	if chdir_err != nil {
		log.Fatal(chdir_err)
	}
	fmt.Println("Done.\n")

	fmt.Println("Getting latest ubuntu22 release...")
	tags_list, tags_list_err := exec.Command("git", "tag").CombinedOutput()
	if tags_list_err != nil {
		fmt.Println(string(tags_list))
		log.Fatal(tags_list_err)
	}
	tags := strings.Split(string(tags_list), "\n")
	latest_tag := ""
	for _, tag := range tags {
		if strings.Contains(tag, "ubuntu22") {
			latest_tag = tag
		}
	}
	if latest_tag == "" {
		log.Fatal("No ubuntu22 tag found.")
	}
	fmt.Println("Done.\n")

	fmt.Println("Checking out latest tag '" + latest_tag + "'...")
	checkout, checkout_err := exec.Command("git", "checkout", latest_tag, "-q").CombinedOutput()
	if checkout_err != nil {
		fmt.Println(string(checkout))
		log.Fatal(checkout_err)
	}
	chdir_err = os.Chdir(wd)
	if chdir_err != nil {
		log.Fatal(chdir_err)
	}
	fmt.Println("Done.\n")

	apply_customizations(template_file, local_dir, git_dir)

	files_dirs := []string{
		"images/ubuntu",
		"helpers",
	}

	for _, file_dir := range files_dirs {
		local_git_dir := local_dir + "/runner-images"
		mkdir_err := exec.Command("mkdir", "-p", local_git_dir+"/"+file_dir).Run()
		if mkdir_err != nil {
			fmt.Println("Error: ", mkdir_err)
		}
		check_diff(file_dir, local_git_dir, git_dir)
	}

	validate_packer(local_dir + "/runner-images/" + template_file)
}

func main() {
	template_file := "images/ubuntu/templates/ubuntu-22.04.pkr.hcl"
	local_dir, local_dir_err := os.Getwd()
	if local_dir_err != nil {
		fmt.Println("Error: ", local_dir_err)
	}

	if len(os.Args) > 1 && os.Args[1] == "--apply" {
		git_dir := local_dir + "/runner-images"
		apply_customizations(template_file, local_dir, git_dir)
	} else {
		update(template_file, local_dir)
	}
}
