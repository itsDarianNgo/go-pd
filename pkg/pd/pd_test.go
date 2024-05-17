package pd_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/itsDarianNgo/go-pd/pkg/pd"
	"github.com/itsDarianNgo/go-pd/pkg/pd/utils"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

const SkipIntegrationTest = "skipping integration test"

var fileIDPost string
var fileIDPut string
var listID string
var testHashFilePath = "test_hashes.csv"

// SetupTestEnvironment cleans up the test environment before running tests
func SetupTestEnvironment() {
	err := os.Setenv("ENV_MODE", "test") // Set environment mode to test
	if err != nil {
		fmt.Printf("Error setting environment variable: %v\n", err)
	}
	// Remove the existing test hashes file to ensure a clean test environment
	testHashFilePath := utils.GetHashFilePath()
	if err := os.Remove(testHashFilePath); err != nil && !os.IsNotExist(err) {
		fmt.Printf("Error removing test hash file: %v\n", err)
	}
}

// CleanupTestEnvironment cleans up the test environment after running tests
func CleanupTestEnvironment() {
	testHashFilePath := utils.GetHashFilePath()
	if err := os.Remove(testHashFilePath); err != nil && !os.IsNotExist(err) {
		fmt.Printf("Error removing test hash file: %v\n", err)
	}
}

// TestMain sets up and tears down the test environment
func TestMain(m *testing.M) {
	SetupTestEnvironment()
	code := m.Run() // Run the tests
	CleanupTestEnvironment()
	os.Exit(code)
} // TestPD_UploadPOST is a unit test for // the POST upload method

func TestPD_UploadPOST(t *testing.T) {
	SetupTestEnvironment()

	server := pd.MockFileUploadServer()
	defer server.Close()
	testURL := server.URL + "/file"

	// Define the hash file path
	hashFilePath := "test_hashes.csv"

	// Initialize hash file
	if err := utils.InitializeHashFile(hashFilePath); err != nil {
		t.Fatalf("Failed to initialize hash file: %v", err)
	}

	req := &pd.RequestUpload{
		PathToFile: "testdata/cat.jpg",
		FileName:   "test_post_cat.jpg",
		Anonymous:  true,
		URL:        testURL,
	}

	c := pd.New(nil, nil)
	rsp, err := c.UploadPOST(req, hashFilePath)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, 201, rsp.StatusCode)
	assert.NotEmpty(t, rsp.ID)
	assert.Equal(t, "https://pixeldrain.com/u/mock-file-id", rsp.GetFileURL())
	fmt.Println("POST Req: " + rsp.GetFileURL())
}

// TestPD_UploadPOST_Integration is an integration test for the POST upload method
func TestPD_UploadPOST_Integration(t *testing.T) {
	SetupTestEnvironment()

	server := pd.MockFileUploadServer()
	defer server.Close()
	testURL := server.URL + "/file"

	// Define the hash file path
	hashFilePath := "test_hashes.csv"

	// Initialize hash file
	if err := utils.InitializeHashFile(hashFilePath); err != nil {
		t.Fatalf("Failed to initialize hash file: %v", err)
	}

	req := &pd.RequestUpload{
		PathToFile: "testdata/cat.jpg",
		FileName:   "test_post_cat.jpg",
		Anonymous:  true,
		URL:        testURL,
	}

	c := pd.New(nil, nil)
	rsp, err := c.UploadPOST(req, testHashFilePath)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, 201, rsp.StatusCode)
	assert.NotEmpty(t, rsp.ID)
	assert.Equal(t, "https://pixeldrain.com/u/mock-file-id", rsp.GetFileURL())
	fmt.Println("POST Req: " + rsp.GetFileURL())
}

