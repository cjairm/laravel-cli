package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func args(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("requires at least one arg")
	}
	maybeDockerLower := strings.ToLower(args[0])
	if !strings.Contains(maybeDockerLower, "docker") {
		return errors.New("Docker is the only template available at this moment")
	}
	return nil
}

func run(cmd *cobra.Command, args []string) {
	if err := existDir(Dir); err != nil {
		fmt.Println("Error:", err)
		return
	}

	if err := isEmptyDir(Dir); err != nil {
		fmt.Println("Error:", err)
		return
	}

	generateDockerComposeFile()
	generateNginxFiles()
	generatePhpFiles()
	generateReadmeFile()

	composerFile := fmt.Sprintf("%s/docker-compose.yml", Dir)
	dockerCreateProjectCmd := exec.Command(
		"docker-compose",
		"-f",
		composerFile,
		"run",
		"--rm",
		"composer",
		"create-project",
		"laravel/laravel:^11.0",
		".",
	)
	dockerCreateProjectCmd.Dir = Dir
	executeBashCmd(*dockerCreateProjectCmd)

	artisanMigrateCmd := exec.Command(
		"docker-compose",
		"-f",
		composerFile,
		"run",
		"--rm",
		"artisan",
		"migrate",
	)
	artisanMigrateCmd.Dir = Dir
	executeBashCmd(*artisanMigrateCmd)

	updateLaravelEnvFile()
}

var (
	AppName   string
	AppPort   int
	Dir       string
	MysqlPass string
)

var createCmd = &cobra.Command{
	Use:   "create [template type]",
	Args:  args,
	Short: "Basic template using laravel",
	Long:  "...",
	Run:   run,
}

func init() {
	rootCmd.AddCommand(createCmd)

	currentDir, err := filepath.Abs("./")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	createCmd.Flags().
		StringVarP(&AppName, "appName", "n", "", "Name of your app. If not provided the name of parent folder will be taken")

	createCmd.Flags().
		IntVarP(&AppPort, "appPort", "p", 8000, "App port number. If not provided the port 8000 will be used")

	createCmd.Flags().StringVarP(
		&Dir,
		"dir",
		"d",
		currentDir,
		"Directory where template is saved (required)",
	)

	err = createCmd.MarkFlagRequired("dir")
	if err != nil {
		fmt.Println("Unexpected error (--dir)")
		os.Exit(1)
	}
}

func existDir(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

func isEmptyDir(path string) error {
	files, err := os.ReadDir(path)
	if err != nil {
		return err
	} else if len(files) > 0 {
		return errors.New("Please provide an empty dir")
	}
	return nil
}

func copyFileToDir(path, dstDir string) (int64, error) {
	srcFile, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dstDir)
	if err != nil {
		return 0, err
	}
	defer dstFile.Close()

	return io.Copy(dstFile, srcFile)
}

func generateNewFile(fileName string) (string, error) {
	fmt.Printf("\nGenerating %v\n", fileName)

	absPath, err := filepath.Abs(
		fmt.Sprintf("./templates/%s", fileName),
	)
	if err != nil {
		return "", err
	}

	dstFile := fmt.Sprintf("%s/%s", Dir, fileName)
	copiedBytes, err := copyFileToDir(absPath, dstFile)
	if err != nil {
		return "", err
	}

	fmt.Printf("%v bytes stored\n\n", copiedBytes)
	return dstFile, nil
}

func removeNonAlphanumeric(s string) string {
	return strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' || r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' {
			return r
		}
		return '_'
	}, s)
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func replaceAllInFile(filePath string, valuesToReplace [][2]string) error {
	fileOpened, err := os.Open(filePath)
	if err != nil {
		return err
	}

	newFileContent := ""
	scanner := bufio.NewScanner(fileOpened)

	for scanner.Scan() {
		line := scanner.Text()
		for _, pairOfValues := range valuesToReplace {
			newValue := pairOfValues[1]
			if pairOfValues[0] == "{{APP_NAME}}" {
				newValue = strings.ToLower(pairOfValues[1])
			}
			line = strings.ReplaceAll(line, pairOfValues[0], newValue)
		}
		newFileContent += fmt.Sprintf("%s\n", line)
	}

	fileOpened.Close()
	if err := scanner.Err(); err != nil {
		return err
	}

	err = os.WriteFile(filePath, []byte(newFileContent), 0644)
	if err != nil {
		return err
	}

	return nil
}

