package git

import (
	"bytes"
	"encoding/json"
	"github.com/devtron-labs/git-sensor/internals/logger"
	"github.com/devtron-labs/git-sensor/internals/sql"
	"github.com/stretchr/testify/assert"
	"io"
	"io/fs"
	"log"
	"os"
	"path"
	"sort"
	"testing"
	"time"
)

func TestCreateGitMaterialBackupBeforeUpdate(t *testing.T) {
	t.Run("Success_With_Non_Existing_Dir", testSuccessWithNonExistingDir)
	t.Run("Success_With_Existing_Dir", testSuccessWithExistingDir)
	t.Run("Source_Dir_Error", testSourceDirError)
}

func testSuccessWithNonExistingDir(t *testing.T) {
	sugaredLogger := logger.NewSugaredLogger()
	impl := &RepositoryManagerImpl{
		logger: sugaredLogger,
	}
	srcBaseDir := impl.getBaseDirForMaterial(validMaterialDetails)
	srcDir := path.Join(srcBaseDir, "github.com/Ash-exp/test.git/.git")
	err := os.MkdirAll(srcDir, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(srcBaseDir)
	fileContent := []byte("file content\nfound\n")
	err = os.WriteFile(path.Join(srcDir, "test.txt"), fileContent, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	backupBaseDir := impl.getBackupDirForMaterial(validMaterialDetails)
	if _, fileErr := os.Stat(backupBaseDir); fileErr == nil {
		err := os.RemoveAll(backupBaseDir)
		if err != nil {
			t.Fatal(err)
		}
	}
	dstDir := path.Join(backupBaseDir, BACKUP_GIT_BASE_SUB_DIR)
	err = os.MkdirAll(dstDir, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(backupBaseDir)
	err = impl.BackupGitMaterialBeforeUpdate(validMaterialDetails)
	assert.NoError(t, err)
	assert.True(t, deepCompare(path.Join(srcDir, "test.txt"), path.Join(dstDir, "github.com/Ash-exp/test.git/.git", "test.txt")))
	file, err := os.ReadFile(path.Join(backupBaseDir, BACKUP_DB_MODELS_FILE_NAME))
	if err != nil {
		t.Fatal(err)
	}
	gotMaterialDetails := sql.GitMaterial{}
	err = json.Unmarshal(file, &gotMaterialDetails)
	if err != nil {
		t.Fatal(err)
	}
	sort.Slice(validMaterialDetails.CiPipelineMaterials, func(i, j int) bool {
		return validMaterialDetails.CiPipelineMaterials[i].Id < validMaterialDetails.CiPipelineMaterials[j].Id
	})
	sort.Slice(gotMaterialDetails.CiPipelineMaterials, func(i, j int) bool {
		return gotMaterialDetails.CiPipelineMaterials[i].Id < gotMaterialDetails.CiPipelineMaterials[j].Id
	})
	assert.EqualExportedValues(t, validMaterialDetails, &gotMaterialDetails)
}

func testSuccessWithExistingDir(t *testing.T) {
	sugaredLogger := logger.NewSugaredLogger()
	impl := &RepositoryManagerImpl{
		logger: sugaredLogger,
	}
	srcBaseDir := impl.getBaseDirForMaterial(validMaterialDetails)
	srcDir := path.Join(srcBaseDir, "github.com/Ash-exp/test.git/.git")
	err := os.MkdirAll(srcDir, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(srcBaseDir)
	fileContent := []byte("file content\nfound\n")
	err = os.WriteFile(path.Join(srcDir, "test.txt"), fileContent, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	backupBaseDir := impl.getBackupDirForMaterial(validMaterialDetails)
	if _, fileErr := os.Stat(backupBaseDir); fileErr == nil {
		err := os.RemoveAll(backupBaseDir)
		if err != nil {
			t.Fatal(err)
		}
	}
	dstDir := path.Join(backupBaseDir, BACKUP_GIT_BASE_SUB_DIR, "github.com/Ash-exp/test.git/.git")
	err = os.MkdirAll(dstDir, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(backupBaseDir)
	tmpFileContent := []byte("invalid file content\nfound\n")
	err = os.WriteFile(path.Join(dstDir, "test.txt"), tmpFileContent, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	err = impl.BackupGitMaterialBeforeUpdate(validMaterialDetails)
	assert.NoError(t, err)
	assert.True(t, deepCompare(path.Join(srcDir, "test.txt"), path.Join(dstDir, "test.txt")))
	file, err := os.ReadFile(path.Join(backupBaseDir, BACKUP_DB_MODELS_FILE_NAME))
	if err != nil {
		t.Fatal(err)
	}
	gotMaterialDetails := sql.GitMaterial{}
	err = json.Unmarshal(file, &gotMaterialDetails)
	if err != nil {
		t.Fatal(err)
	}
	sort.Slice(validMaterialDetails.CiPipelineMaterials, func(i, j int) bool {
		return validMaterialDetails.CiPipelineMaterials[i].Id < validMaterialDetails.CiPipelineMaterials[j].Id
	})
	sort.Slice(gotMaterialDetails.CiPipelineMaterials, func(i, j int) bool {
		return gotMaterialDetails.CiPipelineMaterials[i].Id < gotMaterialDetails.CiPipelineMaterials[j].Id
	})
	assert.EqualExportedValues(t, validMaterialDetails, &gotMaterialDetails)
}

func testSourceDirError(t *testing.T) {
	sugaredLogger := logger.NewSugaredLogger()
	impl := &RepositoryManagerImpl{
		logger: sugaredLogger,
	}
	srcBaseDir := impl.getBaseDirForMaterial(validMaterialDetails)
	_ = os.RemoveAll(srcBaseDir)
	err := impl.BackupGitMaterialBeforeUpdate(validMaterialDetails)
	assert.Error(t, err)
	assert.ErrorIs(t, err, fs.ErrNotExist)
}

const chunkSize = 64000

func deepCompare(file1, file2 string) bool {
	f1s, err := os.Stat(file1)
	if err != nil {
		log.Fatal(err)
	}
	f2s, err := os.Stat(file2)
	if err != nil {
		log.Fatal(err)
	}

	if f1s.IsDir() || f2s.IsDir() {
		return false
	}

	if f1s.Mode() != f2s.Mode() {
		return false
	}

	if f1s.Size() != f2s.Size() {
		return false
	}

	f1, err := os.Open(file1)
	if err != nil {
		log.Fatal(err)
	}

	f2, err := os.Open(file2)
	if err != nil {
		log.Fatal(err)
	}

	for {
		b1 := make([]byte, chunkSize)
		_, err1 := f1.Read(b1)

		b2 := make([]byte, chunkSize)
		_, err2 := f2.Read(b2)

		if err1 != nil || err2 != nil {
			if err1 == io.EOF && err2 == io.EOF {
				return true
			} else if err1 == io.EOF && err2 == io.EOF {
				return false
			} else {
				log.Fatal(err1, err2)
			}
		}

		if !bytes.Equal(b1, b2) {
			return false
		}
	}
}

var validMaterialDetails = &sql.GitMaterial{
	Id:                  1,
	GitProviderId:       2,
	Url:                 "https://github.com/Ash-exp/test.git",
	FetchSubmodules:     false,
	Name:                "2-test",
	CheckoutLocation:    "/git-base/1/github.com/Ash-exp/test.git",
	CheckoutStatus:      true,
	CheckoutMsgAny:      "Please check if repo url is correct and you have proper access to it",
	Deleted:             false,
	LastFetchTime:       time.Now(),
	FetchStatus:         true,
	LastFetchErrorCount: 0,
	FetchErrorMessage:   "",
	CloningMode:         "",
	CreateBackup:        true,
	FilterPattern:       nil,
	GitProvider:         nil,
	CiPipelineMaterials: []*sql.CiPipelineMaterial{
		{
			Id:            1,
			GitMaterialId: 1,
			Type:          "SOURCE_TYPE_BRANCH_FIXED",
			Value:         "main",
			Active:        true,
			LastSeenHash:  "08b31ea791170b568e30ea273eda510e1d1216c6",
			CommitAuthor:  "Asutosh Das \u003casutosh2000ad@gmail.com\u003e",
			CommitDate:    time.Now(),
			CommitMessage: "Create README.md",
			CommitHistory: "[{\"Commit\":\"08b31ea791170b568e30ea273eda510e1d1216c6\",\"Author\":\"Asutosh Das \\u003casutosh2000ad@gmail.com\\u003e\",\"Date\":\"2025-03-20T22:04:58+05:30\",\"Message\":\"Create README.md\",\"webhookData\":null},{\"Commit\":\"0fd5e18414e6247b932d377f99a75dab65a01790\",\"Author\":\"Badal Kumar \\u003c130441461+badal773@users.noreply.github.com\\u003e\",\"Date\":\"2024-04-19T02:15:50+05:30\",\"Message\":\"Delete img/.DS_Store\",\"webhookData\":null},{\"Commit\":\"bbc43bf38e36cb9a98bf000a5e840de320bca12c\",\"Author\":\"Badal Kumar \\u003c130441461+badal773@users.noreply.github.com\\u003e\",\"Date\":\"2024-04-19T02:01:49+05:30\",\"Message\":\"minimized Dockerfile\",\"webhookData\":null},{\"Commit\":\"062949e1e97871608588b6c15947310e4523b07b\",\"Author\":\"Badal Kumar \\u003c130441461+badal773@users.noreply.github.com\\u003e\",\"Date\":\"2024-04-19T01:47:17+05:30\",\"Message\":\"Update Dockerfile\",\"webhookData\":null},{\"Commit\":\"3f1317c4d696b1322a1531288e271fc0cc900e2f\",\"Author\":\"Badal Kumar Prusty \\u003cbadalkumar@Badals-MacBook-Pro.local\\u003e\",\"Date\":\"2024-04-19T01:39:19+05:30\",\"Message\":\"Test version-2.0\",\"webhookData\":null}]",
			Errored:       false,
			ErrorMsg:      "",
		},
		{
			Id:            2,
			GitMaterialId: 1,
			Type:          "SOURCE_TYPE_BRANCH_FIXED",
			Value:         "main",
			Active:        true,
			LastSeenHash:  "08b31ea791170b568e30ea273eda510e1d1216c6",
			CommitAuthor:  "Asutosh Das \u003casutosh2000ad@gmail.com\u003e",
			CommitDate:    time.Now(),
			CommitMessage: "Create README.md",
			CommitHistory: "[{\"Commit\":\"08b31ea791170b568e30ea273eda510e1d1216c6\",\"Author\":\"Asutosh Das \\u003casutosh2000ad@gmail.com\\u003e\",\"Date\":\"2025-03-20T22:04:58+05:30\",\"Message\":\"Create README.md\",\"webhookData\":null},{\"Commit\":\"0fd5e18414e6247b932d377f99a75dab65a01790\",\"Author\":\"Badal Kumar \\u003c130441461+badal773@users.noreply.github.com\\u003e\",\"Date\":\"2024-04-19T02:15:50+05:30\",\"Message\":\"Delete img/.DS_Store\",\"webhookData\":null},{\"Commit\":\"bbc43bf38e36cb9a98bf000a5e840de320bca12c\",\"Author\":\"Badal Kumar \\u003c130441461+badal773@users.noreply.github.com\\u003e\",\"Date\":\"2024-04-19T02:01:49+05:30\",\"Message\":\"minimized Dockerfile\",\"webhookData\":null},{\"Commit\":\"062949e1e97871608588b6c15947310e4523b07b\",\"Author\":\"Badal Kumar \\u003c130441461+badal773@users.noreply.github.com\\u003e\",\"Date\":\"2024-04-19T01:47:17+05:30\",\"Message\":\"Update Dockerfile\",\"webhookData\":null},{\"Commit\":\"3f1317c4d696b1322a1531288e271fc0cc900e2f\",\"Author\":\"Badal Kumar Prusty \\u003cbadalkumar@Badals-MacBook-Pro.local\\u003e\",\"Date\":\"2024-04-19T01:39:19+05:30\",\"Message\":\"Test version-2.0\",\"webhookData\":null}]",
			Errored:       false,
			ErrorMsg:      "",
		},
	},
}