// TestPD_UploadPOST_WithReadCloser_Integration run a real integration test against the service
func TestPD_UploadPOST_WithReadCloser_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip(SkipIntegrationTest)
	}

	file, _ := os.Open("testdata/cat_unique1.jpg") // Ensure unique file

	req := &pd.RequestUpload{
		File:     file,
		FileName: "test_post_cat_unique1.jpg",
	}

	req.Auth = setAuthFromEnv()

	c := pd.New(nil, nil)
	hashFilePath := "test_hashes.csv"
	rsp, err := c.UploadPOST(req, hashFilePath)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, 201, rsp.StatusCode)
	assert.NotEmpty(t, rsp.ID)
	fmt.Println("POST Req: " + rsp.GetFileURL())
}

// TestPD_UploadPOST_DuplicateDetection_Integration runs an integration test for duplicate file detection
func TestPD_UploadPOST_DuplicateDetection_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip(SkipIntegrationTest)
	}

	filePath := "testdata/cat_unique7.jpg"
	fileName := "test_post_cat_unique7.jpg"

	// First upload
	req := &pd.RequestUpload{
		PathToFile: filePath,
		FileName:   fileName,
	}
	req.Auth = setAuthFromEnv()

	c := pd.New(nil, nil)
	hashFilePath := "test_hashes.csv"
	rsp, err := c.UploadPOST(req, hashFilePath)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 201, rsp.StatusCode)
	assert.NotEmpty(t, rsp.ID)

	// Save the file ID for later use
	fileIDPost = rsp.ID

	// Second upload (should be detected as duplicate)
	rsp, err = c.UploadPOST(req, testHashFilePath)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 409, rsp.StatusCode)
	assert.Equal(t, "Duplicate file. Upload skipped.", rsp.Message)
	assert.Empty(t, rsp.ID)
}

// TestPD_UploadPUT is a unit test for the PUT upload method
func TestPD_UploadPUT(t *testing.T) {
	server := pd.MockFileUploadServer()
	defer server.Close()
	testURL := server.URL + "/file/"

	req := &pd.RequestUpload{
		PathToFile: "testdata/cat.jpg",
		FileName:   "test_put_cat.jpg",
		Anonymous:  true,
		URL:        testURL + "test_put_cat.jpg",
	}

	c := pd.New(nil, nil)
	rsp, err := c.UploadPUT(req)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, 201, rsp.StatusCode)
	assert.NotEmpty(t, rsp.ID)
	assert.Equal(t, "https://pixeldrain.com/u/123456", rsp.GetFileURL())
	fmt.Println("PUT Req: " + rsp.GetFileURL())
}

// TestPD_UploadPUT_Integration run a real integration test against the service
func TestPD_UploadPUT_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip(SkipIntegrationTest)
	}

	req := &pd.RequestUpload{
		PathToFile: "testdata/cat_unique2.jpg",
		FileName:   "test_put_cat_unique2.jpg",
	}

	req.Auth = setAuthFromEnv()

	c := pd.New(nil, nil)
	rsp, err := c.UploadPUT(req)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, 201, rsp.StatusCode)
	assert.NotEmpty(t, rsp.ID)
	fileIDPut = rsp.ID
	fmt.Println("PUT Req: " + rsp.GetFileURL())
}

// TestPD_UploadPUT_WithReadCloser_Integration run a real integration test against the service
func TestPD_UploadPUT_WithReadCloser_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip(SkipIntegrationTest)
	}

	// ReadCloser
	file, _ := os.Open("testdata/cat_unique3.jpg")

	req := &pd.RequestUpload{
		File:     file,
		FileName: "test_put_cat_unique3.jpg",
	}

	req.Auth = setAuthFromEnv()

	c := pd.New(nil, nil)
	rsp, err := c.UploadPUT(req)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, 201, rsp.StatusCode)
	assert.NotEmpty(t, rsp.ID)
	fmt.Println("PUT Req: " + rsp.GetFileURL())
}