func updateLaravelEnvFile() {
	dstFile := fmt.Sprintf("%s/%s/.env", Dir, AppName)
	valuesToReplace := make([][2]string, 0)

	valuesToReplace = append(
		valuesToReplace,
		[2]string{"APP_NAME=Laravel", fmt.Sprintf("APP_NAME=\"%s\"", AppName)},
	)
	valuesToReplace = append(
		valuesToReplace,
		[2]string{
			"APP_URL=http://localhost",
			fmt.Sprintf("APP_URL=http://localhost:%d", AppPort),
		},
	)
	valuesToReplace = append(
		valuesToReplace,
		[2]string{"DB_CONNECTION=sqlite", "DB_CONNECTION=mysql"},
	)
	valuesToReplace = append(
		valuesToReplace,
		[2]string{"# DB_HOST=127.0.0.1", "DB_HOST=db"},
	)
	valuesToReplace = append(
		valuesToReplace,
		[2]string{"# DB_PORT=3306", "DB_PORT=3306"},
	)
	valuesToReplace = append(
		valuesToReplace,
		[2]string{"# DB_DATABASE=laravel", "DB_DATABASE=db_dev"},
	)
	valuesToReplace = append(
		valuesToReplace,
		[2]string{"# DB_USERNAME=root", "DB_USERNAME=user_dev"},
	)
	valuesToReplace = append(
		valuesToReplace,
		[2]string{"# DB_PASSWORD=", fmt.Sprintf("DB_PASSWORD=%s", MysqlPass)},
	)
	fmt.Println("Updating env vars...")
	if err := replaceAllInFile(dstFile, valuesToReplace); err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println(".env updated")
}

func generateDockerComposeFile() {
	dstFile, err := generateNewFile("docker-compose.yml")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	valuesToReplace := make([][2]string, 0)

	if AppName != "" {
		AppName = filepath.Base(filepath.Dir(dstFile))
	}
	valuesToReplace = append(
		valuesToReplace,
		[2]string{"{{APP_NAME}}", removeNonAlphanumeric(AppName)},
	)

	valuesToReplace = append(
		valuesToReplace,
		[2]string{"{{APP_PORT}}", strconv.Itoa(AppPort)},
	)

	MysqlPass = generateRandomString(20)
	fmt.Printf("MYSQL password: %v\n", MysqlPass)
	fmt.Println("MYSQL database: db_dev")
	fmt.Println("MYSQL user: user_dev")
	fmt.Printf("MYSQL port: 3306\n\n")
	valuesToReplace = append(
		valuesToReplace,
		[2]string{"{{MYSQL_PASSWORD}}", MysqlPass},
	)

	fmt.Println("Replacing placeholders...")
	if err := replaceAllInFile(dstFile, valuesToReplace); err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Placeholders replaced in docker-composer.yml")
}

func generateNginxFiles() {
	_, err := generateNewFile("nginx.dockerfile")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	dstDir := fmt.Sprintf("%s/nginx", Dir)
	err = os.MkdirAll(dstDir, 0755)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	_, err = generateNewFile("nginx/default.conf")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
}

func generatePhpFiles() {
	_, err := generateNewFile("php.dockerfile")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	dstDir := fmt.Sprintf("%s/php", Dir)
	err = os.MkdirAll(dstDir, 0755)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	_, err = generateNewFile("php/www.conf")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
}

func generateReadmeFile() {
	_, err := generateNewFile("README.md")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
}

func executeBashCmd(bashCmd exec.Cmd) {
	stdout, err := bashCmd.StdoutPipe()
	if err != nil {
		fmt.Println(err)
		return
	}

	stderr, err := bashCmd.StderrPipe()
	if err != nil {
		fmt.Println(err)
		return
	}

	if err := bashCmd.Start(); err != nil {
		fmt.Println(err)
		return
	}

	scannerOut := bufio.NewScanner(stdout)
	for scannerOut.Scan() {
		fmt.Println(scannerOut.Text())
	}

	scannerErr := bufio.NewScanner(stderr)
	for scannerErr.Scan() {
		fmt.Println(scannerErr.Text())
	}

	if err := bashCmd.Wait(); err != nil {
		fmt.Println(err)
	}
}