// TestPD_Download is a unit test for the GET "download" method
func TestPD_Download(t *testing.T) {
	server := pd.MockFileUploadServer()
	defer server.Close()
	testURL := server.URL + "/file/K1dA8U5W"

	req := &pd.RequestDownload{
		PathToSave: "testdata/cat_download.jpg",
		ID:         "K1dA8U5W",
		URL:        testURL,
	}

	c := pd.New(nil, nil)
	rsp, err := c.Download(req)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, 200, rsp.StatusCode)
	assert.Equal(t, true, rsp.Success)
}

// TestPD_Download_Integration run a real integration test against the service
func TestPD_Download_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip(SkipIntegrationTest)
	}

	reqUpload := &pd.RequestUpload{
		PathToFile: "testdata/cat_unique4.jpg", // Ensure unique file
		FileName:   "test_post_cat_unique4.jpg",
	}

	reqUpload.Auth = setAuthFromEnv()

	c := pd.New(nil, nil)
	rspUpload, err := c.UploadPOST(reqUpload, testHashFilePath)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 201, rspUpload.StatusCode)
	fileIDPost = rspUpload.ID

	reqDownload := &pd.RequestDownload{
		PathToSave: "testdata/cat_download.jpg",
		ID:         fileIDPost,
	}

	reqDownload.Auth = setAuthFromEnv()

	rspDownload, err := c.Download(reqDownload)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 200, rspDownload.StatusCode)
	assert.Equal(t, "cat_download.jpg", rspDownload.FileName)
	assert.Equal(t, int64(33692), rspDownload.FileSize)
}

// TestPD_GetFileInfo is a unit test for the GET "file info" method
func TestPD_GetFileInfo(t *testing.T) {
	server := pd.MockFileUploadServer()
	defer server.Close()
	testURL := server.URL + "/file/K1dA8U5W/info"

	req := &pd.RequestFileInfo{
		ID:  "K1dA8U5W",
		URL: testURL,
	}

	c := pd.New(nil, nil)
	rsp, err := c.GetFileInfo(req)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, 200, rsp.StatusCode)
	assert.Equal(t, true, rsp.Success)
	assert.Equal(t, "K1dA8U5W", rsp.ID)
	assert.Equal(t, int64(37621), rsp.Size)
	assert.Equal(t, "1af93d68009bdfd52e1da100a019a30b5fe083d2d1130919225ad0fd3d1fed0b", rsp.HashSha256)
}

// TestPD_GetFileInfo_Integration run a real integration test against the service
func TestPD_GetFileInfo_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip(SkipIntegrationTest)
	}

	reqUpload := &pd.RequestUpload{
		PathToFile: "testdata/cat_unique5.jpg", // Ensure unique file
		FileName:   "test_post_cat_unique5.jpg",
	}

	reqUpload.Auth = setAuthFromEnv()

	c := pd.New(nil, nil)
	rspUpload, err := c.UploadPOST(reqUpload, "test_hashes.csv")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 201, rspUpload.StatusCode)
	fileIDPost = rspUpload.ID

	reqFileInfo := &pd.RequestFileInfo{
		ID: fileIDPost,
	}

	reqFileInfo.Auth = setAuthFromEnv()

	rspFileInfo, err := c.GetFileInfo(reqFileInfo)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 200, rspFileInfo.StatusCode)
	assert.Equal(t, true, rspFileInfo.Success)
	assert.Equal(t, fileIDPost, rspFileInfo.ID)
	assert.Equal(t, int64(50936), rspFileInfo.Size)
	assert.Equal(t, "c3d90e9743e45e488996a252426a71416f001646fd9f6ce1d4b7a0b00369ee3e", rspFileInfo.HashSha256)
}

// TestPD_DownloadThumbnail is a unit test for the GET "download thumbnail" method
func TestPD_DownloadThumbnail(t *testing.T) {
	server := pd.MockFileUploadServer()
	defer server.Close()
	testURL := server.URL + "/file/K1dA8U5W/thumbnail?width=64&height=64"

	req := &pd.RequestThumbnail{
		ID:         "K1dA8U5W",
		Height:     "64",
		Width:      "64",
		PathToSave: "testdata/cat_download_thumbnail.jpg",
		URL:        testURL,
	}

	req.Auth = setAuthFromEnv()

	c := pd.New(nil, nil)
	rsp, err := c.DownloadThumbnail(req)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, 200, rsp.StatusCode)
	assert.Equal(t, "cat_download_thumbnail.jpg", rsp.FileName)
	assert.Equal(t, int64(51680), rsp.FileSize)
}

// TestPD_DownloadThumbnail_Integration run a real integration test against the service
func TestPD_DownloadThumbnail_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip(SkipIntegrationTest)
	}

	reqUpload := &pd.RequestUpload{
		PathToFile: "testdata/cat_unique6.jpg", // Ensure unique file
		FileName:   "test_post_cat_unique6.jpg",
	}

	reqUpload.Auth = setAuthFromEnv()

	c := pd.New(nil, nil)
	rspUpload, err := c.UploadPOST(reqUpload, "test_hashes.csv")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 201, rspUpload.StatusCode)
	fileIDPost = rspUpload.ID

	reqThumbnail := &pd.RequestThumbnail{
		ID:         fileIDPost,
		Height:     "64",
		Width:      "64",
		PathToSave: "testdata/cat_download_thumbnail.jpg",
	}

	reqThumbnail.Auth = setAuthFromEnv()

	rspThumbnail, err := c.DownloadThumbnail(reqThumbnail)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 200, rspThumbnail.StatusCode)
	assert.Equal(t, "cat_download_thumbnail.jpg", rspThumbnail.FileName)
	assert.Equal(t, int64(8849), rspThumbnail.FileSize)
}

// TestPD_CreateList is a unit test for the POST "list" method
func TestPD_CreateList(t *testing.T) {
	server := pd.MockFileUploadServer()
	defer server.Close()
	testURL := server.URL + "/list"

	// files to add
	files := []pd.ListFile{
		{ID: "K1dA8U5W", Description: "Hallo Welt"},
		{ID: "bmrc4iyD", Description: "Hallo Welt 2"},
	}

	// create list request
	req := &pd.RequestCreateList{
		Title:     "Test List",
		Anonymous: false,
		Files:     files,
		URL:       testURL,
	}

	req.Auth = setAuthFromEnv()

	c := pd.New(nil, nil)
	rsp, err := c.CreateList(req)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, 200, rsp.StatusCode)
	assert.Equal(t, true, rsp.Success)
	assert.NotEmpty(t, rsp.ID)
}

// TestPD_Delete_Integration run a real integration test against the service
func TestPD_CreateList_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip(SkipIntegrationTest)
	}

	// files to add
	files := []pd.ListFile{
		{ID: fileIDPost, Description: "Hallo Welt"},
		{ID: fileIDPut, Description: "Hallo Welt 2"},
	}

	// create list request
	req := &pd.RequestCreateList{
		Title:     "Test List",
		Anonymous: false,
		Files:     files,
	}

	req.Auth = setAuthFromEnv()

	c := pd.New(nil, nil)
	rsp, err := c.CreateList(req)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, 201, rsp.StatusCode)
	assert.Equal(t, true, rsp.Success)
	listID = rsp.ID
}

// TestPD_GetList is a unit test for the GET "list/{id}" method
func TestPD_GetList(t *testing.T) {
	server := pd.MockFileUploadServer()
	defer server.Close()
	testURL := server.URL + "/list/123"

	req := &pd.RequestGetList{
		ID:  "123",
		URL: testURL,
	}

	req.Auth = setAuthFromEnv()

	c := pd.New(nil, nil)
	rsp, err := c.GetList(req)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, 200, rsp.StatusCode)
	assert.Equal(t, true, rsp.Success)
	assert.NotEmpty(t, rsp.ID)
	assert.Equal(t, "Rust in Peace", rsp.Title)
	assert.Equal(t, int64(123456), rsp.Files[0].Size)
}

// TestPD_GetList_Integration run a real integration test against the service
func TestPD_GetList_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip(SkipIntegrationTest)
	}

	req := &pd.RequestGetList{
		ID: listID,
	}

	req.Auth = setAuthFromEnv()

	c := pd.New(nil, nil)
	rsp, err := c.GetList(req)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, 200, rsp.StatusCode)
	assert.Equal(t, true, rsp.Success)
	assert.NotEmpty(t, rsp.ID)
	assert.Equal(t, "Test List", rsp.Title)
	assert.Equal(t, int64(69142), rsp.Files[0].Size)
}

// TestPD_GetUser is a unit test for the GET "/user" method
func TestPD_GetUser(t *testing.T) {
	server := pd.MockFileUploadServer()
	defer server.Close()
	testURL := server.URL + "/user"

	req := &pd.RequestGetUser{
		URL: testURL,
	}

	req.Auth = setAuthFromEnv()

	c := pd.New(nil, nil)
	rsp, err := c.GetUser(req)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, 200, rsp.StatusCode)
	assert.Equal(t, true, rsp.Success)
	assert.Equal(t, "TestTest", rsp.Username)
	assert.Equal(t, "Free", rsp.Subscription.Name)
}

// TestPD_GetUser_Integration run a real integration test against the service
func TestPD_GetUser_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip(SkipIntegrationTest)
	}

	req := &pd.RequestGetUser{}

	req.Auth = setAuthFromEnv()

	c := pd.New(nil, nil)
	rsp, err := c.GetUser(req)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, 200, rsp.StatusCode)
	assert.Equal(t, true, rsp.Success)
	assert.Equal(t, "sordidgoose", rsp.Username)
	assert.Equal(t, "Pro", rsp.Subscription.Name)
}

// TestPD_GetUserFiles is a unit test for the GET "/user/files" method
func TestPD_GetUserFiles(t *testing.T) {
	server := pd.MockFileUploadServer()
	defer server.Close()
	testURL := server.URL + "/user/files"

	req := &pd.RequestGetUserFiles{
		URL: testURL,
	}

	req.Auth = setAuthFromEnv()

	c := pd.New(nil, nil)
	rsp, err := c.GetUserFiles(req)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, 200, rsp.StatusCode)
	assert.Equal(t, true, rsp.Success)
	assert.Equal(t, "tUxgDCoQ", rsp.Files[0].ID)
	assert.Equal(t, "test_post_cat.jpg", rsp.Files[0].Name)
}

// TestPD_GetUserFiles_Integration run a real integration test against the service
func TestPD_GetUserFiles_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip(SkipIntegrationTest)
	}

	req := &pd.RequestGetUserFiles{}

	req.Auth = setAuthFromEnv()

	c := pd.New(nil, nil)
	rsp, err := c.GetUserFiles(req)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, 200, rsp.StatusCode)
	assert.Equal(t, true, rsp.Success)

	if len(rsp.Files) >= 2 {
		assert.True(t, true)
	}

	for _, file := range rsp.Files {
		if file.ID == fileIDPost {
			assert.Equal(t, fileIDPost, file.ID)
		}

		if file.ID == fileIDPut {
			assert.Equal(t, fileIDPut, file.ID)
		}
	}
}

// TestPD_GetUserLists is a unit test for the GET "/user/files" method
func TestPD_GetUserLists(t *testing.T) {
	server := pd.MockFileUploadServer()
	defer server.Close()
	testURL := server.URL + "/user/lists"

	req := &pd.RequestGetUserLists{
		URL: testURL,
	}

	req.Auth = setAuthFromEnv()

	c := pd.New(nil, nil)
	rsp, err := c.GetUserLists(req)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, 200, rsp.StatusCode)
	assert.Equal(t, true, rsp.Success)
	assert.Equal(t, "Test List", rsp.Lists[0].Title)
}

// TestPD_GetUserLists_Integration run a real integration test against the service
func TestPD_GetUserLists_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip(SkipIntegrationTest)
	}

	req := &pd.RequestGetUserLists{}

	req.Auth = setAuthFromEnv()

	c := pd.New(nil, nil)
	rsp, err := c.GetUserLists(req)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, 200, rsp.StatusCode)
	assert.Equal(t, true, rsp.Success)
	assert.Equal(t, "Test List", rsp.Lists[0].Title)
}

// TestPD_Delete is a unit test for the DELETE "delete" method
func TestPD_Delete(t *testing.T) {
	server := pd.MockFileUploadServer()
	defer server.Close()
	testURL := server.URL + "/file/K1dA8U5W"

	req := &pd.RequestDelete{
		ID:  "K1dA8U5W",
		URL: testURL,
	}

	req.Auth = setAuthFromEnv()

	c := pd.New(nil, nil)
	rsp, err := c.Delete(req)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, true, rsp.Success)
	assert.Equal(t, "file_deleted", rsp.Value)
	assert.Equal(t, "The file has been deleted.", rsp.Message)
}

// TestPD_Delete_Integration run a real integration test against the service
func TestPD_Delete_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip(SkipIntegrationTest)
	}

	req := &pd.RequestDelete{
		ID: fileIDPost,
	}

	req.Auth = setAuthFromEnv()

	c := pd.New(nil, nil)
	rsp, err := c.Delete(req)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, true, rsp.Success)
	assert.Equal(t, "ok", rsp.Value)
	assert.Equal(t, "The requested action was successfully performed", rsp.Message)

	req = &pd.RequestDelete{
		ID: fileIDPut,
	}

	req.Auth = setAuthFromEnv()

	c = pd.New(nil, nil)
	rsp, err = c.Delete(req)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, true, rsp.Success)
	assert.Equal(t, "ok", rsp.Value)
	assert.Equal(t, "The requested action was successfully performed", rsp.Message)
}

func setAuthFromEnv() pd.Auth {
	// load api key from .env_test file
	currentWorkDirectory, _ := os.Getwd()
	_ = godotenv.Load(currentWorkDirectory + "/.env_test")
	apiKey := os.Getenv("API_KEY")

	return pd.Auth{
		APIKey: apiKey,
	}
}

func TestSaveUploadInfoToCSV(t *testing.T) {
	csvPath := "test_upload_logs.csv"
	defer os.Remove(csvPath) // Cleanup test file after the test

	// Get the actual file size
	testFilePath := "testdata/test_file.jpg"
	fileInfo, err := os.Stat(testFilePath)
	if err != nil {
		t.Fatalf("failed to get file stats: %v", err)
	}
	actualFileSize := fileInfo.Size()

	info := utils.UploadInfo{
		FileName:       "test_file.jpg",
		DirectoryPath:  "/test/path",
		URL:            "https://pixeldrain.com/u/test",
		UploadDateTime: time.Now().Format(time.RFC3339),
		FileSize:       actualFileSize,
		MIMEType:       "image/jpeg",
		Uploader:       "test_user",
		UploadStatus:   "200",
	}

	err = utils.SaveUploadInfoToCSV(info, csvPath)
	if err != nil {
		t.Fatalf("failed to save upload info to CSV: %v", err)
	}

	file, err := os.Open(csvPath)
	if err != nil {
		t.Fatalf("failed to open CSV file: %v", err)
	}
	defer file.Close()

	stats, err := file.Stat()
	if err != nil {
		t.Fatalf("failed to get file stats: %v", err)
	}

	if stats.Size() == 0 {
		t.Fatalf("CSV file is empty, expected data to be written")
	}

	expectedMimeType := "image/jpeg"

	mimeType := utils.GetMimeType(testFilePath)
	if mimeType != expectedMimeType {
		t.Fatalf("expected MIME type %s, got %s", expectedMimeType, mimeType)
	}

	fileSize := utils.GetFileSize(testFilePath)
	if fileSize != actualFileSize {
		t.Fatalf("expected file size %d, got %d", actualFileSize, fileSize)
	}
}

func TestUploadDirectory(t *testing.T) {
	SetupTestEnvironment()
	// Create a mock server
	server := pd.MockFileUploadServer()
	defer server.Close()

	clientOptions := &pd.ClientOptions{
		Debug: true,
	}

	client := pd.New(clientOptions, nil)

	// Mock Auth
	auth := pd.Auth{
		APIKey: "test-api-key",
	}

	// Use the mock server URL as the base URL
	err := client.UploadDirectory("testdata/test_directory", auth, server.URL)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Additional checks can be added to validate the upload and logging
}

func TestUploadDirectory_Integration(t *testing.T) {
	SetupTestEnvironment()
	if testing.Short() {
		t.Skip(SkipIntegrationTest)
	}

	clientOptions := &pd.ClientOptions{
		Debug: true,
	}

	client := pd.New(clientOptions, nil)

	auth := setAuthFromEnv()

	// Use the actual API URL
	apiURL := "https://pixeldrain.com/api"

	err := client.UploadDirectory("testdata/test_directory", auth, apiURL)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Additional checks can be added to validate the upload and logging
}

func TestCalculateFileHash(t *testing.T) {
	filePath := "testdata/cat.jpg"

	expectedHash := "1af93d68009bdfd52e1da100a019a30b5fe083d2d1130919225ad0fd3d1fed0b"
	hash, err := utils.CalculateFileHash(filePath)
	if err != nil {
		t.Fatalf("Failed to calculate file hash: %v", err)
	}

	assert.Equal(t, expectedHash, hash)
}

func TestSaveAndLoadFileHashes(t *testing.T) {
	testHashFilePath := "test_hashes.csv"
	defer os.Remove(testHashFilePath) // Cleanup test file after the test

	// Initialize hash file
	if err := utils.InitializeHashFile(testHashFilePath); err != nil {
		t.Fatalf("Failed to initialize hash file: %v", err)
	}

	filePath := "testdata/cat.jpg"
	fileHash := "1af93d68009bdfd52e1da100a019a30b5fe083d2d1130919225ad0fd3d1fed0b"

	err := utils.SaveFileHash(testHashFilePath, filePath, fileHash)
	if err != nil {
		t.Fatalf("Failed to save file hash: %v", err)
	}

	hashes, err := utils.LoadFileHashes(testHashFilePath)
	if err != nil {
		t.Fatalf("Failed to load file hashes: %v", err)
	}

	assert.Equal(t, fileHash, hashes[filePath])
}

func TestIsDuplicate(t *testing.T) {
	testHashFilePath := "test_hashes.csv"

	// Ensure test_hashes.csv is created
	if err := utils.InitializeHashFile(testHashFilePath); err != nil {
		t.Fatalf("Failed to initialize hash file: %v", err)
	}

	// Cleanup test file after the test
	defer func() {
		if err := os.Remove(testHashFilePath); err != nil && !os.IsNotExist(err) {
			t.Fatalf("Failed to remove test hash file: %v", err)
		}
	}()

	filePath := "testdata/cat.jpg"
	fileHash := "1af93d68009bdfd52e1da100a019a30b5fe083d2d1130919225ad0fd3d1fed0b"

	// Save the file hash to simulate a previous upload
	err := utils.SaveFileHash(testHashFilePath, filePath, fileHash)
	if err != nil {
		t.Fatalf("Failed to save file hash: %v", err)
	}

	// Check for duplicate
	isDuplicate, err := utils.IsDuplicate(testHashFilePath, filePath)
	if err != nil {
		t.Fatalf("Failed to check duplicate: %v", err)
	}

	assert.True(t, isDuplicate)
}
